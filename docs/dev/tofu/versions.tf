terraform {
  required_version = ">= 1.5"

  required_providers {
    # The provider speaks the Consul HTTP API, which OpenSoba implements, so it
    # works unmodified against a `soba` agent.
    consul = {
      source  = "hashicorp/consul"
      version = "~> 2.0"
    }
  }
}
