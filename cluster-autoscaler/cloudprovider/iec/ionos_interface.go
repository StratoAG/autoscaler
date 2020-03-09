package iec

import (
	"context"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

//go:generate mockery -name Client -case snake -dir . -output ./cloudprovider/iec/mocks/iec
type Client interface {

	// GetNodePool gets a single node pool from the iec API
	GetNodePool(ctx context.Context, clusterID, nodepoolID string) (*profitbricks.KubernetesNodePool, error)

	// ListNodePools lists all the node pools in a kubernetes cluster.
	ListNodePools(ctx context.Context, clusterID string) ([]profitbricks.KubernetesNodePool, error)

	// UpdateNodePool updates a specific node pool in a kubernetes cluster.
	UpdateNodePool(ctx context.Context, clusterID, nodepoolID string, nodepool *profitbricks.KubernetesNodePool) (*profitbricks.KubernetesNodePool, error)

	// PollNodePoolNodeCount returns a condition function that checks if the NodeCount is updated
	PollNodePoolNodeCount(ctx context.Context, clusterID, nodepoolID string, targetSize uint32) wait.ConditionFunc

	// GetNode gets a single node for given clusterID, nodepoolID, nodeID
	GetNode(ctx context.Context, clusterID, nodepoolID, nodeID string) (*profitbricks.KubernetesNode, error)

	// GetNodes gets all nodes for given clusterID, nodepoolID
	GetNodes(ctx context.Context, clusterID, nodepoolID string) ([]profitbricks.KubernetesNode, error)

	// DeleteNode deletes a node by clusterID, nodepoolID, nodeID
	DeleteNode(ctx context.Context, clusterID, nodepoolID, nodeID string) error

	// PollTimeout returns the pollTimeout
	PollTimeout() time.Duration

	// PollInterval returns the pollInterval
	PollInterval() time.Duration
}
