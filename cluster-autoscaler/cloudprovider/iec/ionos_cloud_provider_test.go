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
	"github.com/stretchr/testify/mock"
	"io"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/iec/mocks"
)

var defaultNode =	profitbricks.KubernetesNode{
	ID:         "1",
	Metadata: &profitbricks.Metadata{
		State: profitbricks.StateAvailable,
	},
}

func initializedManager(ionosClient *mocks.Client) *IECManagerImpl {
	client := client{
		Client: ionosClient,
		pollTimeout: pollTimeout,
		pollInterval: pollInterval,
	}
	return &IECManagerImpl{
		ionosClient:         client,
		clusterID:           "12345",
		nodeGroups:          []*NodeGroup{
			{
				id:          "1",
				clusterID:   "12345",
				ionosClient: client,
				nodePool:    &profitbricks.KubernetesNodePool{
					ID:         "1",
				},
				minSize:     1,
				maxSize:     3,
			},
		},
	}
}

func testCloudProvider(t *testing.T, ionosClient *mocks.Client, manager IECManager) *IECCloudProvider {
	if ionosClient == nil {
		ionosClient = &mocks.Client{}
	}
	if manager == nil {
		manager = initializedManager(ionosClient)
	}
	rl := &cloudprovider.ResourceLimiter{}

	provider, err := BuildIECCloudProvider(manager, rl)
	assert.NoError(t, err)
	return provider
}

func TestNewIECCloudProvider(t *testing.T) {
	t.Run("cloud provider creation success", func(t *testing.T) {
		_ = testCloudProvider(t, nil, nil)
	})
}

func TestIECCloudProvider_Name(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)

	t.Run("cloud provider name success", func(t *testing.T) {
		name := provider.Name()
		assert.Equal(t, cloudprovider.IECProviderName, name, "provider name does not match")
	})
}

func TestIECCloudProvider_NodeGroups(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)

	t.Run("#node pool match", func(t *testing.T) {
		nodeGroups := provider.NodeGroups()
		assert.Equal(t, 1, len(nodeGroups), "number of node pools does not match")
		for i, n := range nodeGroups {
			assert.Equalf(t, n, provider.manager.GetNodeGroups()[i], "returned nodegroup %d does not match")
		}
	})
}

func TestIECCloudProvider_NodeGroupForNode(t *testing.T) {
	t.Run("success, no nodegroup found for node", func(t *testing.T) {
		ionosClient := &mocks.Client{}
		ionosClient.On("ListKubernetesNodes", "12345", "1").Return(
			&profitbricks.KubernetesNodes{Items: []profitbricks.KubernetesNode{
				defaultNode,
			}}, nil)
		provider := testCloudProvider(t, ionosClient, nil)
		// try to get nodegroup for node 10
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderId("10"),
			},
		}
		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, nodeGroup)
	})

	t.Run("success", func(t *testing.T) {
		ionosClient := &mocks.Client{}
		ionosClient.On("ListKubernetesNodes", "12345", "1").Return(
			&profitbricks.KubernetesNodes{Items: []profitbricks.KubernetesNode{
				defaultNode,
			}}, nil)
		provider := testCloudProvider(t, ionosClient, nil)
		// try to get the nodeGroup for the node with ID 2
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderId("1"),
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Equalf(t, nodeGroup.Id(), "1", "expected nodegroup id 1, got %d", nodeGroup.Id())
	})

	t.Run("fail, nodeGroupForNode error", func(t *testing.T) {
		// replace method mock in provider
		ionosClient := mocks.Client{}
		ionosClient.On("ListKubernetesNodePools", mock.Anything).Return(
			&profitbricks.KubernetesNodePools{
				Items: []profitbricks.KubernetesNodePool{
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
					}}}, nil)
		ionosClient.On("ListKubernetesNodes", mock.Anything, mock.Anything).Return(
			nil, fmt.Errorf("oops something went wrong"))
		provider := testCloudProvider(t, &ionosClient, nil)

		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderId("node-1-2"),
			},
		}

		nodegroup, err := provider.NodeGroupForNode(node)
		assert.Nil(t, nodegroup)
		assert.Error(t, err)
	})
}

func TestIECCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	gpuTypes := provider.GetAvailableGPUTypes()
	assert.Empty(t, gpuTypes)
}

func TestIECCloudProvider_GetAvailableMachineTypes(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	machineTypes, err := provider.GetAvailableMachineTypes()
	assert.Empty(t, machineTypes)
	assert.NoError(t, err)
}

func TestIECCloudProvider_GetResourceLimiter(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	rl, err := provider.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, provider.resourceLimiter, rl)
}

