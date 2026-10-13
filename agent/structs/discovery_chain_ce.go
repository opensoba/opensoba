// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package structs

import (
	"github.com/opensoba/opensoba/acl"
)

func (t *DiscoveryTarget) GetEnterpriseMetadata() *acl.EnterpriseMeta {
	return DefaultEnterpriseMetaInDefaultPartition()
}
