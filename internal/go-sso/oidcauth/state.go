// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oidcauth

import (
	"context"
	"time"
)

// OIDCState is a unit of state that is created when an OIDC authorization code
// flow begins and must be carried through to the completion of the flow. It
// binds together the redirect URI, the OIDC nonce, and any application-specific
// payload for the duration of the flow.
type OIDCState struct {
	// StateID is a random identifier that is passed to the OIDC provider as the
	// "state" parameter and returned verbatim upon redirect.
	StateID string

	// Nonce is a random value that is passed to the OIDC provider as the
	// "nonce" parameter and later verified against the nonce claim in the
	// returned ID token.
	Nonce string

	// RedirectURI is the URI the OIDC provider will send the user back to.
	RedirectURI string

	// Payload is arbitrary application-specific data that was provided when the
	// state was created. It is opaque to the OIDC flow itself but is made
	// available again when the state is consumed.
	Payload interface{}

	// ExpiresAt is the time after which the state is no longer valid.
	ExpiresAt time.Time
}

// StateStore is an interface for storing and consuming OIDC flow state. It
// allows the state to be stored somewhere other than the in-process cache
// (e.g. replicated storage) so that multiple server instances can share OIDC
// login sessions.
//
// Implementations must treat Take as a one-time operation: a state that has
// been taken must not be returned again.
type StateStore interface {
	// Put stores a state. If a state with the same StateID already exists it
	// should be replaced.
	Put(state *OIDCState) error

	// Take atomically retrieves and removes a state by its StateID. It must
	// return (nil, nil) if no state exists for the given StateID.
	Take(stateID string) (*OIDCState, error)
}

type stateStoreCtxKey struct{}

// WithStateStore returns a context that carries a StateStore. If ctx already
// carries a StateStore the new one replaces it.
func WithStateStore(ctx context.Context, store StateStore) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, stateStoreCtxKey{}, store)
}

// stateStoreFromContext returns the StateStore carried by ctx, if any. A nil
// return value signals that the in-process cache should be used instead.
func stateStoreFromContext(ctx context.Context) StateStore {
	if ctx == nil {
		return nil
	}
	store, _ := ctx.Value(stateStoreCtxKey{}).(StateStore)
	return store
}
