// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/opensoba/opensoba/agent/structs"
)

// oidcStatesTableSchema returns the memdb schema for the OIDC state table.
// OIDC state records are ephemeral and are never persisted in snapshots; they
// only need to survive long enough for a single authorization code flow to
// finish, which can span two HTTP requests landing on different servers.
func oidcStatesTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: tableACLOIDCStates,
		Indexes: map[string]*memdb.IndexSchema{
			indexID: {
				Name:         indexID,
				AllowMissing: false,
				Unique:       true,
				Indexer: indexerSingle[Query, *structs.ACLOIDCState]{
					readIndex:  indexFromQuery,
					writeIndex: indexStateIDFromACLOIDCState,
				},
			},
			indexExpiresAt: {
				Name:         indexExpiresAt,
				AllowMissing: true,
				Unique:       false,
				Indexer: indexerSingle[*TimeQuery, *structs.ACLOIDCState]{
					readIndex:  indexFromTimeQuery,
					writeIndex: indexExpiresAtFromACLOIDCState,
				},
			},
		},
	}
}

func indexStateIDFromACLOIDCState(state *structs.ACLOIDCState) ([]byte, error) {
	if state.StateID == "" {
		return nil, errMissingValueForIndex
	}

	var b indexBuilder
	b.String(strings.ToLower(state.StateID))
	return b.Bytes(), nil
}

func indexExpiresAtFromACLOIDCState(state *structs.ACLOIDCState) ([]byte, error) {
	if state.ExpiresAt.IsZero() {
		return nil, errMissingValueForIndex
	}

	var b indexBuilder
	b.Time(state.ExpiresAt)
	return b.Bytes(), nil
}

// OIDCStateSet inserts or updates an OIDC state record.
func (s *Store) OIDCStateSet(idx uint64, state *structs.ACLOIDCState) error {
	tx := s.db.WriteTxn(idx)
	defer tx.Abort()

	if err := oidcStateSetTxn(tx, idx, state); err != nil {
		return err
	}

	return tx.Commit()
}

// OIDCStateBatchSet inserts or updates a set of OIDC state records.
func (s *Store) OIDCStateBatchSet(idx uint64, states []*structs.ACLOIDCState) error {
	tx := s.db.WriteTxn(idx)
	defer tx.Abort()

	for _, state := range states {
		if err := oidcStateSetTxn(tx, idx, state); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func oidcStateSetTxn(tx WriteTxn, idx uint64, state *structs.ACLOIDCState) error {
	existing, err := oidcStateGetTxn(tx, state.StateID)
	if err != nil {
		return err
	}

	if existing != nil {
		state.CreateIndex = existing.CreateIndex
		state.ModifyIndex = idx
	} else {
		state.CreateIndex = idx
		state.ModifyIndex = idx
	}

	return tx.Insert(tableACLOIDCStates, state)
}

// OIDCStateDelete removes an OIDC state record by its state ID. If no such
// record exists this is a no-op.
func (s *Store) OIDCStateDelete(idx uint64, stateID string) error {
	tx := s.db.WriteTxn(idx)
	defer tx.Abort()

	if err := oidcStateDeleteTxn(tx, stateID); err != nil {
		return err
	}

	return tx.Commit()
}

// OIDCStateBatchDelete removes a set of OIDC state records by their state IDs.
func (s *Store) OIDCStateBatchDelete(idx uint64, stateIDs []string) error {
	tx := s.db.WriteTxn(idx)
	defer tx.Abort()

	for _, stateID := range stateIDs {
		if err := oidcStateDeleteTxn(tx, stateID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func oidcStateDeleteTxn(tx WriteTxn, stateID string) error {
	state, err := oidcStateGetTxn(tx, stateID)
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}

	if err := tx.Delete(tableACLOIDCStates, state); err != nil {
		return err
	}

	return nil
}

// OIDCStateGet looks up an OIDC state record by its state ID. A nil pointer
// is returned when no such record exists.
func (s *Store) OIDCStateGet(stateID string) (*structs.ACLOIDCState, error) {
	tx := s.db.ReadTxn()
	defer tx.Abort()

	return oidcStateGetTxn(tx, stateID)
}

func oidcStateGetTxn(tx ReadTxn, stateID string) (*structs.ACLOIDCState, error) {
	idxQuery := Query{Value: strings.ToLower(stateID)}
	raw, err := tx.First(tableACLOIDCStates, indexID, idxQuery)
	if err != nil {
		return nil, fmt.Errorf("failed oidc state lookup: %v", err)
	}
	if raw == nil {
		return nil, nil
	}

	return raw.(*structs.ACLOIDCState), nil
}

// OIDCStateListExpired lists OIDC state records that are expired as of the
// provided time. The returned set will be no larger than the max value
// provided.
func (s *Store) OIDCStateListExpired(asOf time.Time, max int) ([]*structs.ACLOIDCState, error) {
	tx := s.db.ReadTxn()
	defer tx.Abort()

	iter, err := tx.Get(tableACLOIDCStates, indexExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed oidc state listing: %v", err)
	}

	var (
		expired []*structs.ACLOIDCState
		i       int
	)
	for raw := iter.Next(); raw != nil; raw = iter.Next() {
		state := raw.(*structs.ACLOIDCState)
		if !state.ExpiresAt.Before(asOf) {
			return expired, nil
		}

		expired = append(expired, state)
		i += 1
		if i >= max {
			return expired, nil
		}
	}

	return expired, nil
}
