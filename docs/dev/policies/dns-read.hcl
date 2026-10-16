# Read-only access to the catalog.
#
# Use this for the token that `acl.tokens.default` points to in a setup where
# ACLs are enabled and you still want DNS (and other token-less reads) to work.
# DNS requests cannot carry an ACL token, so they are evaluated against the
# default token; with `default_policy = "deny"` and no default token every DNS
# lookup answers NXDOMAIN.
#
# The local development config (docs/dev/soba.hcl) takes the shortcut of using
# the management token for this instead. Swap it for a token with this policy
# if you want the anonymous case to be read-only.

node_prefix "" {
  policy = "read"
}

service_prefix "" {
  policy = "read"
}
