// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package discoverychain

import (
	"github.com/opensoba/opensoba/acl"
	"github.com/opensoba/opensoba/agent/structs"
)

func (c *compiler) GetEnterpriseMeta() *acl.EnterpriseMeta {
	return structs.DefaultEnterpriseMetaInDefaultPartition()
}
