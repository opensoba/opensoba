// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package login

import (
	"path/filepath"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/require"

	"github.com/opensoba/opensoba/sdk/testutil"
)

// TestLoginCommand_oidc verifies that "-type=oidc" routes through the browser
// based OIDC flow rather than the bearer token flow.
//
// A full browser round-trip is exercised by the server side tests
// (agent/oidc_endpoint_test.go and agent/consul/acl_oidc_test.go). Here we only
// assert the dispatch: when the auth method is missing the command fails before
// opening a browser, which makes the test hermetic.
func TestLoginCommand_oidc(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	// Bind an ephemeral port so the test does not depend on 8250 being free.
	origAddr := oidcCallbackAddr
	oidcCallbackAddr = "127.0.0.1:0"
	t.Cleanup(func() { oidcCallbackAddr = origAddr })

	testDir := testutil.TempDir(t, "acl")
	tokenSinkFile := filepath.Join(testDir, "test.token")

	a := newTestAgent(t)

	ui := cli.NewMockUi()
	cmd := New(ui)

	code := cmd.Run([]string{
		"-http-addr=" + a.HTTPAddr(),
		"-method=missing-method",
		"-type=oidc",
		"-token-sink-file", tokenSinkFile,
	})
	require.Equal(t, 1, code)
	require.Contains(t, ui.ErrorWriter.String(), "missing-method")
	require.NoFileExists(t, tokenSinkFile)
}
