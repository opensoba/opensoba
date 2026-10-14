// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sabo

import "github.com/hashicorp/go-hclog"

// Operator endpoint is used to perform low-level operator tasks for Consul.
type Operator struct {
	srv    *Server
	logger hclog.Logger
}
