package iec

import (
	"context"
	"time"

	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	KubernetesNodePoolStateAVAILABLE = "AVAILABLE"
)

type client struct {
	*profitbricks.Client
	pollTimeout  time.Duration
	pollInterval time.Duration
}

var _ Client = &client{}

func newClient(token, url, authUrl string, timeout, interval time.Duration) (Client, error) {
	ionosClient := profitbricks.NewClientbyToken(token)
	if url != "" {
		ionosClient.SetCloudApiURL(url)
	}
	if authUrl != "" {
		ionosClient.SetAuthApiUrl(authUrl)
	}

	return &client{
		Client:       ionosClient,
		pollTimeout:  timeout,
		pollInterval: interval,
	}, nil
}

func (i *client) GetNodePool(ctx context.Context, clusterID, nodepoolID string) (*profitbricks.KubernetesNodePool, error) {
	resp, err := i.GetKubernetesNodePool(clusterID, nodepoolID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *client) ListNodePools(ctx context.Context, clusterID string) ([]profitbricks.KubernetesNodePool, error) {
	resp, err := i.ListKubernetesNodePools(clusterID)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (i *client) UpdateNodePool(ctx context.Context, clusterID, nodepoolID string, np *profitbricks.KubernetesNodePool) (*profitbricks.KubernetesNodePool, error) {
	resp, err := i.UpdateKubernetesNodePool(clusterID, nodepoolID, *np)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *client) PollNodePoolNodeCount(ctx context.Context, clusterID, nodepoolID string, targetSize uint32) wait.ConditionFunc {
	return func() (bool, error) {
		np, err := i.GetNodePool(ctx, clusterID, nodepoolID)
		if err != nil {
			return false, err
		}
		if np.Metadata.State == profitbricks.StateAvailable && np.Properties.NodeCount == targetSize {
			return true, nil
		}
		return false, nil
	}
}

func (i *client) GetNode(ctx context.Context, clusterID, nodepoolID, nodeID string) (*profitbricks.KubernetesNode, error) {
	resp, err := i.GetKubernetesNode(clusterID, nodepoolID, nodeID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *client) GetNodes(ctx context.Context, clusterID, nodepoolID string) ([]profitbricks.KubernetesNode, error) {
	resp, err := i.ListKubernetesNodes(clusterID, nodepoolID)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (i *client) DeleteNode(ctx context.Context, clusterID, nodepoolID, nodeID string) error {
	_, err := i.DeleteKubernetesNode(clusterID, nodepoolID, nodeID)
	if err != nil {
		return err
	}
	return nil
}

func (i *client) PollTimeout() time.Duration {
	return i.pollTimeout
}

func (i *client) PollInterval() time.Duration {
	return i.pollInterval
}
