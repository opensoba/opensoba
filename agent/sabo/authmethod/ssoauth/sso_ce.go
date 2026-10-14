// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package ssoauth

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/opensoba/opensoba/acl"
	"github.com/opensoba/opensoba/agent/sabo/authmethod"
	"github.com/opensoba/opensoba/agent/structs"
	"github.com/opensoba/opensoba/internal/go-sso/oidcauth"
)

func init() {
	authmethod.Register("oidc", func(logger hclog.Logger, method *structs.ACLAuthMethod) (authmethod.Validator, error) {
		v, err := NewValidator(logger, method)
		if err != nil {
			return nil, err
		}
		return v, nil
	})
}

func validateType(typ string) error {
	switch typ {
	case "jwt", "oidc":
		return nil
	default:
		return fmt.Errorf("type should be %q or %q", "jwt", "oidc")
	}
}

func (v *Validator) ssoEntMetaFromClaims(_ *oidcauth.Claims) *acl.EnterpriseMeta {
	return nil
}

// enterpriseConfig holds the configuration only used by the OIDC authorization
// code flow. Upstream this lives in the enterprise build; OpenSoba enables the
// OIDC workflow in the MPL-2.0 build.
type enterpriseConfig struct {
	OIDCClientID        string   `json:",omitempty"`
	OIDCClientSecret    string   `json:",omitempty"`
	OIDCScopes          []string `json:",omitempty"`
	OIDCACRValues       []string `json:",omitempty"`
	AllowedRedirectURIs []string `json:",omitempty"`
	VerboseOIDCLogging  bool     `json:",omitempty"`
}

func (c *Config) enterpriseConvertForLibrary(ssoConfig *oidcauth.Config) {
	ssoConfig.OIDCClientID = c.OIDCClientID
	ssoConfig.OIDCClientSecret = c.OIDCClientSecret
	ssoConfig.OIDCScopes = c.OIDCScopes
	ssoConfig.OIDCACRValues = c.OIDCACRValues
	ssoConfig.AllowedRedirectURIs = c.AllowedRedirectURIs
	ssoConfig.VerboseOIDCLogging = c.VerboseOIDCLogging
}