func TestIECCloudProvider_NewNodeGroup(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	nodeGroup, err := provider.NewNodeGroup(
		"",
		map[string]string{},
		map[string]string{},
		[]apiv1.Taint{},
		map[string]resource.Quantity{})
	assert.Nil(t, nodeGroup)
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())
}

func TestIECCloudProvider_GPULabel(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	label := provider.GPULabel()
	assert.Equal(t, "", label)
}

func TestIECCloudProvider_Pricing(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	_, err := provider.Pricing()
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())
}

func TestIECCloudProvider_Cleanup(t *testing.T) {
	provider := testCloudProvider(t, nil, nil)
	ret := provider.Cleanup()
	assert.Nil(t, ret)
}

type iecManagerMock struct {
	mock.Mock
}

func (m *iecManagerMock) Refresh() error {
	args := m.Called()
	return args.Error(0)
}

func (m *iecManagerMock) GetNodeGroups() []*NodeGroup {
	args := m.Called()
	return args.Get(0).([]*NodeGroup)
}

func (m *iecManagerMock) GetClusterID() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *iecManagerMock) GetPollTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *iecManagerMock) GetPollInterval() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func TestIECCloudProvider_Refresh(t *testing.T) {
	t.Run("succeed", func(t *testing.T) {
		iecManagerMock := iecManagerMock{}
		iecManagerMock.On("Refresh").Return(nil)
		provider := testCloudProvider(t, nil, &iecManagerMock)
		err := 	provider.Refresh()
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, &iecManagerMock)
	})

	t.Run("failure", func(t *testing.T) {
		iecManagerMock := iecManagerMock{}
		iecManagerMock.On("Refresh").Return(errors.New("oops, something went wrong"))
		provider := testCloudProvider(t, nil, &iecManagerMock)
		err := 	provider.Refresh()
		assert.Error(t, err)
		mock.AssertExpectationsForObjects(t, &iecManagerMock)
	})
}

func assertBuildIECPanics(t *testing.T,
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
	panic string) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		} else if !strings.HasPrefix(r.(string), panic) {
				t.Errorf("Wrong panic, got: %s, want: %s", r.(string), panic)
		}
	}()
	BuildIEC(opts, do, rl)
}

func logFatalTest(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func TestBuildIEC(t *testing.T) {
	opts := config.AutoscalingOptions{CloudConfig: "test-file"}
	do := cloudprovider.NodeGroupDiscoveryOptions{}
	rl := &cloudprovider.ResourceLimiter{}

	t.Run("succeed", func(t *testing.T) {
		manager := &IECManagerImpl{
			ionosClient: client{},
			clusterID:   "12345",
			nodeGroups:  nil,
		}
		closeCalled := false
		open = func(name string) (*os.File, error) {
			return nil, nil
		}
		close = func(file io.ReadCloser) {
			closeCalled = true
		}
		createIECManager = func(reader io.Reader) (IECManager, error) {
			return manager, nil
		}
		provider := BuildIEC(opts, do, rl)
		assert.Equal(t, &IECCloudProvider{
			manager:         manager,
			resourceLimiter: rl,
		}, provider)
		assert.True(t, closeCalled)
		open = os.Open
		close = closeFile
		createIECManager = CreateIECManager
	})

	logFatal = logFatalTest

	t.Run("failed to open config file", func(t *testing.T) {
		open = func(name string) (*os.File, error) {
			return nil, errors.New("oops, something went wrong")
		}
		assertBuildIECPanics(t, opts, do, rl, "Couldn't open cloud provider configuration")
		open = os.Open
	})

	t.Run("no cloudconfig provided", func(t *testing.T) {
		assertBuildIECPanics(t, config.AutoscalingOptions{}, do, rl, "Failed to create Ionos Enterprise manager:")
	})

	t.Run("failed to create iec manager", func(t *testing.T) {
		createManagerFromFile = func(config string) (IECManager, error) {
			return nil, errors.New("oops, something went wrong")
		}
		assertBuildIECPanics(t, opts, do, rl, "Failed to create Ionos Enterprise manager:")
		createManagerFromFile = createManager
	})

	t.Run("failed to build iec provider", func(t *testing.T) {
		createManagerFromFile = func(config string) (IECManager, error) {
			return nil, nil
		}
		buildProvider = func(manager IECManager, rl *cloudprovider.ResourceLimiter) (*IECCloudProvider, error) {
			return nil, errors.New("oops, something went wrong")
		}
		assertBuildIECPanics(t, opts, do, rl, "Failed to create IEC cloud provider:")
		createManagerFromFile = createManager
		buildProvider = BuildIECCloudProvider
	})
	logFatal = klog.Fatalf
}


