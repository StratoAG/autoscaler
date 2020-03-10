package iec

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ionosmock "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/iec/mocks/ionos"
)

func testManager(t *testing.T, cfg string, ionosClient *ionosmock.Client) *Manager {
	manager, err := CreateIECManager(bytes.NewBufferString(cfg))
	assert.NoError(t, err)
	if ionosClient == nil {
		ionosClient = &ionosmock.Client{}
		ionosClient.On("ListKubernetesNodePools", manager.clusterID).Return(
			&profitbricks.KubernetesNodePools{
				Items: []profitbricks.KubernetesNodePool{
					{
						ID:         "1",
						Properties: &profitbricks.KubernetesNodePoolProperties{
							Name:             "nodepool-1",
							NodeCount:        2,
							//AutoscalingLimitMin: 1
							//AutoscalingLimitMax: 3
						},
					},
					{
						ID:         "2",
						Properties: &profitbricks.KubernetesNodePoolProperties{
							Name:             "nodepool-2",
							NodeCount:        2,
							//AutoscalingLimitMin: 1
							//AutoscalingLimitMax: 2
						},
					},
					{
						ID:         "3",
						Properties: &profitbricks.KubernetesNodePoolProperties{
							Name:             "nodepool-3",
							NodeCount:        2,
							//AutoscalingLimitMin: 0
							//AutoscalingLimitMax: 0
						},
					},
					{
						ID:         "4",
						Properties: &profitbricks.KubernetesNodePoolProperties{
							Name:             "nodepool-4",
							NodeCount:        2,
						},
					},
				},
			}, nil)
		ionosClient.On("ListKubernetesNodes", manager.clusterID, "1").Return(
			&profitbricks.KubernetesNodes{
				Items: []profitbricks.KubernetesNode{
					{
						ID: "1",
						Metadata: &profitbricks.Metadata{
							State: profitbricks.StateAvailable,
						},
						Properties: &profitbricks.KubernetesNodeProperties{
							Name: "node-1-1",
						},
					},
					{
						ID: "2",
						Metadata: &profitbricks.Metadata{
							State: profitbricks.StateAvailable,
						},
						Properties: &profitbricks.KubernetesNodeProperties{
							Name: "node-1-2",
						},
					},
				},
			}, nil).Once()
		ionosClient.On("ListKubernetesNodes", manager.clusterID, "2").Return(
			&profitbricks.KubernetesNodes{
				Items: []profitbricks.KubernetesNode{
					{
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
					},
				},
			}, nil).Once()
		ionosClient.On("ListKubernetesNodes", manager.clusterID, "3").Return(
			&profitbricks.KubernetesNodes{
				Items:[]profitbricks.KubernetesNode{
					{
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
					},
				},
			}, nil).Once()
		ionosClient.On("ListKubernetesNodes", manager.clusterID, "4").Return(
			&profitbricks.KubernetesNodes{
				Items: []profitbricks.KubernetesNode{
					{
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
				},
			}, nil).Once()
		ionosClient.On("GetKubernetesNodePool", manager.clusterID, "1").Return(
			&profitbricks.KubernetesNodePool{
				ID:         "1",
				Properties: &profitbricks.KubernetesNodePoolProperties{
					Name: "nodepool-1",
					NodeCount: 2,
				},
			}, nil)
		ionosClient.On("GetKubernetesNodePool", manager.clusterID, "2").Return(
			&profitbricks.KubernetesNodePool{
				ID:         "2",
				Properties: &profitbricks.KubernetesNodePoolProperties{
					Name: "nodepool-2",
					NodeCount: 2,
				},
			}, nil)
		ionosClient.On("GetKubernetesNodePool", manager.clusterID, "3").Return(
			// This nodepool has autoscalingLimit.Max set to 0,
			// therefore autoscaling is disabled and it should not show up
			&profitbricks.KubernetesNodePool{
				ID:         "3",
				Properties: &profitbricks.KubernetesNodePoolProperties{
					Name: "nodepool-3",
					NodeCount: 2,
				},
			}, nil)
		ionosClient.On("GetKubernetesNodePool", manager.clusterID, "4").Return(
			// This nodepool does not provide any AutoscalingLimits,
			// therefore autoscaling is disabled and it should not show up.
			&profitbricks.KubernetesNodePool{
				ID:         "4",
				Properties: &profitbricks.KubernetesNodePoolProperties{
					Name: "nodepool-4",
					NodeCount: 2,
				},
			}, nil)
	}
	manager.ionosClient = client{
		Client: ionosClient,
		pollTimeout: time.Millisecond * 10,
		pollInterval: time.Millisecond * 10,
	}
	manager.Refresh()
	return manager
}

func TestNewManager(t *testing.T) {
	t.Run("Manager creation success", func(t *testing.T) {
		cfg := `{"cluster_id": "12345", "ionos_token": "secret_ionos_token"}`
		manager, err := CreateIECManager(bytes.NewBufferString(cfg))
		assert.NoError(t, err)
		assert.Equal(t, "12345", manager.clusterID, "cluster ID does not match")
	})

	tests := []struct {
		name,
		cluster_id,
		ionos_token string
		err error
	}{
		{
			name:          "failure, empty cluster ID",
			cluster_id:    "",
			ionos_token:   "secret_ionos_token",
			err:           errors.New("cluster ID is not provided"),
		}, {
			name:          "failure, empty Ionos Token",
			cluster_id:    "12345",
			ionos_token:   "",
			err:           errors.New("iec access token is not provided"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := fmt.Sprintf(`{"cluster_id": "%s", "ionos_token": "%s"}`,
				tc.cluster_id, tc.ionos_token)
			_, err := CreateIECManager(bytes.NewBufferString(cfg))
			assert.EqualErrorf(t, err, tc.err.Error(), "error is expected: %s", tc.err)
		})
	}
}

func TestIECManager_Refresh(t *testing.T) {
	t.Skip("Missing autoscalingLimits")
	cfg := `{"cluster_id": "12345", "ionos_token": "secret_ionos_token"}`
	manager := testManager(t, cfg, nil)

	err := manager.Refresh()
	assert.NoError(t, err)
	assert.Len(t, manager.nodeGroups, 2,"number of nodegroups do not match")

	// Second nodegroup
	assert.Equalf(t, 1, manager.nodeGroups[0].minSize,
		"minimum node size for nodegroup %s does not match", manager.nodeGroups[0].id)
	assert.Equalf(t, 3, manager.nodeGroups[0].maxSize,
		"maximum node size for nodegroup %s does not match", manager.nodeGroups[0].id)

	// First nodegroup
	assert.Equalf(t, 1, manager.nodeGroups[1].minSize,
		"minimum node size for nodegroup %s does not match", manager.nodeGroups[0].id)
	assert.Equal(t, 2, manager.nodeGroups[1].maxSize,
		"maximum node size for nodegroup %s does not match", manager.nodeGroups[0].id)
}
