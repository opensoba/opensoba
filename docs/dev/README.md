# Local development environment

Spin up a complete OpenSoba environment on one machine: a single server agent
with ACLs enabled, the web UI, DNS, ACL policies and roles, optionally OIDC
single sign-on, and a few demo services that exercise the UI's external-source
logos.

Everything binds to `127.0.0.1`, and the tokens are hard-coded on purpose — this
is a throwaway lab, not a deployment guide. Delete the `data/` directory and you
are back to a clean slate.

## Contents

- [What you get](#what-you-get)
- [Prerequisites](#prerequisites)
- [1. Build the binary](#1-build-the-binary)
- [2. Start a server agent](#2-start-a-server-agent)
- [3. Provision the ACLs](#3-provision-the-acls)
- [4. OIDC single sign-on](#4-oidc-single-sign-on)
- [5. Demo services](#5-demo-services)
- [6. Provisioning with OpenTofu or Terraform](#6-provisioning-with-opentofu-or-terraform)
- [DNS and ACLs](#dns-and-acls)
- [Troubleshooting](#troubleshooting)
- [Cleanup](#cleanup)

## What you get

| Component           | Where                                                       |
| ------------------- | ----------------------------------------------------------- |
| HTTP API + web UI   | <http://localhost:8500/ui/>                                 |
| DNS                 | `127.0.0.1:8600`, records under `.service.sabo`             |
| ACL management token | `11111111-1111-1111-1111-111111111111`                      |
| OIDC auth method     | `keycloak` (created in step 4, `Log in with Keycloak`)      |

## Prerequisites

- [Go](https://go.dev/dl/) — use the version in `go.mod`.
- Docker, only if you want to rebuild the web UI.
- Optional: [OpenTofu](https://opentofu.org/) or
  [Terraform](https://developer.hashicorp.com/terraform) 1.5+ for step 6.
- Optional: a Keycloak (or any other OIDC) instance for step 4.

## 1. Build the binary

`make dev` installs the command line tool as `./bin/soba`, with the version and
git revision stamped in:

```shell
make dev
./bin/soba version
```

## 2. Start a server agent

The configuration in [`soba.hcl`](./soba.hcl) is a complete single-server setup:

- `server = true` with `bootstrap_expect = 1`, so it elects itself.
- ACLs enabled with `default_policy = "deny"`.
- All three bootstrap tokens (`initial_management`, `agent`, `default`) pinned to
  the same public value, so there is no bootstrap step.
  See [DNS and ACLs](#dns-and-acls) for why `default` matters and what to use
  instead outside of a lab.
- `server_service_name = "sabo"` and `domain = "sabo"`, so the server registers
  itself as `sabo` and DNS answers under `.service.sabo`. Both default to the
  upstream values (`consul`) if you drop the lines.
- `data_dir = "data"`, relative to the config file.

Run it from this directory:

```shell
mkdir -p data
../../bin/soba agent -config-file=soba.hcl -log-level=info
```

The UI is at <http://localhost:8500/ui/>. Because `acl.tokens.default` is set,
the agent answers token-less requests as well, so you can browse *everything*
without logging in — see [DNS and ACLs](#dns-and-acls).

In another shell, export the address and management token so the CLI commands
below do not need flags:

```shell
export CONSUL_HTTP_ADDR=http://localhost:8500
export CONSUL_HTTP_TOKEN=11111111-1111-1111-1111-111111111111
../../bin/soba members
../../bin/soba catalog services
```

> **Environment variables keep their upstream names.** `CONSUL_HTTP_ADDR`,
> `CONSUL_HTTP_TOKEN`, `CONSUL_HTTP_TOKEN_FILE`, `CONSUL_CACERT` and friends are
> unchanged, so existing tooling and scripts keep working.

## 3. Provision the ACLs

An OIDC login mints a token, and a token only gets permissions through a
**role**. A binding rule can bind an identity to a role, a service or a node —
never directly to a policy. So the chain is always:

```
policy  ->  role  ->  binding rule  ->  (login)
```

Create the write policy and the role that logins will be bound to:

```shell
../../bin/soba acl policy create -name=dev-admin -description="dev admin" -rules=@policies/admin.hcl
../../bin/soba acl role create -name=dev-admins -description="dev admins" -policy-name=dev-admin
../../bin/soba acl role list
```

`policies/admin.hcl` grants full write access and is only ever appropriate in a
lab. For a narrower policy, edit that file and re-run:

```shell
../../bin/soba acl policy update -name=dev-admin -rules=@policies/admin.hcl
```

## 4. OIDC single sign-on

The full reference for the auth method — every config parameter, client secret
handling, load balancer behaviour and the built-in fake provider used by the
tests — lives in [`docs/oidc.md`](../oidc.md). This section is the short version.

### 4.1 Register a client with your identity provider

In Keycloak (the same steps apply to any OIDC provider):

1. **Clients → Create client**, client type *OpenID Connect*.
2. Client ID: `opensoba` (must match `OIDCClientID` exactly).
3. **Client authentication: ON** — a confidential client, so you get a secret.
4. **Standard flow: ON.** Direct access grants can stay off.
5. Valid redirect URIs — copy them verbatim, including scheme, port and path:
   ```
   http://localhost:8500/ui/oidc/callback
   http://127.0.0.1:8500/ui/oidc/callback
   http://127.0.0.1:8250/oidc/callback
   ```
   The UI callback is derived from the address you open the UI with, so both
   `localhost` and `127.0.0.1` are listed. The third one belongs to the CLI.
6. **Save**, then **Credentials → Client secret** and copy the value.
7. Create a user (**Users → Add user**, then *Credentials → Set password* with
   *Temporary* off) so you have something to log in as.
8. The issuer is shown as `http://<host>:<port>/realms/<realm>`; that is the
   `OIDCDiscoveryURL`.

### 4.2 Create the auth method and the binding rule

The auth method has to exist before the binding rule, because the selector is
validated against the identity fields the auth method produces:

```shell
../../bin/soba acl auth-method create \
  -name=keycloak \
  -type=oidc \
  -display-name="Log in with Keycloak" \
  -max-token-ttl=8h \
  -config='{
    "OIDCDiscoveryURL": "https://keycloak.example.com/realms/myrealm",
    "OIDCClientID": "opensoba",
    "OIDCClientSecret": "<client-secret-from-step-4.1>",
    "OIDCScopes": ["openid", "profile", "email"],
    "AllowedRedirectURIs": [
      "http://localhost:8500/ui/oidc/callback",
      "http://127.0.0.1:8500/ui/oidc/callback",
      "http://127.0.0.1:8250/oidc/callback"
    ],
    "ClaimMappings": { "email": "email", "sub": "sub" },
    "ListClaimMappings": { "groups": "groups" }
  }'

# Every authenticated identity gets the role.
../../bin/soba acl binding-rule create \
  -method=keycloak -bind-type=role -bind-name=dev-admins \
  -description="any authenticated user"

# ...or be selective. The selector runs against the mapped claims:
#   -selector='value.email == "alice@example.com"'
#   -selector='list.groups contains "admins"'
```

`OIDCClientSecret` is write-only: it is never returned by the API, and omitting
it on an update preserves the stored value.

### 4.3 Log in

**Web UI** — open <http://localhost:8500/ui/>, pick the auth method on the login
screen and you are redirected to your identity provider. The login is a
full-page redirect rather than a popup, so popup blockers are not an issue. After
authenticating you land back on `/ui/oidc/callback`, the token is stored, and you
return to the UI still logged in.

**Command line** — `soba login` opens a browser and listens on loopback port
8250 for the redirect:

```shell
../../bin/soba login -method=keycloak -type=oidc -token-sink-file=./soba.token
CONSUL_HTTP_TOKEN_FILE=./soba.token ../../bin/soba members

# Revoke the token again when you are done. `logout` has no sink-file flag; it
# uses the token from the environment.
CONSUL_HTTP_TOKEN_FILE=./soba.token ../../bin/soba logout
```

## 5. Demo services

The UI shows a logo next to a service when its metadata carries an
`external-source` value the front end knows (see
`ui/packages/sabo-ui/app/helpers/service/external-source.js`). The three
services stand in for external products that register themselves: OpenBao,
OpenTofu and OpenWonton.

If you provision with OpenTofu or Terraform (step 6), `tofu apply` creates them
for you. Without OpenTofu, register the same three by hand:

```shell
for svc in openbao:8200 opentofu:8080 openwonton:4646; do
  name=${svc%%:*}; port=${svc##*:}
  ../../bin/soba services register \
    -name=demo-$name \
    -address=127.0.0.1 \
    -port=$port \
    -meta=external-source=$name
done
```

Either way, check that they are there:

```shell
../../bin/soba catalog services      # demo-openbao, demo-opentofu, demo-openwonton
dig @127.0.0.1 -p 8600 demo-openbao.service.sabo +short
```

Pick **one** of the two routes: the CLI registers the services through the
local agent (on the agent's own node), while OpenTofu registers them on a
dedicated, agent-less node — see
[step 6](#6-provisioning-with-opentofu-or-terraform) for why. Running both
leaves you with two instances of every service.

## 6. Provisioning with OpenTofu or Terraform

Everything from steps 3 to 5 can be applied declaratively instead. The
configuration in [`tofu/`](./tofu) uses the Consul provider, which speaks the
same HTTP API as OpenSoba:

```shell
cd tofu
tofu init          # or: terraform init
tofu apply         # or: terraform apply

# Add the OIDC auth method and binding rule as well:
tofu apply \
  -var enable_oidc=true \
  -var oidc_discovery_url=https://keycloak.example.com/realms/myrealm \
  -var oidc_client_secret=<client-secret>
```

Notes:

- `consul_address` / `consul_token` default to `127.0.0.1:8500` and the
  development management token; override with `-var` or `TF_VAR_...`.
- `enable_oidc=true` needs a *reachable* identity provider: creating an OIDC auth
  method makes the agent fetch
  `<discovery-url>/.well-known/openid-configuration` and refuses the method if
  that fails. Use the CLI from step 4 if you have no IdP handy.
- The OIDC client secret is write-only at the API level, so the auth method
  resource ignores changes to `config_json` to keep plans clean.
- The demo services from [step 5](#5-demo-services) are provisioned here as
  well ([`services.tf`](./tofu/services.tf)). They are registered on a
  dedicated, agent-less node (`demo-external`): the provider writes straight to
  the catalog, and a service placed on a node that runs an agent is pruned by
  anti-entropy. Because the provider reserves the `external-source` meta key
  for itself, the resource writes the value on create and then ignores later
  changes to `meta`.
- `tofu destroy` removes everything it created. It does not stop the agent.

## DNS and ACLs

DNS is the one part of the API that cannot be authenticated: a DNS query has
nowhere to put a token, so the agent evaluates it against the *default* token
(`acl.tokens.default`). With `default_policy = "deny"` and no default token,
every lookup answers NXDOMAIN — including for services that exist:

```shell
dig @127.0.0.1 -p 8600 demo-openbao.service.sabo +short
# => no output at all, even though the service is registered
```

That is why `soba.hcl` sets `acl.tokens.default`. It is set to the management
token for convenience, which has a real consequence worth understanding:

| Token-less request     | With `default` = management token | With `default` = read-only token |
| ---------------------- | --------------------------------- | -------------------------------- |
| DNS lookups            | resolve                            | resolve                          |
| UI without logging in  | full read **and write**            | full read, no write              |
| Calls with your own token | unaffected                      | unaffected                       |

For a setup that is closer to production, create a read-only token instead:

```shell
../../bin/soba acl token create -description="read-only" -policy-name=dns-read
# Copy the SecretID it prints into acl.tokens.default in soba.hcl, then restart
# the agent. `policies/dns-read.hcl` grants node and service reads.
```

Leave `acl.tokens.default` out entirely if you would rather have DNS return
nothing and force every client to authenticate.

## Troubleshooting

| Symptom | Cause and fix |
| --- | --- |
| `redirect_uri` mismatch at the provider | The redirect URI must be byte-identical in `AllowedRedirectURIs` *and* at the provider. `localhost` and `127.0.0.1` are different, and so are ports 8500 and 8250. |
| `iss` mismatch / discovery error | `OIDCDiscoveryURL` must equal the issuer the *browser* sees. Fix the proxy/frontend URL at the provider rather than the value in OpenSoba. |
| Audience error | Leave `BoundAudiences` empty; the client ID must appear in the ID token's `aud`. |
| Login screen shows no providers | ACLs are not enabled (`acl { enabled = true }`), or the auth method does not exist. Both `GET /v1/internal/ui/oidc-auth-methods` and `soba acl auth-method list` should show it. |
| `Invalid Auth Method: error checking OIDCDiscoveryURL` | The agent could not fetch `<discovery-url>/.well-known/openid-configuration` while creating the auth method. The URL must be reachable from the agent, not just from your browser. |
| Selector never matches | The claim is not mapped. Add it to `ClaimMappings` (single value, usable as `value.<name>`) or `ListClaimMappings` (list, usable as `list.<name>`). |
| Binding rule rejected | The auth method must exist first; the selector is validated against its identity fields. |
| Login button does nothing | The UI is served from a host that is not in `AllowedRedirectURIs`. |
| `soba login` fails | Port 8250 is already in use by another process. |
| DNS returns NXDOMAIN for everything | No `acl.tokens.default` (see [DNS and ACLs](#dns-and-acls)). |
| Client secret looks empty | By design: `OIDCClientSecret` is never returned by the API. |
| Logs are too quiet | Start the agent with `-log-level=debug`, and for OIDC troubleshooting set `"VerboseOIDCLogging": true` in the auth method config. Local only — it logs tokens. |

## Cleanup

Stop the agent with `Ctrl-C` (or `pkill -f "soba agent"`), then remove the lab
state:

```shell
rm -rf data soba.token
cd tofu && tofu destroy   # if you used the declarative route
```
