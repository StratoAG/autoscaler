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
	"context"
	"errors"
	"fmt"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ionosmock "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/iec/mocks/ionos"
)

var (
	none  = uint32(0)
	one   = uint32(1)
	two   = uint32(2)
	three = uint32(3)
	ten   = uint32(10)
)

func testNodeGroup(client Client, inp *profitbricks.KubernetesNodePool) NodeGroup {
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

func initNodeGroup(size uint32, client *ionosmock.Client) NodeGroup {
	return testNodeGroup(client, &profitbricks.KubernetesNodePool{
		ID: "1",
		Properties: &profitbricks.KubernetesNodePoolProperties{
			NodeCount: size,
		},
//		AutoscalingLimitMin: 1,
//		AutoscalingLimitMax: 10,
	})
}

func TestNodeGroup_Target_size(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := uint32(3)

		client := ionosmock.Client{}
		ng := testNodeGroup(&client, &profitbricks.KubernetesNodePool{
			ID: "123",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: numberOfNodes,
			},
			// AutoscalingLimitMin: 1
			// AutoscalingLimitMax: 3
		})

		size, err := ng.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, numberOfNodes, uint32(size), "target size is not correct")
	})
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		// TODO: Remove as soon as autoscaling limits are implemented
		t.Skip("Missing autoscalingLimits")
		delta := uint32(2)
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updateInput := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		updatedNodePool := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		client.On("UpdateNodePool", ctx, ng.clusterID, ng.id, &updateInput).
			Return(&updatedNodePool, nil).Once()
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)
		// Poll 5 times before it became true
		//		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID).
		//			Return(wait.ConditionFunc(func() (bool, error) { return false, nil })).Times(4)
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil })).Once()

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
		updateInput := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		updatedNodePool := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}

		client.On("UpdateNodePool", ctx, ng.clusterID, ng.id, &updateInput).
			Return(&updatedNodePool, nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil }))
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)

		err := ng.IncreaseSize(int(delta))
		assert.NoError(t, err)

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
		// TODO: Remove as soon as autoscaling limits are implemented
		delta := uint32(2)
		numberOfNodes := uint32(1)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updateInput := profitbricks.KubernetesNodePool{
			ID: "1",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		updatedNodePool := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}

		pollError := errors.New("Oops something went wrong")

		client.On("UpdateNodePool", ctx, ng.clusterID, ng.id, &updateInput).
			Return(&updatedNodePool, nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID, uint32(3)).
			Return(wait.ConditionFunc(func() (bool, error) { return false, errors.New("Oops something went wrong") }))
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)

		err := ng.IncreaseSize(int(delta))
		assert.EqualError(t, err, fmt.Errorf("failed to wait for nodepool update: %v", pollError).Error(),
			"Wrong error")
	})

	t.Run("PollNodePoolNodeCount never becomes true", func(t *testing.T) {
		// TODO: Remove as soon as autoscaling limits are implemented
		delta := uint32(2)
		numberOfNodes := uint32(1)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := numberOfNodes + delta
		updateInput := profitbricks.KubernetesNodePool{
			ID: "1",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		updatedNodePool := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}

		pollError := errors.New("timed out waiting for the condition")

		client.On("UpdateNodePool", ctx, ng.clusterID, ng.id, &updateInput).
			Return(&updatedNodePool, nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID, uint32(3)).
			Return(wait.ConditionFunc(func() (bool, error) { return false, nil }))
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)

		err := ng.IncreaseSize(int(delta))
		assert.EqualError(t, err, fmt.Errorf("failed to wait for nodepool update: %v", pollError).Error(),
			"Wrong error")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		delta := -1
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := uint32(int(numberOfNodes) + delta)
		updateInput := profitbricks.KubernetesNodePool{
			ID: "1",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		updatedNodePool := profitbricks.KubernetesNodePool{
			ID: "1",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}

		client.On("UpdateNodePool", ctx, ng.clusterID, ng.id, &updateInput).
			Return(&updatedNodePool, nil)
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID, newCount).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil }))
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("successfully decrease to minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := uint32(3)
		client := ionosmock.Client{}
		ng := initNodeGroup(numberOfNodes, &client)

		newCount := uint32(int(numberOfNodes) + delta)
		updateInput := profitbricks.KubernetesNodePool{
			ID: "1",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}
		updatedNodePool := profitbricks.KubernetesNodePool{
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: newCount,
			},
		}

		client.On("UpdateNodePool", ctx, ng.clusterID, ng.id, &updateInput).
			Return(&updatedNodePool, nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, ng.nodePool.ID, uint32(1)).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil }))
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)

		err := ng.DecreaseTargetSize(delta)
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
	ctx := context.TODO()
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
		ng := testNodeGroup(&client, &profitbricks.KubernetesNodePool{
			ID: "123",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: 2,
			},
		})
		np := *ng.nodePool
		// Delete first node,
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "1").Return(nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, np.ID, uint32(1)).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil })).Once()
		// Delete second node
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "2").Return(nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, np.ID, uint32(0)).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil }))
		// Delete third node
		//		client.On("DeleteNode", ctx, ng.clusterID, ng.id, uint32(3)).Return(nil).Once()
		//		newCount = *np.NodeCount - 1
		//		np.NodeCount = &newCount
		//		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, np.ID).
		//			Return(wait.ConditionFunc(func() (bool, error) { return true, nil }))
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)
		err := ng.DeleteNodes(nodes[0:2])
		assert.NoError(t, err)
	})

	t.Run("client node deletion fails", func(t *testing.T) {
		ng := testNodeGroup(&client, &profitbricks.KubernetesNodePool{
			ID: "123",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: 3,
			},
		})
		np := *ng.nodePool
		// Delete first node
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "1").Return(nil).Once()
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, np.ID, uint32(2)).
			Return(wait.ConditionFunc(func() (bool, error) { return true, nil })).Once()
		// Fail on second node
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "2").
			Return(errors.New("Uups something went wrong.")).Once()
		client.On("PollInterval").Return(1 * time.Second)
		client.On("PollTimeout").Return(2 * time.Second)
		client.On("PollNodePoolNodeCount", ctx, ng.clusterID, np.ID, uint32(1)).
			Return(wait.ConditionFunc(func() (bool, error) { return false, nil }))

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	ctx := context.TODO()

	numberOfNodes := uint32(3)
	client := ionosmock.Client{}
	ng := testNodeGroup(&client, &profitbricks.KubernetesNodePool{
		ID: "123",
		Properties: &profitbricks.KubernetesNodePoolProperties{
			NodeCount: numberOfNodes,
		},
		// AutoscalingLimitMin: 1,
		// AutoscalingLimitMax: 10,
	})

	t.Run("success", func(t *testing.T) {
		client.On("GetNodes", ctx, ng.clusterID, ng.id).Return([]profitbricks.KubernetesNode{
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
		client := ionosmock.Client{}
		ng := testNodeGroup(&client, nil)
		exp := errors.New("node pool instance is not created")

		_, err := ng.Nodes()
		assert.EqualError(t, err, exp.Error(), "Nodes() should return the error: %s", exp)
	})

	t.Run("failure (client GetNodes fails)", func(t *testing.T) {
		e := errors.New("Uups something went wrong")
		client.On("GetNodes", ctx, ng.clusterID, ng.id).
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
		ng := testNodeGroup(&client, &profitbricks.KubernetesNodePool{
			ID: "123",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: numberOfNodes,
			},
			// AutoscalingLimitMin: 1,
			// AutoscalingLimitMax: 3
		})

		d := ng.Debug()
		exp := "cluster ID: 12345, nodegroup ID: 123 (min: 1, max: 3)"
		assert.Equal(t, exp, d, "debug string do not match")
	})
}

func TestNodeGroup_Exist(t *testing.T) {
	client := ionosmock.Client{}

	t.Run("success", func(t *testing.T) {
		numberOfNodes := uint32(3)
		ng := testNodeGroup(&client, &profitbricks.KubernetesNodePool{
			ID: "123",
			Properties: &profitbricks.KubernetesNodePoolProperties{
				NodeCount: numberOfNodes,
			},
			// AutoscalingLimitMin: 1,
			// AutoscalingLimitMax: 10,
		})
		exist := ng.Exist()
		assert.Equal(t, true, exist, "node pool should exist")
	})

	t.Run("failure", func(t *testing.T) {
		ng := testNodeGroup(&client, nil)
		exist := ng.Exist()
		assert.Equal(t, false, exist, "node pool should not exist")
	})
}
