provider "consul" {
  address = var.consul_address
  token   = var.consul_token
}

# ---------------------------------------------------------------------------
# ACL: policy -> role -> binding rule
#
# Binding rules can only bind an identity to a role, a service or a node, never
# straight to a policy. Hence the policy + role pair below.
# ---------------------------------------------------------------------------

resource "consul_acl_policy" "admin" {
  name        = "dev-admin"
  description = "Local development: full write access"
  rules       = file("${path.module}/../policies/admin.hcl")
}

resource "consul_acl_role" "admins" {
  name        = "dev-admins"
  description = "Local development administrators"
  policies    = [consul_acl_policy.admin.id]
}

# Read-only catalog access for the token that `acl.tokens.default` points at.
# See the "DNS and ACLs" section of docs/dev/README.md.
resource "consul_acl_policy" "dns_read" {
  name        = "dns-read"
  description = "Read-only catalog access so that DNS and token-less reads work"
  rules       = file("${path.module}/../policies/dns-read.hcl")
}

# ---------------------------------------------------------------------------
# OIDC single sign-on
#
# Creating an OIDC auth method makes the agent fetch the provider's discovery
# document, so `enable_oidc = true` requires a reachable identity provider.
# ---------------------------------------------------------------------------

resource "consul_acl_auth_method" "oidc" {
  count = var.enable_oidc ? 1 : 0

  name          = "keycloak"
  type          = "oidc"
  display_name  = var.oidc_display_name
  description   = "OIDC single sign-on for local development"
  max_token_ttl = "8h"

  config_json = jsonencode({
    OIDCDiscoveryURL    = var.oidc_discovery_url
    OIDCClientID        = var.oidc_client_id
    OIDCClientSecret    = var.oidc_client_secret
    OIDCScopes          = ["openid", "profile", "email"]
    AllowedRedirectURIs = var.oidc_redirect_uris
    ClaimMappings = {
      email = "email"
      sub   = "sub"
    }
    ListClaimMappings = {
      groups = "groups"
    }
  })

  lifecycle {
    # OIDCClientSecret is write-only: the API never returns it, so without this
    # the provider reports a diff on every plan. Remove the ignore (or taint the
    # resource) when you change any of the other settings.
    ignore_changes = [config_json]
  }
}

resource "consul_acl_binding_rule" "oidc" {
  count = var.enable_oidc ? 1 : 0

  auth_method = consul_acl_auth_method.oidc[0].name
  description = "Local development OIDC login"
  selector    = var.oidc_bind_selector
  bind_type   = "role"
  bind_name   = consul_acl_role.admins.name
}
