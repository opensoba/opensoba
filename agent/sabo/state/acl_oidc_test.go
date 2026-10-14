// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opensoba/opensoba/agent/structs"
)

func TestStateStore_OIDCStateSet_Get_Delete(t *testing.T) {
	s := testStateStore(t)

	expiresAt := time.Now().Add(10 * time.Minute)
	state := &structs.ACLOIDCState{
		StateID:     "state-1",
		Nonce:       "nonce-1",
		RedirectURI: "https://example.com/oidc/callback",
		Payload:     []byte(`{"ClientNonce":"client-1"}`),
		ExpiresAt:   expiresAt,
	}

	require.NoError(t, s.OIDCStateSet(1, state))

	got, err := s.OIDCStateGet("state-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "state-1", got.StateID)
	require.Equal(t, "nonce-1", got.Nonce)
	require.Equal(t, "https://example.com/oidc/callback", got.RedirectURI)
	require.Equal(t, []byte(`{"ClientNonce":"client-1"}`), got.Payload)
	require.WithinDuration(t, expiresAt, got.ExpiresAt, time.Second)
	require.Equal(t, uint64(1), got.ModifyIndex)

	// Lookups are case-insensitive like other ID indexed tables.
	got, err = s.OIDCStateGet("STATE-1")
	require.NoError(t, err)
	require.NotNil(t, got)

	require.NoError(t, s.OIDCStateDelete(2, "state-1"))

	got, err = s.OIDCStateGet("state-1")
	require.NoError(t, err)
	require.Nil(t, got)

	// Deleting a missing state is a no-op.
	require.NoError(t, s.OIDCStateDelete(3, "state-1"))
}

func TestStateStore_OIDCStateBatchSet_BatchDelete(t *testing.T) {
	s := testStateStore(t)

	expiresAt := time.Now().Add(10 * time.Minute)
	states := []*structs.ACLOIDCState{
		{StateID: "state-1", Nonce: "n1", RedirectURI: "https://example.com/a", ExpiresAt: expiresAt},
		{StateID: "state-2", Nonce: "n2", RedirectURI: "https://example.com/b", ExpiresAt: expiresAt},
		{StateID: "state-3", Nonce: "n3", RedirectURI: "https://example.com/c", ExpiresAt: expiresAt},
	}
	require.NoError(t, s.OIDCStateBatchSet(1, states))

	for _, want := range states {
		got, err := s.OIDCStateGet(want.StateID)
		require.NoError(t, err)
		require.NotNil(t, got, "missing state %q", want.StateID)
		require.Equal(t, want.Nonce, got.Nonce)
	}

	require.NoError(t, s.OIDCStateBatchDelete(2, []string{"state-1", "state-3"}))

	got, err := s.OIDCStateGet("state-1")
	require.NoError(t, err)
	require.Nil(t, got)

	got, err = s.OIDCStateGet("state-2")
	require.NoError(t, err)
	require.NotNil(t, got)

	got, err = s.OIDCStateGet("state-3")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestStateStore_OIDCStateListExpired(t *testing.T) {
	s := testStateStore(t)

	now := time.Now()
	states := []*structs.ACLOIDCState{
		{StateID: "expired-1", Nonce: "n1", ExpiresAt: now.Add(-10 * time.Minute)},
		{StateID: "expired-2", Nonce: "n2", ExpiresAt: now.Add(-5 * time.Minute)},
		{StateID: "live-1", Nonce: "n3", ExpiresAt: now.Add(5 * time.Minute)},
		{StateID: "live-2", Nonce: "n4", ExpiresAt: now.Add(10 * time.Minute)},
	}
	require.NoError(t, s.OIDCStateBatchSet(1, states))

	expired, err := s.OIDCStateListExpired(now, 100)
	require.NoError(t, err)
	require.Len(t, expired, 2)
	require.Equal(t, "expired-1", expired[0].StateID)
	require.Equal(t, "expired-2", expired[1].StateID)

	// The max argument bounds the result set.
	expired, err = s.OIDCStateListExpired(now, 1)
	require.NoError(t, err)
	require.Len(t, expired, 1)
	require.Equal(t, "expired-1", expired[0].StateID)

	// Nothing is expired as of a point in the distant past.
	expired, err = s.OIDCStateListExpired(now.Add(-time.Hour), 100)
	require.NoError(t, err)
	require.Empty(t, expired)
}
