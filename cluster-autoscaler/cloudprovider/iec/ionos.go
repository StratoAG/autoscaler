package iec

import (
	"time"

	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"k8s.io/apimachinery/pkg/util/wait"
)

type client struct {
	Client
	pollTimeout  time.Duration
	pollInterval time.Duration
}

var _ Client = &profitbricks.Client{}

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

func (i *client) PollNodePoolNodeCount(clusterID, nodepoolID string, targetSize uint32) wait.ConditionFunc {
	return func() (bool, error) {
		np, err := i.GetKubernetesNodePool(clusterID, nodepoolID)
		if err != nil {
			return false, err
		}
		if np.Metadata.State == profitbricks.StateAvailable && np.Properties.NodeCount == targetSize {
			return true, nil
		}
		return false, nil
	}
}

func (i *client) PollTimeout() time.Duration {
	return i.pollTimeout
}

func (i *client) PollInterval() time.Duration {
	return i.pollInterval
}
