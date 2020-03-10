/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iec

import (
	"errors"
	"fmt"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ionosmock "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/iec/mocks/ionos"
)

func testNodeGroup(client client, inp *profitbricks.KubernetesNodePool) NodeGroup {
	var minNodes, maxNodes int
	var id string
	if inp != nil {
		minNodes = 1
		maxNodes = 3
//		minNodes = int(*knp.AutoscalingLimit.Min)
//		maxNodes = int(*knp.AutoscalingLimit.Max)
	}
	if inp != nil {
		id = inp.ID
	}

	return NodeGroup{
		id:          id,
		clusterID:   "12345",
		ionosClient: client,
		nodePool:    inp,
		minSize:     minNodes,
		maxSize:     maxNodes,
	}

}

// TODO: Add autoscaling limits to signature
func initNodeGroup(
	size uint32,
	ionosclient *ionosmock.Client,
	) NodeGroup {
	client := client{
		Client: ionosclient,
		pollInterval: time.Millisecond * 10,
		pollTimeout: time.Millisecond * 200,
	}
	return testNodeGroup(client, initK8sNodePool(size, "123", profitbricks.StateAvailable))
}

func initK8sNodePool(nodecount uint32, id, state string) *profitbricks.KubernetesNodePool{
	np := &profitbricks.KubernetesNodePool{
		Properties: &profitbricks.KubernetesNodePoolProperties{
			NodeCount: nodecount,
		},
		Metadata: &profitbricks.Metadata{
			State: profitbricks.StateAvailable,
		},
		// TODO: Add autoscaling limits to created nodegroup
		//		AutoscalingLimitMin: 1,
		//		AutoscalingLimitMax: 10,
	}
	if id != "" {
		np.ID = id
	}
	return np
}

