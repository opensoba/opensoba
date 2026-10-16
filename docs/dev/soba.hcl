# OpenSoba local development server.
#
#   ../../bin/soba agent -config-file=soba.hcl -log-level=info
#
# Run this from the directory that contains this file: `data_dir` is relative,
# and the whole environment is meant to be throwaway. See README.md.

datacenter = "dc1"
data_dir   = "data"

# Branding: register the server in the catalog as "sabo" and serve DNS records
# under `.service.sabo` instead of `.service.consul`. Both options default to
# the upstream values ("consul"), so existing configurations are unaffected.
server_service_name = "sabo"
domain              = "sabo"

server           = true
bootstrap_expect = 1

bind_addr   = "127.0.0.1"
client_addr = "127.0.0.1"

ui_config {
  enabled = true
}

acl {
  enabled        = true
  default_policy = "deny"

  tokens {
    # Management token. Hard-coded on purpose: this environment is public and
    # disposable, so nobody has to run `soba acl bootstrap`.
    initial_management = "11111111-1111-1111-1111-111111111111"

    # Token the agent uses for its own operations (node info, coordinates,
    # anti-entropy, ...). Without it the log fills up with "blocked by ACLs"
    # warnings and anti-entropy cannot clean anything up.
    agent = "11111111-1111-1111-1111-111111111111"

    # Token used for requests that carry no token of their own. DNS cannot send
    # a token at all, so with `default_policy = "deny"` every DNS lookup returns
    # NXDOMAIN unless this is set.
    #
    # DEV ONLY: reusing the management token here means anonymous API clients can
    # read *and write* everything. That is convenient locally and unacceptable
    # anywhere else; see "DNS and ACLs" in README.md for the read-only variant.
    default = "11111111-1111-1111-1111-111111111111"
  }
}
