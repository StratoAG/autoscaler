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
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ionosmock "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/iec/mocks/ionos"
)

func testCloudProvider(t *testing.T, ionosClient *ionosmock.Client) *IECloudProvider {
	cfg := `{"ionos_token": "secret-iec-token", "cluster_id": "123456"}`
	manager := testManager(t, cfg, nil)
	rl := &cloudprovider.ResourceLimiter{}

	provider, err := BuildIECloudProvider(manager, rl)
	assert.NoError(t, err)
	return provider
}

func TestNewIonosEnterpriseCloudProvider(t *testing.T) {
	t.Run("cloud provider creation success", func(t *testing.T) {
		_ = testCloudProvider(t, nil)
	})
}

func TestIonosEnterpriseCloudProvider_Name(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("cloud provider name success", func(t *testing.T) {
		name := provider.Name()
		assert.Equal(t, cloudprovider.IECProviderName, name, "provider name does not match")
	})
}

func TestIonosEnterpriseCloudProvider_NodeGroups(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("#node pool match", func(t *testing.T) {
		t.Skip("Missing autoscalingLimits")
		nodepools := provider.NodeGroups()
		assert.Equal(t, len(nodepools), 2, "number of node pools does not match")
	})

	t.Run("zero groups", func(t *testing.T) {
		provider.manager.nodeGroups = []*NodeGroup{}
		nodes := provider.NodeGroups()
		assert.Equal(t, len(nodes), 0, "number of node pools does not match")
	})
}

func TestIonosEnterpriseCloudProvider_NodeGroupForNode(t *testing.T) {
	provider := testCloudProvider(t, nil)
	t.Run("success", func(t *testing.T) {

		// lets get the nodeGroup for the node with ID 2
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderId("node-1-2"),
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, nodeGroup)
	})
}
