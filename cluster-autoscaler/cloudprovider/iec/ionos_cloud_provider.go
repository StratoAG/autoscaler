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
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog"
)

var _ cloudprovider.CloudProvider = &IECCloudProvider{}

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = ""
)

// IECCloudProvider implements CloudProvider interface.
type IECCloudProvider struct {
	manager         IECManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func BuildIECCloudProvider(manager IECManager, rl *cloudprovider.ResourceLimiter) (*IECCloudProvider, error) {
	return &IECCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (d *IECCloudProvider) Name() string {
	return cloudprovider.IECProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *IECCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(d.manager.GetNodeGroups()))
	for i, ng := range d.manager.GetNodeGroups() {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (d *IECCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	providerID := node.Spec.ProviderID
	klog.V(5).Infof("checking nodegroup for node ID: %q", providerID)

	// NOTE(arslan): the number of node groups per cluster is usually very
	// small. So even though this looks like quadratic runtime, it's OK to
	// proceed with this.
	for _, group := range d.manager.GetNodeGroups() {
		klog.V(5).Infof("iterating over node group %q", group.Id())
		nodes, err := group.Nodes()
		if err != nil {
			return nil, err
		}

		for _, n := range nodes {
			klog.V(6).Infof("checking node have: %q want: %q", n.Id, providerID)
			if n.Id != providerID {
				continue
			}

			return group, nil
		}
	}

	// there is no "ErrNotExist" error, so we have to return a nil error
	return nil, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (d *IECCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (d *IECCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (d *IECCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (d *IECCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (d *IECCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (d *IECCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (d *IECCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (d *IECCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group cache")
	return d.manager.Refresh()
}

// For testing purposes
var (
	buildProvider         = BuildIECCloudProvider
	close                 = closeFile
	createManagerFromFile = createManager
	createIECManager      = CreateIECManager
	logFatal              = klog.Fatalf
	open				  = os.Open
)

func closeFile(file io.ReadCloser) {
	file.Close()
}

func createManager(config string) (IECManager, error) {
	var configFile io.ReadCloser
	if config != "" {
		var err error
		configFile, err = open(config)
		if err != nil {
			logFatal("Couldn't open cloud provider configuration %s: %#v", config, err)
		}
		defer close(configFile)
	}
	return createIECManager(configFile)
}

// BuildIEC builds the Ionos Enterprise cloud provider.
func BuildIEC(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	manager, err := createManagerFromFile(opts.CloudConfig)
	if err != nil {
		logFatal("Failed to create Ionos Enterprise manager: %v", err)
	}

	// the cloud provider automatically uses all node pools in Ionos Enterprise.
	// This means we don't use the cloudprovider.NodeGroupDiscoveryOptions
	// flags (which can be set via '--node-group-auto-discovery' or '-nodes')
	provider, err := buildProvider(manager, rl)
	if err != nil {
		logFatal("Failed to create IEC cloud provider: %v", err)
	}

	return provider
}
