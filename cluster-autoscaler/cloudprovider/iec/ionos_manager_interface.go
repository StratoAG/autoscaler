package iec

import "time"

type IECManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh() error
	// GetNodesGroups
	GetNodeGroups() []*NodeGroup
	// GetClusterID
	GetClusterID() string
	// GetPollTimeout
	GetPollTimeout() time.Duration
	// GetPollInterval
	GetPollInterval() time.Duration
}
