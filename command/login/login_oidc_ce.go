// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package login

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/hashicorp/go-uuid"

	"github.com/opensoba/opensoba/api"
)

// oidcCallbackAddr is the localhost address the CLI listens on while waiting
// for the OIDC provider to redirect the browser back. The redirect URI derived
// from this must be listed in the auth method's AllowedRedirectURIs.
//
// It is a var so that tests can bind an ephemeral port.
var oidcCallbackAddr = "127.0.0.1:8250"

// oidcLogin performs the OIDC authorization code flow using the user's browser.
// It starts a temporary localhost server, directs the user to the provider, and
// exchanges the returned code for an OpenSoba ACL token via the
// /v1/acl/oidc endpoints.
func (c *cmd) oidcLogin() int {
	// Ensure that we don't try to use a token when performing a login
	// operation.
	c.http.SetToken("")
	c.http.SetTokenFile("")

	client, err := c.http.APIClient()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error connecting to OpenSoba agent: %s", err))
		return 1
	}

	clientNonce, err := uuid.GenerateUUID()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error generating nonce: %s", err))
		return 1
	}

	redirectURI := fmt.Sprintf("http://%s/oidc/callback", oidcCallbackAddr)

	listener, err := net.Listen("tcp", oidcCallbackAddr)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error starting OIDC callback listener on %s: %s", oidcCallbackAddr, err))
		return 1
	}
	defer listener.Close()

	type callbackResult struct {
		code  string
		state string
		err   error
	}
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oidc/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errCode := q.Get("error"); errCode != "" {
			err := fmt.Errorf("OIDC provider returned an error: %s: %s", errCode, q.Get("error_description"))
			resultCh <- callbackResult{err: err}
			http.Error(w, "Login failed. You may close this window.", http.StatusBadRequest)
			return
		}
		resultCh <- callbackResult{code: q.Get("code"), state: q.Get("state")}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Login complete. You may close this window and return to the terminal.")
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	authURL, _, err := client.ACL().OIDCAuthURL(&api.ACLOIDCAuthURLParams{
		AuthMethod:  c.authMethodName,
		RedirectURI: redirectURI,
		ClientNonce: clientNonce,
		Meta:        c.meta,
	}, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error requesting OIDC auth URL: %s", err))
		return 1
	}

	c.UI.Output("")
	c.UI.Output("Please open the following URL in your browser to complete authentication:")
	c.UI.Output("")
	c.UI.Output("  " + authURL)
	c.UI.Output("")
	openURL(authURL)

	var res callbackResult
	select {
	case res = <-resultCh:
	case <-time.After(5 * time.Minute):
		c.UI.Error("Timed out waiting for the OIDC provider callback.")
		return 1
	}
	if res.err != nil {
		c.UI.Error(res.err.Error())
		return 1
	}
	if res.code == "" || res.state == "" {
		c.UI.Error("OIDC callback did not include a code and state.")
		return 1
	}

	tok, _, err := client.ACL().OIDCCallback(&api.ACLOIDCCallbackParams{
		AuthMethod:  c.authMethodName,
		State:       res.state,
		Code:        res.code,
		ClientNonce: clientNonce,
	}, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error logging in: %s", err))
		return 1
	}

	if err := c.writeToSink(tok); err != nil {
		c.UI.Error(fmt.Sprintf("Error writing token to file sink: %s", err))
		return 1
	}

	return 0
}

// openURL makes a best-effort attempt to open the provided URL in the user's
// default browser. Failures are ignored since the URL is also printed.
func openURL(url string) {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name = "open"
	case "windows":
		name = "rundll32"
		args = append(args, "url.dll,FileProtocolHandler")
	default:
		name = "xdg-open"
	}
	args = append(args, url)
	_ = exec.Command(name, args...).Start()
}
