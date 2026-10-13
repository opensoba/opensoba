// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

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

const synopsis = "Interact with OpenSoba Connect"
const help = `
Usage: soba connect <subcommand> [options] [args]

  This command has subcommands for interacting with OpenSoba Connect.

  Here are some simple examples, and more detailed examples are available
  in the subcommands or the documentation.

  Run the built-in Connect mTLS proxy

      $ soba connect proxy

  For more examples, ask for subcommand help or view the documentation.
`
