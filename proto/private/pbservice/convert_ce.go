// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package pbservice

import (
	"github.com/opensoba/opensoba/acl"
	"github.com/opensoba/opensoba/proto/private/pbcommon"
)

func EnterpriseMetaToStructs(_ *pbcommon.EnterpriseMeta) acl.EnterpriseMeta {
	return acl.EnterpriseMeta{}
}

func NewEnterpriseMetaFromStructs(_ acl.EnterpriseMeta) *pbcommon.EnterpriseMeta {
	return &pbcommon.EnterpriseMeta{}
}
