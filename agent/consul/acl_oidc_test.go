// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package consul

import (
	"net/url"
	"testing"
	"time"

	msgpackrpc "github.com/hashicorp/consul-net-rpc/net-rpc-msgpackrpc"
	"github.com/stretchr/testify/require"

	"github.com/opensoba/opensoba/agent/structs"
	"github.com/opensoba/opensoba/internal/go-sso/oidcauth/oidcauthtest"
)

// TestACLEndpoint_OIDCStateReaper verifies that the leader garbage collects
// expired OIDC authorization code flow state while leaving live state intact.
func TestACLEndpoint_OIDCStateReaper(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}
	t.Parallel()

	_, srv, _ := testACLServerWithConfig(t, nil, false)
	waitForLeaderEstablishment(t, srv)

	now := time.Now()
	require.NoError(t, func() error {
		req := structs.ACLOIDCStateSetRequest{
			States: []*structs.ACLOIDCState{
				{StateID: "expired", Nonce: "n1", RedirectURI: "https://example.com", ExpiresAt: now.Add(-time.Minute)},
				{StateID: "live", Nonce: "n2", RedirectURI: "https://example.com", ExpiresAt: now.Add(time.Hour)},
			},
		}
		_, err := srv.raftApply(structs.ACLOIDCStateSetRequestType, &req)
		return err
	}())

	srv.reapOIDCStates()

	expired, err := srv.fsm.State().OIDCStateGet("expired")
	require.NoError(t, err)
	require.Nil(t, expired)

	live, err := srv.fsm.State().OIDCStateGet("live")
	require.NoError(t, err)
	require.NotNil(t, live)
	require.Equal(t, "n2", live.Nonce)
}

func oidcURLParam(t *testing.T, inputURL, param string) string {
	t.Helper()

	u, err := url.Parse(inputURL)
	require.NoError(t, err)
	v := u.Query().Get(param)
	require.NotEmpty(t, v, "query param %q not found in %q", param, inputURL)
	return v
}

// TestACLEndpoint_OIDC_LoadBalanced verifies that an OIDC login flow can begin
// on one server and complete on a different server. This is the core
// requirement for running OpenSoba behind a load balancer: flow state must be
// shared through Raft rather than held in process-local memory.
func TestACLEndpoint_OIDC_LoadBalanced(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}
	t.Parallel()

	const (
		clientID     = "soba-lb-client"
		clientSecret = "lb-secret"
		redirectURI  = "http://127.0.0.1:8250/oidc/callback"
	)

	idp := oidcauthtest.Start(t)
	idp.SetClientCreds(clientID, clientSecret)
	idp.SetAllowedRedirectURIs([]string{redirectURI})

	cluster := newTestCluster(t, &testClusterConfig{
		Datacenter: "dc1",
		Servers:    3,
		ServerConf: testServerACLConfig,
	})

	// Pick the leader (for setup) and two distinct followers so that the
	// auth-url and callback requests land on different servers.
	var leaderIdx = -1
	var followerIdxs []int
	for i, srv := range cluster.Servers {
		if srv.IsLeader() {
			leaderIdx = i
		} else {
			followerIdxs = append(followerIdxs, i)
		}
	}
	require.NotEqual(t, -1, leaderIdx, "no leader elected")
	require.GreaterOrEqual(t, len(followerIdxs), 2, "expected at least two followers")

	leaderCodec := cluster.ServerCodecs[leaderIdx]
	authURLCodec := cluster.ServerCodecs[followerIdxs[0]]
	callbackCodec := cluster.ServerCodecs[followerIdxs[1]]

	// Create an OIDC auth method that points at the fake identity provider.
	var method structs.ACLAuthMethod
	require.NoError(t, msgpackrpc.CallWithCodec(leaderCodec, "ACL.AuthMethodSet", &structs.ACLAuthMethodSetRequest{
		Datacenter: "dc1",
		AuthMethod: structs.ACLAuthMethod{
			Name:        "lb-oidc",
			Type:        "oidc",
			DisplayName: "LB SSO",
			Config: map[string]interface{}{
				"OIDCDiscoveryURL":    idp.Addr(),
				"OIDCDiscoveryCACert": idp.CACert(),
				"OIDCClientID":        clientID,
				"OIDCClientSecret":    clientSecret,
				"AllowedRedirectURIs": []string{redirectURI},
				"JWTSupportedAlgs":    []string{"ES256"},
				"BoundAudiences":      []string{clientID},
				"ClaimMappings":       map[string]string{"first_name": "name"},
			},
		},
		WriteRequest: structs.WriteRequest{Token: TestDefaultInitialManagementToken},
	}, &method))

	// A binding rule so a matching identity can mint a token.
	var rule structs.ACLBindingRule
	require.NoError(t, msgpackrpc.CallWithCodec(leaderCodec, "ACL.BindingRuleSet", &structs.ACLBindingRuleSetRequest{
		Datacenter: "dc1",
		BindingRule: structs.ACLBindingRule{
			AuthMethod: "lb-oidc",
			Selector:   "value.name == jeff",
			BindType:   structs.BindingRuleBindTypeService,
			BindName:   "lb-web",
		},
		WriteRequest: structs.WriteRequest{Token: TestDefaultInitialManagementToken},
	}, &rule))

	// Start the flow on the first follower.
	var authURLResp structs.ACLOIDCAuthURLResponse
	require.NoError(t, msgpackrpc.CallWithCodec(authURLCodec, "ACL.OIDCAuthURL", &structs.ACLOIDCAuthURLRequest{
		Datacenter:   "dc1",
		AuthMethod:   "lb-oidc",
		RedirectURI:  redirectURI,
		ClientNonce:  "lb-client-nonce",
		WriteRequest: structs.WriteRequest{},
	}, &authURLResp))
	require.NotEmpty(t, authURLResp.AuthURL)

	state := oidcURLParam(t, authURLResp.AuthURL, "state")
	nonce := oidcURLParam(t, authURLResp.AuthURL, "nonce")

	idp.SetCustomClaims(map[string]interface{}{
		"nonce":      nonce,
		"first_name": "jeff",
	})
	idp.SetExpectedAuthCode("lb-auth-code")

	// Complete the flow on a different follower.
	var token structs.ACLToken
	require.NoError(t, msgpackrpc.CallWithCodec(callbackCodec, "ACL.OIDCCallback", &structs.ACLOIDCCallbackRequest{
		Datacenter:   "dc1",
		AuthMethod:   "lb-oidc",
		State:        state,
		Code:         "lb-auth-code",
		ClientNonce:  "lb-client-nonce",
		WriteRequest: structs.WriteRequest{},
	}, &token))

	require.NotEmpty(t, token.AccessorID)
	require.NotEmpty(t, token.SecretID)
	require.Equal(t, "lb-oidc", token.AuthMethod)
	require.Len(t, token.ServiceIdentities, 1)
	require.Equal(t, "lb-web", token.ServiceIdentities[0].ServiceName)

	// The state must not be usable again, even from another server.
	err := msgpackrpc.CallWithCodec(authURLCodec, "ACL.OIDCCallback", &structs.ACLOIDCCallbackRequest{
		Datacenter:   "dc1",
		AuthMethod:   "lb-oidc",
		State:        state,
		Code:         "lb-auth-code",
		ClientNonce:  "lb-client-nonce",
		WriteRequest: structs.WriteRequest{},
	}, &token)
	require.Error(t, err)
}
