// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package structs

import (
	"github.com/opensoba/opensoba/acl"
	"github.com/opensoba/opensoba/api"
	"github.com/opensoba/opensoba/types"
)

// These are used to manage the built-in "serfHealth" check that's attached
// to every node in the catalog.
const (
	SerfCheckID           types.CheckID = "serfHealth"
	SerfCheckName                       = "Serf Health Status"
	SerfCheckAliveOutput                = "Agent alive and reachable"
	SerfCheckFailedOutput               = "Agent not live or unreachable"
)

var (
	// These are used to manage the service that's attached to every Consul
	// server node in the catalog.
	//
	// OpenSoba: the name is configurable through the "server_service_name"
	// agent configuration option so that a rebranded build can register
	// itself as e.g. "sabo". The default is still "consul" for backwards
	// compatibility. Everything that looks up the server service (DNS, mesh
	// gateways, the state store, ...) reads these variables so that the name
	// stays consistent across the agent.
	ConsulServiceID   = "consul"
	ConsulServiceName = "consul"
)

// SetConsulServiceName overrides the service name and ID that is registered in
// the catalog for every Consul server. An empty name is ignored so that the
// built-in default ("consul") is kept.
//
// This is a process-wide setting and is expected to be applied once during
// agent startup, from the resolved configuration.
func SetConsulServiceName(name string) {
	if name == "" {
		return
	}
	ConsulServiceName = name
	ConsulServiceID = name
}

type CatalogContents struct {
	Nodes    []*Node
	Services []*ServiceNode
	Checks   []*HealthCheck
}

type CatalogSummary struct {
	Nodes    []HealthSummary
	Services []HealthSummary
	Checks   []HealthSummary
}

type HealthSummary struct {
	Name string `json:",omitempty"`

	Total    int
	Passing  int
	Warning  int
	Critical int

	acl.EnterpriseMeta
}

func (h *HealthSummary) Add(status string) {
	h.Total++
	switch status {
	case api.HealthPassing:
		h.Passing++
	case api.HealthWarning:
		h.Warning++
	case api.HealthCritical:
		h.Critical++
	}
}

type AssignServiceManualVIPsRequest struct {
	Service    string
	ManualVIPs []string

	DCSpecificRequest
}

type AssignServiceManualVIPsResponse struct {
	Found          bool
	UnassignedFrom []PeeredServiceName
}
