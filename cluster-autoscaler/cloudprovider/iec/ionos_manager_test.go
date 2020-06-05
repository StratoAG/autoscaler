package iec

import (
	"bytes"
	"errors"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/iec/mocks"
	"testing"
	"time"
)

var (
	pollTimeout = time.Millisecond * 10
	pollInterval = time.Millisecond * 10
	nodePools =[]profitbricks.KubernetesNodePool{
		{
			ID:         "1",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				Name:             "nodepool-1",
				NodeCount:        2,
				Autoscaling: &profitbricks.Autoscaling{
					MinNodeCount: 1,
					MaxNodeCount: 3,
				},
			},
		},
		{
			ID:         "2",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				Name:             "nodepool-2",
				NodeCount:        2,
				Autoscaling: &profitbricks.Autoscaling{
					MinNodeCount: 1,
					MaxNodeCount: 2,
				},
			},
		},
		{
			ID:         "3",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				Name:             "nodepool-3",
				NodeCount:        2,
				Autoscaling: &profitbricks.Autoscaling{
					MinNodeCount: 0,
					MaxNodeCount: 0,
				},
			},
		},
		{
			ID:         "4",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				Name:             "nodepool-4",
				NodeCount:        2,
			},
		},
	}
	nodes = []profitbricks.KubernetesNode{
		{
			ID: "1",
			Metadata: &profitbricks.Metadata{
				State: profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name: "node-1-1",
			},
		}, {
			ID: "2",
			Metadata: &profitbricks.Metadata{
				State: profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name: "node-1-2",
			},
		}, {
			ID:         "3",
			Metadata:   &profitbricks.Metadata{
				State:  profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name:   "node-2-3",
			},
		},
		{
			ID:         "4",
			Metadata:   &profitbricks.Metadata{
				State:  profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name:   "node-2-4",
			},
		}, {
			ID: "5",
			Metadata: &profitbricks.Metadata{
				State: profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name: "node-3-5",
			},
		},
		{
			ID: "6",
			Metadata: &profitbricks.Metadata{
				State: profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name: "node-3-6",
			},
		}, {
			ID:         "7",
			Metadata:   &profitbricks.Metadata{
				State:  profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name:   "node-4-7",
			},
		},
		{
			ID:         "8",
			Metadata:   &profitbricks.Metadata{
				State:  profitbricks.StateAvailable,
			},
			Properties: &profitbricks.KubernetesNodeProperties{
				Name:   "node-4-8",
			},
		},
	}
)

func TestNewManager(t *testing.T) {

	t.Run("success, manager creation", func(t *testing.T) {
		cfg := `{"cluster_id": "12345", "ionos_token": "secret_ionos_token", "ionos_url": "https://api.ionos.com", "ionos_auth_url": "https://auth.ionos.com"}`
		manager, err := CreateIECManager(bytes.NewBufferString(cfg))
		assert.NoError(t, err)
		assert.Equal(t, "12345", manager.GetClusterID(), "cluster ID does not match")
	})

	t.Run("success, manager creation with durations", func(t *testing.T) {
		cfg := `{"cluster_id": "12345", "ionos_token": "secret_ionos_token", "poll_timeout": "300ms", "poll_interval": "3s"}`
		manager, err := CreateIECManager(bytes.NewBufferString(cfg))
		assert.NoError(t, err)
		assert.Equal(t, "12345", manager.GetClusterID(), "cluster ID does not match")
		assert.Equal(t, time.Millisecond*300, manager.GetPollTimeout(), "pollTimeout does not match")
		assert.Equal(t, time.Second*3, manager.GetPollInterval(), "pollInterval does not match")
	})

	t.Run("failure, read cfg error", func(t *testing.T) {
		read = func(r io.Reader) ([]byte, error) {
			return []byte{}, errors.New("oops, something went wrong")
		}
		_, err := CreateIECManager(bytes.NewBufferString(""))
		assert.Error(t, err, "error expected")
		read = ioutil.ReadAll
	})

	tests := []struct {
		name,
		cfg string
	}{
		{
			name: "failure, empty cluster ID",
			cfg:  `{"cluster_id": "", "ionos_token": "token"}`,
		}, {
			name: "failure, empty Ionos Token",
			cfg:  `{"cluster_id": "12345", "ionos_token": ""}`,
		}, {
			name: "failure, manager creation with wrong pollTimeout",
			cfg:  `{"cluster_id": "12345", "ionos_token": "token", "poll_timeout": "3X"}`,
		}, {
			name: "failure, manager creation with wrong pollTimeout",
			cfg:  `{"cluster_id": "12345", "ionos_token": "token", "poll_interval": "3X"}`,
		}, {
			name: "failure, malformed json",
			cfg:  "{",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CreateIECManager(bytes.NewBufferString(tc.cfg))
			assert.Error(t, err, "error expected")
		})
	}

	t.Run("failure, creating new client", func(t *testing.T) {
		cfg := `{"cluster_id": "12345", "ionos_token": "token"}`
		createClient = func(token, url, authUrl string, timeout, interval time.Duration) (Client, error) {
			return &mocks.Client{}, errors.New("oops, something went wrong")
		}
		_, err := CreateIECManager(bytes.NewBufferString(cfg))
		assert.Error(t, err, "error expected")
		createClient = newClient
	})
}

