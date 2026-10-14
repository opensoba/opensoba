// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package testauth

import "github.com/opensoba/opensoba/acl"

type enterpriseConfig struct{}

func (v *Validator) testAuthEntMetaFromFields(fields map[string]string) *acl.EnterpriseMeta {
	return nil
}
