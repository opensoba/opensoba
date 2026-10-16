// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dns

import (
	"math/rand"
	"regexp"
)

// MaxLabelLength is the maximum length for a name that can be used in DNS.
const MaxLabelLength = 63

// InvalidNameRe is a regex that matches characters which can not be included in
// a DNS name.
var InvalidNameRe = regexp.MustCompile(`[^A-Za-z0-9\\-]+`)

// InvalidNodeNameRe is a regex that matches characters which can not be
// included in a node name.
//
// Node names are allowed to contain dots, unlike service names. The default
// node name is the hostname of the machine, which on macOS (and on any other
// host configured with a search domain) is a FQDN such as "my-mac.local".
// Lookups for such nodes still work, because the DNS parser joins the
// remaining labels back together (see parseRequest in agent/dns.go).
var InvalidNodeNameRe = regexp.MustCompile(`[^A-Za-z0-9\-.]+`)

type RecursorStrategy string

const (
	RecursorStrategySequential RecursorStrategy = "sequential"
	RecursorStrategyRandom     RecursorStrategy = "random"
)

func (s RecursorStrategy) Indexes(max int) []int {
	switch s {
	case RecursorStrategyRandom:
		return rand.Perm(max)
	default:
		idxs := make([]int, max)
		for i := range idxs {
			idxs[i] = i
		}
		return idxs

	}
}
