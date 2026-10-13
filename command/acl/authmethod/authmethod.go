// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package authmethod

import (
	"github.com/mitchellh/cli"
	"github.com/opensoba/opensoba/command/flags"
)

func New() *cmd {
	return &cmd{}
}

type cmd struct{}

func (c *cmd) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return flags.Usage(help, nil)
}

const synopsis = "Manage OpenSoba's ACL auth methods"
const help = `
Usage: soba acl auth-method <subcommand> [options] [args]

  This command has subcommands for managing OpenSoba's ACL auth methods.
  Here are some simple examples, and more detailed examples are available in
  the subcommands or the documentation.

  Create a new auth method:

    $ soba acl auth-method create -type "kubernetes" \
                            -name "my-k8s" \
                            -description "This is an example kube auth method" \
                            -kubernetes-host "https://apiserver.example.com:8443" \
                            -kubernetes-ca-cert @/path/to/kube.ca.crt \
                            -kubernetes-service-account-jwt "JWT_CONTENTS"

  List all auth methods:

    $ soba acl auth-method list

  Update all editable fields of the auth method:

    $ soba acl auth-method update -name "my-k8s" \
                            -description "new description" \
                            -kubernetes-host "https://new-apiserver.example.com:8443" \
                            -kubernetes-ca-cert @/path/to/new-kube.ca.crt \
                            -kubernetes-service-account-jwt "NEW_JWT_CONTENTS"

  Read an auth method:

    $ soba acl auth-method read -name my-k8s

  Delete an auth method:

    $ soba acl auth-method delete -name my-k8s

  For more examples, ask for subcommand help or view the documentation.
`
