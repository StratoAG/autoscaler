package iec

import (
	"errors"
	"fmt"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	providerIDPrefix      = "ionos://"
	koreHostNameNamespace = "kubernetes.io"
	nodeHostName          = koreHostNameNamespace + "/hostname"
)

type NodeGroup struct {
	id          string
	clusterID   string
	ionosClient client
	nodePool    *profitbricks.KubernetesNodePool

	minSize int
	maxSize int
}

// Maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// Minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// Target size of the node group. Might be different to the actual size, but
// should stabilize.
func (n *NodeGroup) TargetSize() (int, error) {
	return int(n.nodePool.Properties.NodeCount), nil
}

// Increases node group size
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.nodePool.Properties.NodeCount + uint32(delta)

	if targetSize > uint32(n.MaxSize()) {
		return fmt.Errorf("size increase is too large. current: %d, desired: %d, max: %d",
			n.nodePool.Properties.NodeCount, targetSize, n.MaxSize())
	}

	upgradeInput := *n.nodePool
	upgradeInputProperties := *upgradeInput.Properties
	upgradeInputProperties.NodeCount = targetSize
	upgradeInput.Properties = &upgradeInputProperties

	_, err := n.ionosClient.UpdateKubernetesNodePool(n.clusterID, n.nodePool.ID, upgradeInput)
	if err != nil {
		return err
	}
	err = wait.PollImmediate(
		n.ionosClient.PollInterval(),
		n.ionosClient.PollTimeout(),
		n.ionosClient.PollNodePoolNodeCount(n.clusterID, n.nodePool.ID, targetSize))
	if err != nil {
		return fmt.Errorf("failed to wait for nodepool update: %v", err)
	}

	n.nodePool, err = n.ionosClient.GetKubernetesNodePool(n.clusterID, n.nodePool.ID)
	if err != nil {
		return fmt.Errorf("failed to get nodepool")
	}
	return nil
}

// DeleteNodes deletes nodes from this node group (and also decreasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(kubernetesNodes []*apiv1.Node) error {

	for _, node := range kubernetesNodes {
		// Use node.Spec.ProviderID as to retrieve nodeID
		nodeID := strings.TrimPrefix(node.Spec.ProviderID, providerIDPrefix)

		_, err := n.ionosClient.DeleteKubernetesNode(n.clusterID, n.id, nodeID)
		if err != nil {
			return fmt.Errorf("deleting node failed for cluster: %q node pool: %q node: %q: %v",
				n.clusterID, n.id, nodeID, err)
		}
		targetSize := n.nodePool.Properties.NodeCount - 1
		err = wait.PollImmediate(
			n.ionosClient.PollInterval(),
			n.ionosClient.PollTimeout(),
			n.ionosClient.PollNodePoolNodeCount(n.clusterID, n.nodePool.ID, targetSize))
		if err != nil {
			return fmt.Errorf("failed to wait for nodepool update: %v", err)
		}
		n.nodePool.Properties.NodeCount = targetSize
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	targetSize := int(n.nodePool.Properties.NodeCount) + delta
	if targetSize < 0 || targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d, desired: %d, min: %d",
			n.nodePool.Properties.NodeCount, targetSize, n.MinSize())
	}

	upgradeInput := *n.nodePool
	upgradeInputProperties := *upgradeInput.Properties
	upgradeInputProperties.NodeCount = uint32(targetSize)
	upgradeInput.Properties = &upgradeInputProperties

	updatedNodePool, err := n.ionosClient.UpdateKubernetesNodePool(n.clusterID, n.nodePool.ID, upgradeInput)
	if err != nil {
		return err
	}

	if updatedNodePool.Properties.NodeCount != uint32(targetSize) {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.Properties.NodeCount)
	}

	// update internal cache
	n.nodePool.Properties.NodeCount = uint32(targetSize)
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return fmt.Sprint(n.id)
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("cluster ID: %s, nodegroup ID: %s (min: %d, max: %d)", n.clusterID, n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.  It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	if n.nodePool == nil {
		return nil, errors.New("node pool instance is not created")
	}
	nodes, err := n.ionosClient.ListKubernetesNodes(n.clusterID, n.id)
	if err != nil {
		return nil, fmt.Errorf("getting nodes for cluster: %q node pool: %q: %v",
			n.clusterID, n.id, err)
	}

	return toInstances(nodes.Items), nil
}

// TemplateNodeInfo returns a schedulernodeinfo.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return n.nodePool != nil
}

// Create creates the node group on the cloud provider side. Implementation
// optional. Not implemented for iec enterprise cloud.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.  This will be
// executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional. Not implemented for iec enterprise cloud.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled to 0.
// Atuoprovisioned groups are curretly not supported for iec enterprise cloud.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}
