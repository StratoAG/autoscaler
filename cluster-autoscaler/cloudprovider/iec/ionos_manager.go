package iec

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"k8s.io/klog"
)

// IECManagerImpl handles Ionos Enterprise Cloud communication and data caching of
// node groups (node pools in IEC)
type IECManagerImpl struct {
	ionosClient  client
	clusterID    string
	nodeGroups   []*NodeGroup
}

var (
	read = ioutil.ReadAll
	createClient = newClient
)

const (
	ionosToken = "IONOS_TOKEN"
	ionosURL = "IONOS_URL"
	ionosAuthURL = "IONOS_AUTH_URL"
	ionosClusterID = "IONOS_CLUSTER_ID"
	ionosPollTimeout = "IONOS_POLL_TIMEOUT"
	ionosPollInterval = "IONOS_POLL_INTERVAL"
)

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
	PollTimeout string `json:"poll_timeout"`

	// PollInterval is the interval in which a nodegroup is polled after an update,
	// decreasing/increasing nodecount
	PollInterval string `json:"poll_interval"`
}

func (c *Config) processConfig() (timeout, interval time.Duration, err error) {
	klog.V(4).Info("Processing config")
	var result *multierror.Error
	if c.IonosToken ==  "" {
		if value, ok := os.LookupEnv(ionosToken); ok {
			klog.V(4).Info("Got token in env")
			c.IonosToken = value
		} else {
			klog.V(5).Info("Failed to get IEC access token.")
			 result = multierror.Append(result, errors.New("IOC access token is not provided"))
		}
	}

	if c.IonosURL == "" {
		if value, ok := os.LookupEnv(ionosURL); ok {
			klog.V(4).Info("Got url in env")
			c.IonosURL = value
		} else {
		}
	}

	if c.IonosAuthURL == "" {
		if value, ok := os.LookupEnv(ionosAuthURL); ok {
			klog.V(4).Info("Got auth_url in env")
			c.IonosAuthURL = value
		}
	}

	if c.ClusterID == "" {
		if value, ok := os.LookupEnv(ionosClusterID); ok {
			klog.V(4).Info("Got clusterID in env")
			c.ClusterID = value
		} else {
			klog.V(5).Info("Failed to get cluster ID.")
			result = multierror.Append(errors.New("cluster id is not provided"), )
		}
	}

	if c.PollTimeout == "" {
		if value, ok := os.LookupEnv(ionosPollTimeout); ok {
			klog.V(4).Info("Got poll timeout in env")
			c.PollTimeout = value
		}
	}
	timeout = 20 * time.Minute
	if c.PollTimeout != "" {
		t, err := time.ParseDuration(c.PollTimeout)
		if err != nil {
			result = multierror.Append(result, errors.New("error parsing poll timeout"))
		}
		timeout = t
	}

	if c.PollInterval == "" {
		if value, ok := os.LookupEnv(ionosPollInterval); ok {
			klog.V(4).Info("Got poll interval in env")
			c.PollInterval = value
		}
	}
	interval = 20 * time.Second
	if c.PollInterval != "" {
		i, err := time.ParseDuration(c.PollInterval)
		if err != nil {
			result = multierror.Append(result, errors.New("error parsing poll interval"))
		}
		interval = i
	}
	fmt.Printf("Result: %v", result)

	return timeout, interval, result.ErrorOrNil()
}

func CreateIECManager(configReader io.Reader) (IECManager, error) {
	klog.V(4).Info("Creating IEC manager")
	cfg := &Config{}
	if configReader != nil {
		klog.V(5).Info("Reading from config reader")
		body, err := read(configReader)
		if err != nil {
			klog.Error("Error reading from config reader")
			return nil, err
		}
		klog.V(5).Infof("Unmarshaling config json: %s", string(body))
		err = json.Unmarshal(body, cfg)
		if err != nil {
			klog.Errorf("Error unmarshaling config json: %s", body)
			return nil, err
		}
	}

	timeout, interval, err := cfg.processConfig()
	fmt.Printf("Error: %v\n", err)

	if err != nil {
		fmt.Println("Error processing config")
		klog.V(5).Infof("Error processing config, %v", err)
		return nil, err
	}

	iecClient, err := createClient(
		cfg.IonosToken,
		cfg.IonosURL,
		cfg.IonosAuthURL,
		timeout,
		interval)

	if err != nil {
		klog.Errorf("Error creating IEC client, %v", err)
		return nil, err
	}

	m := &IECManagerImpl{
		ionosClient: client{
			Client: iecClient,
			pollInterval: interval,
			pollTimeout: timeout,
		},
		clusterID:   cfg.ClusterID,
		nodeGroups:  make([]*NodeGroup, 0),
	}

	return m, nil
}

// Refreshes the cache holding the nodegroups. This is called by the CA based
// on the `--scan-interval`. By default it's 10 seconds.
func (m *IECManagerImpl) Refresh() error {
	klog.V(4).Info("Refreshing")
	nodePools, err := m.ionosClient.ListKubernetesNodePools(m.clusterID)
	if err != nil {
		klog.Errorf("error getting nodepools: %v", err)
		return err
	}

	var groups []*NodeGroup

	for _, nodePool := range nodePools.Items {
		klog.V(4).Infof("Processing nodepool: %s", nodePool.ID)
		if nodePool.Properties.Autoscaling == nil {
			klog.V(4).Infof("No autoscaling limit in nodepool, skipping: %v", nodePool)
			continue
		}
		min := nodePool.Properties.Autoscaling.MinNodeCount
		max := nodePool.Properties.Autoscaling.MaxNodeCount
		// AutoscalingLimit.Max == 0, autoscaling is disabled for this nodepool
		// AutoscalingLimit.Min cannot be 0, there is no way to scale out an empty nodepool
		if min == uint32(0) && max == uint32(0) {
			klog.V(4).Infof("Autoscaling limit min and max == 0, skipping: %v", nodePool)
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
	klog.V(4).Infof("groups: %+v", groups)

	if len(groups) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools configured")
	}

	m.nodeGroups = groups
	return nil
}

func (m *IECManagerImpl) GetNodeGroups() []*NodeGroup {
	return m.nodeGroups
}

func (m *IECManagerImpl) GetClusterID() string {
	return m.clusterID
}

func (m *IECManagerImpl) GetPollTimeout() time.Duration {
	return m.ionosClient.PollTimeout()
}

func (m *IECManagerImpl) GetPollInterval() time.Duration {
	return m.ionosClient.PollInterval()
}
