// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inmem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opensoba/opensoba/internal/storage"
	"github.com/opensoba/opensoba/internal/storage/conformance"
	"github.com/opensoba/opensoba/internal/storage/inmem"
)

func TestBackend_Conformance(t *testing.T) {
	conformance.Test(t, conformance.TestOptions{
		NewBackend: func(t *testing.T) storage.Backend {
			backend, err := inmem.NewBackend()
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			go backend.Run(ctx)

			return backend
		},
		SupportsStronglyConsistentList: true,
	})
}
