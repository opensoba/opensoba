// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package raft

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

const synopsis = "Provides cluster-level tools for OpenSoba operators"
const help = `
Usage: soba operator raft <subcommand> [options]

The Raft operator command is used to interact with OpenSoba's Raft subsystem. The
command can be used to verify Raft peers or in rare cases to recover quorum by
removing invalid peers.
`
