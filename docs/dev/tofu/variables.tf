variable "consul_address" {
  description = "HTTP address of the OpenSoba agent."
  type        = string
  default     = "http://127.0.0.1:8500"
}

variable "consul_token" {
  description = "ACL token the provider authenticates with. Defaults to the development management token from docs/dev/soba.hcl."
  type        = string
  default     = "11111111-1111-1111-1111-111111111111"
  sensitive   = true
}

variable "enable_oidc" {
  description = "Create the OIDC auth method and binding rule. Requires oidc_discovery_url and oidc_client_secret."
  type        = bool
  default     = false
}

variable "oidc_discovery_url" {
  description = "OIDC issuer/discovery URL, without the /.well-known/openid-configuration suffix."
  type        = string
  default     = ""
}

variable "oidc_client_id" {
  description = "OAuth/OIDC client ID registered with the identity provider."
  type        = string
  default     = "opensoba"
}

variable "oidc_client_secret" {
  description = "OAuth/OIDC client secret for the confidential client."
  type        = string
  default     = ""
  sensitive   = true
}

variable "oidc_display_name" {
  description = "Label shown on the OpenSoba login screen."
  type        = string
  default     = "Log in with Keycloak"
}

variable "oidc_redirect_uris" {
  description = "Every redirect URI the UI and CLI use. These must also be registered with the identity provider."
  type        = list(string)
  default = [
    "http://localhost:8500/ui/oidc/callback",
    "http://127.0.0.1:8500/ui/oidc/callback",
    "http://127.0.0.1:8250/oidc/callback",
  ]
}

variable "oidc_bind_selector" {
  description = "Binding rule selector for the OIDC login. Empty matches every authenticated identity."
  type        = string
  default     = ""
}
