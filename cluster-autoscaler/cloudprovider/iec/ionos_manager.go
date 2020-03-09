package iec

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"k8s.io/klog"
)

// Manager handles Ionos Enterprise Cloud communication and data caching of
// node groups (node pools in IEC)
type Manager struct {
	ionosClient  Client
	clusterID    string
	nodeGroups   []*NodeGroup
	pollTimeout  time.Duration
	pollInterval time.Duration
}

type Config struct {
	// IonosToken is the authentication token used by the Ionos Enterprise Cloud
	// Cluster Autoscaler to authenticate against the iec API
	IonosToken string `json:"ionos_token"`

	// IonosURL is the url of the Client used by the Ionos Cluster Autoscaler
	IonosURL string `json:"ionos_url"`

	// IonosURL is the url of the Client used by the Ionos Cluster Autoscaler
	IonosAuthURL string `json:"ionos_auth_url"`

	// ClusterID is the id associated with the cluster where Ionos Enterprise Cloud
	// Cluster Autoscaler is running
	ClusterID string `json:"cluster_id"`

	// PollTimeout is the timeout for polling a nodegroup after an update, e.g.
	// decreasing/increasing nodecount, until this update should have taken place.
	PollTimeout string `"json:"poll_timeout"`

	// PollInterval is the interval in which a nodegroup is polled after an update,
	// decreasing/increasing nodecount
	PollInterval string `"json:"poll_interval"`
}

func CreateIECManager(configReader io.Reader) (*Manager, error) {
	cfg := &Config{}
	if configReader != nil {
		body, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.IonosToken == "" {
		return nil, errors.New("iec access token is not provided")
	}

	if cfg.ClusterID == "" {
		return nil, errors.New("cluster ID is not provided")
	}

	timeout := 15 * time.Minute
	if cfg.PollTimeout != "" {
		t, err := time.ParseDuration(cfg.PollTimeout)
		if err != nil {
			return nil, errors.New("error parsing poll timeout")
		}
		timeout = t
	}
	interval := 5 * time.Second
	if cfg.PollInterval != "" {
		i, err := time.ParseDuration(cfg.PollInterval)
		if err != nil {
			return nil, errors.New("error parsing poll interval")
		}
		interval = i
	}

	iecClient, err := newClient(
		cfg.IonosToken,
		cfg.IonosURL,
		cfg.IonosAuthURL,
		timeout,
		interval)

	if err != nil {
		return nil, err
	}

	m := &Manager{
		ionosClient: iecClient,
		clusterID:   cfg.ClusterID,
		nodeGroups:  make([]*NodeGroup, 0),
	}

	return m, nil
}

// Refreshes the cache holding the nodegroups. This is called by the CA based
// on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh() error {
	klog.V(4).Info("Refreshing")
	ctx := context.Background()
	nodePools, err := m.ionosClient.ListNodePools(ctx, m.clusterID)
	if err != nil {
		klog.Errorf("eroor getting nodepools: %v", err)
		return err
	}

	var groups []*NodeGroup

	for _, nodePool := range nodePools {

		// TODO: Include this as soon as autoscaling limits are part of the nodepool on ionos side
		// AutoscalingLimit not set, autoscaling is disabled for this nodepool
//		if nodePool.AutoscalingLimit == nil {
//			klog.V(4).Infof("No autoscaling limit in nodepool, skipping: %v", *nodePool)
//			continue
//		}
//		min := *nodePool.AutoscalingLimit.Min
//		max := *nodePool.AutoscalingLimit.Max
		min := uint32(1)
		max := uint32(3)
		// AutoscalingLimit.Max == 0, autoscaling is disabled for this nodepool
		// AutoscalingLimit.Min cannot be 0, there is no way to scale out an empty nodepool
		if min == uint32(0) || max == uint32(0) {
			klog.V(4).Infof("Autoscaling limit min or max == 0, skipping: %v", nodePool)
			continue
		}
		name := nodePool.Properties.Name
		klog.V(4).Infof("adding node pool %q name: %s min: %d max %d", nodePool.ID, name, min, max)

		groups = append(groups, &NodeGroup{
			id:          nodePool.ID,
			clusterID:   m.clusterID,
			ionosClient: m.ionosClient,
			nodePool:    &nodePool,
			minSize:     int(min),
			maxSize:     int(max),
		})
	}
	klog.V(4).Infof("goups: %v", groups)

	if len(groups) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools configured")
	}

	m.nodeGroups = groups
	return nil
}