func TestNodeGroup_Target_size(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := uint32(3)

		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		size, err := ng.TargetSize()
		assert.NoError(t, err)
		assert.EqualValues(t, numberOfNodes, size, "target size is not correct")
	})
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// TODO: Remove as soon as autoscaling limits are implemented
		t.Skip("Missing autoscalingLimits")
		delta := uint32(2)
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updated := initK8sNodePool(newCount, "", profitbricks.StateAvailable)
		client.On("UpdateKubernetesNodePool", ng.clusterID, ng.id, *updated).
			Return(updated, nil).Once()
		// Poll 5 times before it became true
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil).Times(4)
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(updated, nil).Once()

		err := ng.IncreaseSize(int(delta))
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("successfully increase to maximum", func(t *testing.T) {
		// TODO: Remove as soon as autoscaling limits are implemented
		t.Skip("Missing autoscalingLimits")
		delta := uint32(7)
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updated := initK8sNodePool(newCount, "", profitbricks.StateAvailable)

		client.On("UpdateKubernetesNodePool", ng.clusterID, ng.id, *updated).
			Return(updated, nil).Once()
		// Poll 5 times before it became true
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil).Times(4)
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(updated, nil).Once()

		err := ng.IncreaseSize(int(delta))
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("negative increase", func(t *testing.T) {
		delta := -1
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		err := ng.IncreaseSize(delta)
		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("zero increase", func(t *testing.T) {
		delta := 0
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		err := ng.IncreaseSize(delta)
		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("increase above maximum", func(t *testing.T) {
		delta := 8
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		err := ng.IncreaseSize(delta)
		exp := fmt.Errorf("size increase is too large. current: %d, desired: %d, max: %d",
			numberOfNodes, numberOfNodes+uint32(delta), ng.MaxSize())
		assert.EqualError(t, err, exp.Error(), "size increase is too large")
	})

	t.Run("PollNodePoolNodeCount fails", func(t *testing.T) {
		delta := uint32(2)
		numberOfNodes := uint32(1)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updated := initK8sNodePool(newCount, "123", profitbricks.StateAvailable)

		pollError := errors.New("Oops something went wrong")

		client.On("UpdateKubernetesNodePool", ng.clusterID, ng.id, *updated).
			Return(updated, nil).Once()
		// Poll 5 times before it errors
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil).Times(4)
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(nil, pollError).Once()

		err := ng.IncreaseSize(int(delta))
		client.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Errorf("failed to wait for nodepool update: %v", pollError).Error(),
			"Wrong error")
	})

	t.Run("UpdateKubernetesNode times out", func(t *testing.T) {
		delta := uint32(2)
		numberOfNodes := uint32(1)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updated := initK8sNodePool(newCount, "123", profitbricks.StateAvailable)

		pollError := errors.New("timed out waiting for the condition")

		client.On("UpdateKubernetesNodePool", ng.clusterID, ng.id, *updated).
			Return(updated, nil).Once()
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil)

		err := ng.IncreaseSize(int(delta))
		assert.EqualError(t, err, fmt.Errorf("failed to wait for nodepool update: %v", pollError).Error(),
			"Wrong error")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		delta := -1
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := uint32(int(numberOfNodes) + delta)
		updated := initK8sNodePool(newCount, "123", profitbricks.StateAvailable)

		client.On("UpdateKubernetesNodePool", ng.clusterID, ng.id, *updated).
			Return(updated, nil)

		err := ng.DecreaseTargetSize(delta)
		client.AssertExpectations(t)
		assert.NoError(t, err)
	})

	t.Run("successfully decrease to minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := uint32(int(numberOfNodes) + delta)
		updated := initK8sNodePool(newCount, "123", profitbricks.StateAvailable)

		client.On("UpdateKubernetesNodePool", ng.clusterID, ng.id, *updated).
			Return(updated, nil).Once()

		err := ng.DecreaseTargetSize(delta)
		client.AssertExpectations(t)
		assert.NoError(t, err)
	})

	t.Run("positive decrease", func(t *testing.T) {
		delta := 2
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		err := ng.DecreaseTargetSize(delta)
		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("zero decrease", func(t *testing.T) {
		delta := 0
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		err := ng.DecreaseTargetSize(delta)
		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("decrease below minimum", func(t *testing.T) {
		delta := -4
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		err := ng.DecreaseTargetSize(delta)
		exp := fmt.Errorf("size decrease is too small. current: %d, desired: %d, min: %d",
			numberOfNodes, int(numberOfNodes)+delta, ng.MinSize())
		assert.EqualError(t, err, exp.Error(), "size decrease is too small")
	})
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	client := ionosmock.Client{}

	nodes := []*corev1.Node{
		{
			Spec: corev1.NodeSpec{
				ProviderID: "ionos://1",
			},
		},
		{
			Spec: corev1.NodeSpec{
				ProviderID: "ionos://2",
			},
		},
		{
			Spec: corev1.NodeSpec{
				ProviderID: "ionos://3",
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		ng := initNodeGroup(3, &client)
		// Delete first node,
		client.On("DeleteKubernetesNode", ng.clusterID, ng.id, "1").Return(&http.Header{}, nil).Once()
		// Poll 5 times before success
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil).Times(1)
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(initK8sNodePool(ng.nodePool.Properties.NodeCount - 1, "", profitbricks.StateAvailable), nil).Once()
		// Delete second node
		client.On("DeleteKubernetesNode", ng.clusterID, ng.id, "2").Return(&http.Header{}, nil).Once()
		// Poll 5 times before success
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(initK8sNodePool(ng.nodePool.Properties.NodeCount - 1, "", profitbricks.StateAvailable), nil).Times(1)
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(initK8sNodePool(ng.nodePool.Properties.NodeCount - 2, "", profitbricks.StateAvailable), nil).Once()
		err := ng.DeleteNodes(nodes[0:2])
		client.AssertExpectations(t)
		assert.NoError(t, err)
	})

	t.Run("client node deletion fails", func(t *testing.T) {
		ng := initNodeGroup(3, &client)
		// Delete first node
		client.On("DeleteKubernetesNode", ng.clusterID, ng.id, "1").Return(&http.Header{}, nil).Once()
		// Poll 5 times before success
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil).Times(4)
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(initK8sNodePool(ng.nodePool.Properties.NodeCount - 1, "", profitbricks.StateAvailable), nil).Once()
		// Fail on second node
		client.On("DeleteKubernetesNode", ng.clusterID, ng.id, "2").
			Return(&http.Header{}, errors.New("Uups something went wrong.")).Once()

		err := ng.DeleteNodes(nodes)
		assert.EqualError(t, err,
			fmt.Sprintf("deleting node failed for cluster: \"%s\" node pool: \"%s\" node: \"%s\": Uups something went wrong.",
			ng.clusterID, ng.id, "2"))
	})

	t.Run("client node deletion times out", func(t *testing.T) {
		ng := initNodeGroup(3, &client)
		// Delete first node
		client.On("DeleteKubernetesNode", ng.clusterID, ng.id, "1").Return(&http.Header{}, nil).Once()
		// Poll does not success
		client.On("GetKubernetesNodePool", ng.clusterID, ng.nodePool.ID).
			Return(ng.nodePool, nil)

		err := ng.DeleteNodes(nodes)
		assert.EqualError(t, err, "failed to wait for nodepool update: timed out waiting for the condition")
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	numberOfNodes := uint32(3)
	ionosclient := ionosmock.Client{}
	ng := initNodeGroup(numberOfNodes, &ionosclient)

	t.Run("success", func(t *testing.T) {
		ionosclient.On("ListKubernetesNodes", ng.clusterID, ng.id).Return(
			&profitbricks.KubernetesNodes{
				Items: []profitbricks.KubernetesNode{
				{
					ID: "1",
					Metadata: &profitbricks.Metadata{
						State: profitbricks.StateAvailable,
					},
				},
				{
					ID: "2",
					Metadata: &profitbricks.Metadata{
						State: profitbricks.StateBusy,
					},
				},
				{
					ID: "3",
					Metadata: &profitbricks.Metadata{
						State: profitbricks.StateInactive,
					},
				},
				{
					ID: "4",
					Metadata: &profitbricks.Metadata{
						State: profitbricks.StateUnknown,
					},
				},
			},
		}, nil).Once()

		exp := []cloudprovider.Instance{
			{
				Id: "ionos://1",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceRunning,
				},
			}, {
				Id: "ionos://2",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			}, {
				Id: "ionos://3",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			}, {
				Id: "ionos://4",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			},
		}

		nodes, err := ng.Nodes()
		assert.NoError(t, err)
		assert.Equal(t, exp, nodes, "nodes do not match")
	})

	t.Run("failure (nil node pool)", func(t *testing.T) {
		client := client{
			Client: &ionosclient,
			pollInterval: time.Millisecond * 10,
			pollTimeout: time.Millisecond * 100,
		}
		ng := testNodeGroup(client, nil)
		exp := errors.New("node pool instance is not created")

		_, err := ng.Nodes()
		assert.EqualError(t, err, exp.Error(), "Nodes() should return the error: %s", exp)
	})

	t.Run("failure (client GetNodes fails)", func(t *testing.T) {
		e := errors.New("Uups something went wrong")
		ionosclient.On("ListKubernetesNodes", ng.clusterID, ng.id).
			Return(nil, e)
		exp := fmt.Errorf("getting nodes for cluster: %q node pool: %q: %v", ng.clusterID, ng.id, e)

		_, err := ng.Nodes()
		assert.EqualError(t, err, exp.Error(), "Nodes() should return the error: %s", exp)

	})
}

func TestNodeGroup_Debug(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		d := ng.Debug()
		exp := "cluster ID: 12345, nodegroup ID: 123 (min: 1, max: 3)"
		assert.Equal(t, exp, d, "debug string do not match")
	})
}

func TestNodeGroup_Exist(t *testing.T) {
	ionosclient := ionosmock.Client{}

	t.Run("success", func(t *testing.T) {
		numberOfNodes := uint32(3)
		ng := initNodeGroup(numberOfNodes, &ionosclient)
		exist := ng.Exist()
		assert.Equal(t, true, exist, "node pool should exist")
	})

	t.Run("failure", func(t *testing.T) {
		client := client{
			Client: &ionosclient,
			pollInterval: time.Millisecond * 10,
			pollTimeout: time.Millisecond * 100,
		}
		ng := testNodeGroup(client, nil)
		exist := ng.Exist()
		assert.Equal(t, false, exist, "node pool should not exist")
	})
}
