// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oidcauth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// testStateStore is an in-memory StateStore used to exercise the pluggable
// state seam without touching the process-wide cache.
type testStateStore struct {
	mu     sync.Mutex
	states map[string]*OIDCState
	puts   int
	takes  int
}

func newTestStateStore() *testStateStore {
	return &testStateStore{states: make(map[string]*OIDCState)}
}

func (s *testStateStore) Put(state *OIDCState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.puts++
	s.states[state.StateID] = state
	return nil
}

func (s *testStateStore) Take(stateID string) (*OIDCState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.takes++
	state, ok := s.states[stateID]
	if !ok {
		return nil, nil
	}
	delete(s.states, stateID)
	return state, nil
}

func TestOIDC_StateStore(t *testing.T) {
	oa, srv := setupForOIDC(t)

	store := newTestStateStore()
	ctx := WithStateStore(context.Background(), store)

	origPayload := map[string]string{"foo": "bar"}
	authURL, err := oa.GetAuthCodeURL(ctx, "https://example.com", origPayload)
	require.NoError(t, err)

	state := getQueryParam(t, authURL, "state")
	nonce := getQueryParam(t, authURL, "nonce")

	// The library must have delegated storage to the provided store rather than
	// its own in-process cache.
	store.mu.Lock()
	require.Equal(t, 1, store.puts)
	require.Len(t, store.states, 1)
	store.mu.Unlock()

	srv.SetCustomClaims(sampleClaims(nonce))
	srv.SetExpectedAuthCode("abc")

	claims, payload, err := oa.ClaimsFromAuthCode(ctx, state, "abc")
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, origPayload, payload)

	// The state must be removed from the store when consumed.
	store.mu.Lock()
	require.Equal(t, 1, store.takes)
	require.Empty(t, store.states)
	store.mu.Unlock()

	// A state may only be used once; replaying it must fail.
	_, _, err = oa.ClaimsFromAuthCode(ctx, state, "abc")
	require.Error(t, err)
	var provErr *ProviderLoginFailedError
	require.True(t, errors.As(err, &provErr), "expected a *ProviderLoginFailedError, got: %v", err)
}
