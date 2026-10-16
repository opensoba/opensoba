# Full write access. Intended for the role that OIDC logins are bound to in the
# local development environment. Do not use this outside of a lab.

acl      = "write"
operator = "write"
keyring  = "write"

agent_prefix "" {
  policy = "write"
}

event_prefix "" {
  policy = "write"
}

key_prefix "" {
  policy = "write"
}

node_prefix "" {
  policy = "write"
}

query_prefix "" {
  policy = "write"
}

service_prefix "" {
  policy    = "write"
  intentions = "write"
}

session_prefix "" {
  policy = "write"
}
