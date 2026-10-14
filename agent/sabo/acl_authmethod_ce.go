// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package sabo

func (s *Server) enterpriseEvaluateRoleBindings() error {
	return nil
}
