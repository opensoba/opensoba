// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import "github.com/opensoba/opensoba/agent/config"

// ConfigReloader is a function type which may be implemented to support reloading
// of configuration.
type ConfigReloader func(rtConfig *config.RuntimeConfig) error
