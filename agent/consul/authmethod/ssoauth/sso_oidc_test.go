// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoauth

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"

	"github.com/opensoba/opensoba/agent/consul/authmethod"
	"github.com/opensoba/opensoba/agent/structs"
	"github.com/opensoba/opensoba/internal/go-sso/oidcauth"
	"github.com/opensoba/opensoba/internal/go-sso/oidcauth/oidcauthtest"
)

// memStateStore is a minimal in-memory oidcauth.StateStore for tests.
type memStateStore struct {
	mu     sync.Mutex
	states map[string]*oidcauth.OIDCState
}

func newMemStateStore() *memStateStore {
	return &memStateStore{states: make(map[string]*oidcauth.OIDCState)}
}

func (s *memStateStore) Put(state *oidcauth.OIDCState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state.StateID] = state
	return nil
}

func (s *memStateStore) Take(stateID string) (*oidcauth.OIDCState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.states[stateID]
	if !ok {
		return nil, nil
	}
	delete(s.states, stateID)
	return state, nil
}

func oidcAuthMethod(t *testing.T, srv *oidcauthtest.Server, extra map[string]interface{}) *structs.ACLAuthMethod {
	t.Helper()

	config := map[string]interface{}{
		"OIDCDiscoveryURL":    srv.Addr(),
		"OIDCDiscoveryCACert": srv.CACert(),
		"OIDCClientID":        "abc",
		"OIDCClientSecret":    "def",
		"AllowedRedirectURIs": []string{"https://example.com"},
		"BoundAudiences":      []string{"abc"},
		"JWTSupportedAlgs":    []string{"ES256"},
	}
	for k, v := range extra {
		config[k] = v
	}

	return &structs.ACLAuthMethod{
		Name:   "test-oidc",
		Type:   "oidc",
		Config: config,
	}
}

func TestOIDC_NewValidator(t *testing.T) {
	nullLogger := hclog.NewNullLogger()
	srv := oidcauthtest.Start(t)
	srv.SetClientCreds("abc", "def")

	t.Run("valid", func(t *testing.T) {
		v, err := NewValidator(nullLogger, oidcAuthMethod(t, srv, nil))
		require.NoError(t, err)
		require.NotNil(t, v)
		defer v.Stop()

		require.Equal(t, "oidc", v.methodType)
	})

	t.Run("missing client secret", func(t *testing.T) {
		method := oidcAuthMethod(t, srv, nil)
		delete(method.Config, "OIDCClientSecret")
		_, err := NewValidator(nullLogger, method)
		require.Error(t, err)
	})

	t.Run("missing discovery url", func(t *testing.T) {
		method := oidcAuthMethod(t, srv, nil)
		delete(method.Config, "OIDCDiscoveryURL")
		_, err := NewValidator(nullLogger, method)
		require.Error(t, err)
	})
}

func TestOIDC_AuthCodeFlow(t *testing.T) {
	nullLogger := hclog.NewNullLogger()
	srv := oidcauthtest.Start(t)
	srv.SetClientCreds("abc", "def")
	srv.SetAllowedRedirectURIs([]string{"https://example.com"})

	method := oidcAuthMethod(t, srv, map[string]interface{}{
		"ClaimMappings": map[string]string{
			"first_name":   "name",
			"/org/primary": "primary_org",
		},
		"ListClaimMappings": map[string]string{
			"https://consul.test/groups": "groups",
		},
	})

	v, err := NewValidator(nullLogger, method)
	require.NoError(t, err)
	defer v.Stop()

	ctx := oidcauth.WithStateStore(context.Background(), newMemStateStore())

	payload := map[string]string{"foo": "bar"}
	authURL, err := v.GetAuthCodeURL(ctx, "https://example.com", payload)
	require.NoError(t, err)
	require.Contains(t, authURL, srv.Addr()+"/auth?")

	state := oidcQueryParam(t, authURL, "state")
	nonce := oidcQueryParam(t, authURL, "nonce")

	srv.SetCustomClaims(map[string]interface{}{
		"nonce":      nonce,
		"first_name": "jeff",
		"org": map[string]interface{}{
			"primary": "engineering",
		},
		"https://consul.test/groups": []string{"foo", "bar"},
	})
	srv.SetExpectedAuthCode("abc")

	id, gotPayload, err := v.ClaimsFromAuthCode(ctx, state, "abc")
	require.NoError(t, err)
	require.Equal(t, payload, gotPayload)
	authmethod.RequireIdentityMatch(t, id,
		map[string]string{
			"value.name":        "jeff",
			"value.primary_org": "engineering",
		},
		"value.name == jeff",
		"value.primary_org == engineering",
		"foo in list.groups",
		"bar in list.groups",
		"salt not in list.groups",
	)

	// A state may only be consumed once.
	_, _, err = v.ClaimsFromAuthCode(ctx, state, "abc")
	require.Error(t, err)

	// Calling the OIDC-specific methods against a JWT method must fail.
	jwtMethod := &structs.ACLAuthMethod{
		Name: "test-jwt",
		Type: "jwt",
		Config: map[string]interface{}{
			"OIDCDiscoveryURL":    srv.Addr(),
			"OIDCDiscoveryCACert": srv.CACert(),
		},
	}
	jv, err := NewValidator(nullLogger, jwtMethod)
	require.NoError(t, err)
	defer jv.Stop()

	_, err = jv.GetAuthCodeURL(context.Background(), "https://example.com", nil)
	require.Error(t, err)

	_, _, err = jv.ClaimsFromAuthCode(context.Background(), "state", "code")
	require.Error(t, err)
}

func oidcQueryParam(t *testing.T, inputURL, param string) string {
	t.Helper()

	idx := -1
	for i := 0; i < len(inputURL); i++ {
		if inputURL[i] == '?' {
			idx = i
			break
		}
	}
	if idx == -1 {
		t.Fatalf("no query string in %q", inputURL)
	}

	m, err := url.ParseQuery(inputURL[idx+1:])
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m[param]
	if !ok {
		t.Fatalf("query param %q not found in %q", param, inputURL)
	}
	return v[0]
}
