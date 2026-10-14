# OIDC login (single sign-on)

OpenSoba supports logging in with an OpenID Connect (OIDC) identity provider
(IdP) such as Keycloak, Okta, Auth0, Microsoft Entra ID, or Google. Users can
sign in to the web UI and to the `soba` command line using the browser-based
authorization code flow. No tokens need to be pasted by hand.

This page describes the supported configuration and a minimal end-to-end setup.

> **Note**
>
> The OIDC auth method type is available in the community edition of OpenSoba.
> It is based on the MPL-2.0 licensed `internal/go-sso/oidcauth` library that
> ships with the source tree.

## How it works

1. The UI or CLI calls `POST /v1/acl/oidc/auth-url` to begin a flow. The server
   stores the flow state (a random state ID, a nonce, the redirect URI, and an
   opaque payload) in the replicated state store and returns the IdP
   authorization URL.
2. The user's browser is sent to the IdP, authenticates, and is redirected back
   to the redirect URI with a `code` and `state`.
3. The UI or CLI calls `POST /v1/acl/oidc/callback` with the `state` and `code`.
   The server exchanges the code for an ID token, verifies it, applies the
   configured claim mappings, and mints an OpenSoba ACL token.

Because the flow state is stored in Raft (not in the memory of a single
process), a flow may start on one server and finish on another. This is what
makes OIDC work behind a load balancer that does not pin a client to a single
server. The state is one-time use: the callback consumes it.

## Config parameters

Create an auth method of type `oidc`:

```hcl
Name        = "keycloak"
Type        = "oidc"
DisplayName = "Log in with Keycloak"

Config = {
  OIDCDiscoveryURL    = "https://keycloak.example.com/realms/myrealm"
  OIDCDiscoveryCACert = "<optional PEM CA bundle, newline terminated>"
  OIDCClientID        = "opensoba"
  OIDCClientSecret    = "<client secret>"

  # Every redirect URI used by the UI and CLI must be listed here.
  AllowedRedirectURIs = [
    "https://soba.example.com/oidc/callback",
    "http://127.0.0.1:8250/oidc/callback",
  ]

  # Optional: map IdP claims into the login identity.
  ClaimMappings = {
    "email"      = "email"
    "first_name" = "name"
  }
}
```

Key parameters:

| Parameter             | Description                                                                                                   |
| --------------------- | ------------------------------------------------------------------------------------------------------------- |
| `OIDCDiscoveryURL`    | Issuer/discovery base URL, **without** the `/.well-known/openid-configuration` suffix. Required.             |
| `OIDCDiscoveryCACert` | PEM encoded CA bundle used to verify the IdP's TLS certificate. Optional; system roots are used if omitted.   |
| `OIDCClientID`        | The OAuth/OIDC client ID. Required.                                                                           |
| `OIDCClientSecret`    | The OAuth/OIDC client secret for a confidential client. Required.                                             |
| `AllowedRedirectURIs` | The exact redirect URIs the server will accept. Must include the UI and CLI callback URLs. Required.          |
| `BoundAudiences`      | Optional list of acceptable `aud` values for the ID token.                                                    |
| `ClaimMappings`       | Optional map of IdP claim (key) to identity attribute (value). Supports nested claims with `/` separators.    |
| `ListClaimMappings`   | Optional map of IdP claim (key) to list-typed identity attribute (value).                                     |
| `JWTSupportedAlgs`    | Optional list of accepted signing algorithms. Defaults to `RS256`.                                            |

### Redirect URIs

The client (browser) determines the redirect URI, so both the UI and the CLI
must be accounted for:

- **UI**: `https://<ui-host>/oidc/callback`. The UI derives `<ui-host>` from the
  URL the bundle is served from, so it is correct for whatever hostname you use
  to reach the web UI.
- **CLI**: `http://127.0.0.1:8250/oidc/callback`. The `soba login` command
  starts a temporary loopback listener on port 8250.

Add all of these to `AllowedRedirectURIs`, and register them with your IdP as
valid redirect URIs for the client. Loopback redirects are matched
port-agnostically by the server library, but they must still be present in
`AllowedRedirectURIs`.

### Client secret handling

`OIDCClientSecret` is treated as write-only:

- It is **required** when creating an auth method.
- It is **omitted** from any auth method returned by the HTTP API (for example by
  `soba acl auth-method read` or `GET /v1/acl/auth-method/:name`).
- On update, if the secret is omitted from the request, the server preserves the
  previously stored value. This makes a read/modify/write cycle safe.

## Binding rules

A login only succeeds if at least one binding rule matches the verified
identity. Create a binding rule for the auth method:

```hcl
AuthMethod = "keycloak"
Selector   = "value.email == \"alice@example.com\""
BindType   = "role"
BindName   = "admins"
```

The `Selector` is written against the identity produced by the claim mappings.
See the upstream auth method documentation for the full selector and binding
rule syntax.

## Logging in

### Web UI

Open the web UI and choose the auth method from the login screen. The list of
available OIDC providers is served pre-authentication from
`GET /v1/internal/ui/oidc-auth-methods`.

### Command line

```shell
soba login -method=keycloak -type=oidc -token-sink-file=./soba.token
```

The command prints an authorization URL, opens the default browser, and waits
for the IdP to redirect back to the local callback listener. On success, the
minted token's `SecretID` is written to the `-token-sink-file`.

Log out by destroying the token:

```shell
soba logout -token-sink-file=./soba.token
```

## Running behind a load balancer

OIDC login works behind an L4 load balancer that distributes requests across
servers:

- Flow state is written to the replicated state store, so it is visible to every
  server and survives a restart of the server that began the flow.
- The `auth-url` and `callback` requests may land on different servers.
- No sticky sessions are required.

The only requirement is that the browser can reach the redirect URI
(`https://<ui-host>/oidc/callback`), which is normally the same hostname used to
reach the UI. The CLI's loopback redirect never traverses the load balancer.

## Using the built-in test provider

Developers can exercise the flow without a real IdP using the in-tree fake
provider. It is only intended for tests and local development. The end-to-end
tests in `agent/oidc_endpoint_test.go` and
`agent/consul/acl_oidc_test.go` show a complete example, including the
load-balanced case with multiple servers.

## Reference

- `POST /v1/acl/oidc/auth-url`
- `POST /v1/acl/oidc/callback`
- `GET /v1/internal/ui/oidc-auth-methods`
- `soba login -type=oidc`
