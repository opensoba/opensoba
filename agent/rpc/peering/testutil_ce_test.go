// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package peering_test

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/opensoba/opensoba/agent/sabo"
)

func newDefaultDepsEnterprise(t *testing.T, logger hclog.Logger, c *sabo.Config) sabo.EnterpriseDeps {
	t.Helper()
	return sabo.EnterpriseDeps{}
}
