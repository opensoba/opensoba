// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opensoba/opensoba/agent/structs"
	"github.com/opensoba/opensoba/internal/go-sso/oidcauth/oidcauthtest"
	testrpc "github.com/opensoba/opensoba/testrpc"
)

// oidcTestQueryParam extracts a query parameter from a full URL string.
func oidcTestQueryParam(t *testing.T, inputURL, param string) string {
	t.Helper()

	idx := -1
	for i := 0; i < len(inputURL); i++ {
		if inputURL[i] == '?' {
			idx = i
			break
		}
	}
	require.NotEqual(t, -1, idx, "no query string in %q", inputURL)

	m, err := url.ParseQuery(inputURL[idx+1:])
	require.NoError(t, err)
	v, ok := m[param]
	require.True(t, ok, "query param %q not found in %q", param, inputURL)
	require.NotEmpty(t, v)
	return v[0]
}

func TestACL_OIDC_HTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	const (
		clientID     = "soba-test-client"
		clientSecret = "super-secret-value"
		redirectURI  = "http://127.0.0.1:8250/oidc/callback"
	)

	// The fake identity provider acts as our Keycloak stand-in.
	oidcServer := oidcauthtest.Start(t)
	oidcServer.SetClientCreds(clientID, clientSecret)
	oidcServer.SetAllowedRedirectURIs([]string{redirectURI})

	a := NewTestAgent(t, TestACLConfig())
	defer a.Shutdown()

	testrpc.WaitForLeader(t, a.RPC, "dc1")

	// helpers -----------------------------------------------------------------

	do := func(method, path string, body any, handler func(http.ResponseWriter, *http.Request) (interface{}, error)) interface{} {
		t.Helper()
		var req *http.Request
		if body != nil {
			req, _ = http.NewRequest(method, path, jsonBody(body))
		} else {
			req, _ = http.NewRequest(method, path, nil)
		}
		req.Header.Add("X-Consul-Token", "root")
		resp := httptest.NewRecorder()
		obj, err := handler(resp, req)
		require.NoError(t, err)
		return obj
	}

	// A policy and role for the binding rule to hand out.
	do("PUT", "/v1/acl/policy", &structs.ACLPolicy{
		Name:  "oidc-policy",
		Rules: `acl = "read"`,
	}, a.srv.ACLPolicyCreate)

	do("PUT", "/v1/acl/role", &structs.ACLRole{
		Name:     "oidc-role",
		Policies: []structs.ACLRolePolicyLink{{Name: "oidc-policy"}},
	}, a.srv.ACLRoleCreate)

	methodConfig := map[string]interface{}{
		"OIDCDiscoveryURL":    oidcServer.Addr(),
		"OIDCDiscoveryCACert": oidcServer.CACert(),
		"OIDCClientID":        clientID,
		"OIDCClientSecret":    clientSecret,
		"AllowedRedirectURIs": []string{redirectURI},
		"JWTSupportedAlgs":    []string{"ES256"},
		"BoundAudiences":      []string{clientID},
		"ClaimMappings":       map[string]string{"first_name": "name"},
	}

	// Create the OIDC auth method.
	created := do("PUT", "/v1/acl/auth-method", &structs.ACLAuthMethod{
		Name:        "test-oidc",
		Type:        "oidc",
		DisplayName: "OpenSoba SSO",
		Description: "OIDC test method",
		Config:      methodConfig,
	}, a.srv.ACLAuthMethodCreate)
	method, ok := created.(*structs.ACLAuthMethod)
	require.True(t, ok)
	require.Equal(t, "test-oidc", method.Name)
	// The secret must never be echoed back to the API client.
	require.NotContains(t, method.Config, "OIDCClientSecret")

	// A binding rule that only matches the "jeff" identity.
	do("PUT", "/v1/acl/binding-rule", &structs.ACLBindingRule{
		AuthMethod: "test-oidc",
		Selector:   "value.name == jeff",
		BindType:   "role",
		BindName:   "oidc-role",
	}, a.srv.ACLBindingRuleCreate)

	// read + masking ----------------------------------------------------------

	t.Run("read masks the client secret", func(t *testing.T) {
		read := do("GET", "/v1/acl/auth-method/test-oidc", nil, func(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
			return a.srv.ACLAuthMethodCRUD(resp, req)
		})
		got, ok := read.(*structs.ACLAuthMethod)
		require.True(t, ok)
		require.NotContains(t, got.Config, "OIDCClientSecret")
		// Non-sensitive fields are still present.
		require.Equal(t, clientID, got.Config["OIDCClientID"])
	})

	// update preserves the secret --------------------------------------------

	t.Run("update preserves the stored secret", func(t *testing.T) {
		// Simulate a read/modify/write cycle from the CLI/UI: the config
		// returned by the read has no OIDCClientSecret, so neither does the
		// update.
		updatedConfig := map[string]interface{}{
			"OIDCDiscoveryURL":    oidcServer.Addr(),
			"OIDCDiscoveryCACert": oidcServer.CACert(),
			"OIDCClientID":        clientID,
			"AllowedRedirectURIs": []string{redirectURI},
			"JWTSupportedAlgs":    []string{"ES256"},
			"BoundAudiences":      []string{clientID},
			"ClaimMappings":       map[string]string{"first_name": "name"},
		}
		do("PUT", "/v1/acl/auth-method/test-oidc", &structs.ACLAuthMethod{
			Name:        "test-oidc",
			Type:        "oidc",
			DisplayName: "OpenSoba SSO",
			Description: "updated",
			Config:      updatedConfig,
		}, func(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
			return a.srv.ACLAuthMethodCRUD(resp, req)
		})

		// The auth method must still be usable, which is only possible if the
		// secret was carried forward server-side. The flow below exercises that.
	})

	// UI provider list (pre-authentication) -----------------------------------

	t.Run("ui provider list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/internal/ui/oidc-auth-methods", nil)
		resp := httptest.NewRecorder()
		raw, err := a.srv.UIOIDCProviders(resp, req)
		require.NoError(t, err)

		providers, ok := raw.([]*structs.ACLOIDCAuthMethod)
		require.True(t, ok)
		require.NotEmpty(t, providers)

		var found *structs.ACLOIDCAuthMethod
		for _, p := range providers {
			if p.Name == "test-oidc" {
				found = p
				break
			}
		}
		require.NotNil(t, found)
		require.Equal(t, "OpenSoba SSO", found.DisplayName)
		require.Equal(t, "no-icon", found.Kind)
	})

	// full authorization code flow --------------------------------------------

	t.Run("auth url and callback", func(t *testing.T) {
		const clientNonce = "cli-nonce-1234567890"

		authURLRaw := do("POST", "/v1/acl/oidc/auth-url", &structs.ACLOIDCAuthURLRequest{
			AuthMethod:  "test-oidc",
			RedirectURI: redirectURI,
			ClientNonce: clientNonce,
			Meta:        map[string]string{"foo": "bar"},
		}, a.srv.ACLOIDCAuthURL)
		authURLResp, ok := authURLRaw.(*structs.ACLOIDCAuthURLResponse)
		require.True(t, ok)
		require.NotEmpty(t, authURLResp.AuthURL)

		state := oidcTestQueryParam(t, authURLResp.AuthURL, "state")
		nonce := oidcTestQueryParam(t, authURLResp.AuthURL, "nonce")

		// Configure the fake provider to hand back an ID token that matches the
		// flow nonce and bakes in the mapped claim.
		oidcServer.SetCustomClaims(map[string]interface{}{
			"nonce":      nonce,
			"first_name": "jeff",
		})
		oidcServer.SetExpectedAuthCode("auth-code-abc")

		callbackRaw := do("POST", "/v1/acl/oidc/callback", &structs.ACLOIDCCallbackRequest{
			AuthMethod:  "test-oidc",
			State:       state,
			Code:        "auth-code-abc",
			ClientNonce: clientNonce,
		}, a.srv.ACLOIDCCallback)
		token, ok := callbackRaw.(*structs.ACLToken)
		require.True(t, ok)
		require.Len(t, token.AccessorID, 36)
		require.Len(t, token.SecretID, 36)
		require.Equal(t, "test-oidc", token.AuthMethod)
		require.Equal(t, `token created via login: {"foo":"bar"}`, token.Description)
		require.True(t, token.Local)
		require.Len(t, token.Roles, 1)
		require.Equal(t, "oidc-role", token.Roles[0].Name)
	})

	// state is single-use ------------------------------------------------------

	t.Run("callback state is single use", func(t *testing.T) {
		authURLRaw := do("POST", "/v1/acl/oidc/auth-url", &structs.ACLOIDCAuthURLRequest{
			AuthMethod:  "test-oidc",
			RedirectURI: redirectURI,
		}, a.srv.ACLOIDCAuthURL)
		authURLResp := authURLRaw.(*structs.ACLOIDCAuthURLResponse)

		state := oidcTestQueryParam(t, authURLResp.AuthURL, "state")
		nonce := oidcTestQueryParam(t, authURLResp.AuthURL, "nonce")

		oidcServer.SetCustomClaims(map[string]interface{}{
			"nonce":      nonce,
			"first_name": "jeff",
		})
		oidcServer.SetExpectedAuthCode("auth-code-def")

		// First use works.
		do("POST", "/v1/acl/oidc/callback", &structs.ACLOIDCCallbackRequest{
			AuthMethod: "test-oidc",
			State:      state,
			Code:       "auth-code-def",
		}, a.srv.ACLOIDCCallback)

		// Replaying the same state must fail.
		req, _ := http.NewRequest("POST", "/v1/acl/oidc/callback", jsonBody(&structs.ACLOIDCCallbackRequest{
			AuthMethod: "test-oidc",
			State:      state,
			Code:       "auth-code-def",
		}))
		req.Header.Add("X-Consul-Token", "root")
		resp := httptest.NewRecorder()
		_, err := a.srv.ACLOIDCCallback(resp, req)
		require.Error(t, err)
	})

	// a callback with a mismatched client nonce must be rejected ---------------

	t.Run("callback rejects mismatched client nonce", func(t *testing.T) {
		authURLRaw := do("POST", "/v1/acl/oidc/auth-url", &structs.ACLOIDCAuthURLRequest{
			AuthMethod:  "test-oidc",
			RedirectURI: redirectURI,
			ClientNonce: "the-real-nonce",
		}, a.srv.ACLOIDCAuthURL)
		authURLResp := authURLRaw.(*structs.ACLOIDCAuthURLResponse)

		state := oidcTestQueryParam(t, authURLResp.AuthURL, "state")
		nonce := oidcTestQueryParam(t, authURLResp.AuthURL, "nonce")

		oidcServer.SetCustomClaims(map[string]interface{}{
			"nonce":      nonce,
			"first_name": "jeff",
		})
		oidcServer.SetExpectedAuthCode("auth-code-ghi")

		req, _ := http.NewRequest("POST", "/v1/acl/oidc/callback", jsonBody(&structs.ACLOIDCCallbackRequest{
			AuthMethod:  "test-oidc",
			State:       state,
			Code:        "auth-code-ghi",
			ClientNonce: "a-different-nonce",
		}))
		req.Header.Add("X-Consul-Token", "root")
		resp := httptest.NewRecorder()
		_, err := a.srv.ACLOIDCCallback(resp, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "client nonce")
	})

	// a non-matching binding rule yields no token ------------------------------

	t.Run("callback denied when no binding rule matches", func(t *testing.T) {
		authURLRaw := do("POST", "/v1/acl/oidc/auth-url", &structs.ACLOIDCAuthURLRequest{
			AuthMethod:  "test-oidc",
			RedirectURI: redirectURI,
		}, a.srv.ACLOIDCAuthURL)
		authURLResp := authURLRaw.(*structs.ACLOIDCAuthURLResponse)

		state := oidcTestQueryParam(t, authURLResp.AuthURL, "state")
		nonce := oidcTestQueryParam(t, authURLResp.AuthURL, "nonce")

		oidcServer.SetCustomClaims(map[string]interface{}{
			"nonce":      nonce,
			"first_name": "not-jeff",
		})
		oidcServer.SetExpectedAuthCode("auth-code-jkl")

		req, _ := http.NewRequest("POST", "/v1/acl/oidc/callback", jsonBody(&structs.ACLOIDCCallbackRequest{
			AuthMethod: "test-oidc",
			State:      state,
			Code:       "auth-code-jkl",
		}))
		req.Header.Add("X-Consul-Token", "root")
		resp := httptest.NewRecorder()
		_, err := a.srv.ACLOIDCCallback(resp, req)
		require.Error(t, err)
	})
}