func TestIECManager_Refresh(t *testing.T) {
	manager := IECManagerImpl{
		clusterID:   "12345",
	}
	t.Run("success, some nodegroups", func(t *testing.T) {
		ionosClient := mocks.Client{}
		manager.ionosClient = client{
			Client:       &ionosClient,
			pollTimeout:  pollTimeout,
			pollInterval: pollInterval,
		}
		ionosClient.On("ListKubernetesNodePools", manager.clusterID).Return(
			&profitbricks.KubernetesNodePools{
				Items: nodePools}, nil).Once()

		err := manager.Refresh()
		assert.NoError(t, err)
		assert.Len(t, manager.nodeGroups, 2,"number of nodegroups do not match")

		// First nodegroup
		assert.Equalf(t, 1, manager.nodeGroups[0].minSize,
			"minimum node size for nodegroup %s does not match", manager.nodeGroups[0].id)
		assert.Equalf(t, 3, manager.nodeGroups[0].maxSize,
			"maximum node size for nodegroup %s does not match", manager.nodeGroups[0].id)

		// Second nodegroup
		assert.Equalf(t, 1, manager.nodeGroups[1].minSize,
			"minimum node size for nodegroup %s does not match", manager.nodeGroups[0].id)
		assert.Equal(t, 2, manager.nodeGroups[1].maxSize,
			"maximum node size for nodegroup %s does not match", manager.nodeGroups[0].id)
		ionosClient.AssertExpectations(t)
	})

	t.Run("success, no nodepolls with autoscaling limits set", func(t *testing.T) {
		ionosClient := mocks.Client{}
		manager.ionosClient = client{
			Client:       &ionosClient,
			pollTimeout:  pollTimeout,
			pollInterval: pollInterval,
		}
		ionosClient.On("ListKubernetesNodePools", manager.clusterID).Return(
			&profitbricks.KubernetesNodePools{
				Items: nodePools[2:4]}, nil).Once()

		err := manager.Refresh()
		assert.NoError(t, err)
		assert.Len(t, manager.nodeGroups, 0,"number of nodegroups do not match")
		ionosClient.AssertExpectations(t)
	})

	t.Run("success, no nodepools", func(t *testing.T) {
		ionosClient := mocks.Client{}
		manager.ionosClient = client{
			Client:       &ionosClient,
			pollTimeout:  pollTimeout,
			pollInterval: pollInterval,
		}
		ionosClient.On("ListKubernetesNodePools", manager.clusterID).Return(
			&profitbricks.KubernetesNodePools{
				Items: []profitbricks.KubernetesNodePool{}}, nil).Once()

		err := manager.Refresh()
		assert.NoError(t, err)
		assert.Len(t, manager.nodeGroups, 0,"number of nodegroups do not match")
		ionosClient.AssertExpectations(t)
	})

	t.Run("failure, error getting nodepools", func(t *testing.T) {
		ionosClient := mocks.Client{}
		manager.ionosClient = client{
			Client:       &ionosClient,
			pollTimeout:  pollTimeout,
			pollInterval: pollInterval,
		}
		ionosClient.On("ListKubernetesNodePools", manager.clusterID).Return(
			nil, errors.New("oops, something went wrong")).Once()

		err := manager.Refresh()
		assert.Error(t, err)
		ionosClient.AssertExpectations(t)
	})
}
