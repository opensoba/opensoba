// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package troubleshoot

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

const synopsis = `CLI tools for troubleshooting OpenSoba service mesh`
const help = `
Usage: soba troubleshoot <subcommand> [options]

  This command has subcommands for troubleshooting the service mesh.

  Here are some simple examples, and more detailed examples are available
  in the subcommands or the documentation.

  Troubleshoot Get Upstreams

    $ soba troubleshoot upstreams

  Troubleshoot Proxy

    $ soba troubleshoot proxy -upstream [options]

  For more examples, ask for subcommand help or view the documentation.
`
