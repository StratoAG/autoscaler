// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"

import mock "github.com/stretchr/testify/mock"
import profitbricks "github.com/profitbricks/profitbricks-sdk-go/v5"
import time "time"
import wait "k8s.io/apimachinery/pkg/util/wait"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// DeleteNode provides a mock function with given fields: ctx, clusterID, nodepoolID, nodeID
func (_m *Client) DeleteNode(ctx context.Context, clusterID string, nodepoolID string, nodeID string) error {
	ret := _m.Called(ctx, clusterID, nodepoolID, nodeID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, clusterID, nodepoolID, nodeID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetNode provides a mock function with given fields: ctx, clusterID, nodepoolID, nodeID
func (_m *Client) GetNode(ctx context.Context, clusterID string, nodepoolID string, nodeID string) (*profitbricks.KubernetesNode, error) {
	ret := _m.Called(ctx, clusterID, nodepoolID, nodeID)

	var r0 *profitbricks.KubernetesNode
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *profitbricks.KubernetesNode); ok {
		r0 = rf(ctx, clusterID, nodepoolID, nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNode)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, clusterID, nodepoolID, nodeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNodePool provides a mock function with given fields: ctx, clusterID, nodepoolID
func (_m *Client) GetNodePool(ctx context.Context, clusterID string, nodepoolID string) (*profitbricks.KubernetesNodePool, error) {
	ret := _m.Called(ctx, clusterID, nodepoolID)

	var r0 *profitbricks.KubernetesNodePool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *profitbricks.KubernetesNodePool); ok {
		r0 = rf(ctx, clusterID, nodepoolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNodePool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, clusterID, nodepoolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNodes provides a mock function with given fields: ctx, clusterID, nodepoolID
func (_m *Client) GetNodes(ctx context.Context, clusterID string, nodepoolID string) ([]profitbricks.KubernetesNode, error) {
	ret := _m.Called(ctx, clusterID, nodepoolID)

	var r0 []profitbricks.KubernetesNode
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []profitbricks.KubernetesNode); ok {
		r0 = rf(ctx, clusterID, nodepoolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]profitbricks.KubernetesNode)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, clusterID, nodepoolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListNodePools provides a mock function with given fields: ctx, clusterID
func (_m *Client) ListNodePools(ctx context.Context, clusterID string) ([]profitbricks.KubernetesNodePool, error) {
	ret := _m.Called(ctx, clusterID)

	var r0 []profitbricks.KubernetesNodePool
	if rf, ok := ret.Get(0).(func(context.Context, string) []profitbricks.KubernetesNodePool); ok {
		r0 = rf(ctx, clusterID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]profitbricks.KubernetesNodePool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, clusterID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PollInterval provides a mock function with given fields:
func (_m *Client) PollInterval() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// PollNodePoolNodeCount provides a mock function with given fields: ctx, clusterID, nodepoolID, targetSize
func (_m *Client) PollNodePoolNodeCount(ctx context.Context, clusterID string, nodepoolID string, targetSize uint32) wait.ConditionFunc {
	ret := _m.Called(ctx, clusterID, nodepoolID, targetSize)

	var r0 wait.ConditionFunc
	if rf, ok := ret.Get(0).(func(context.Context, string, string, uint32) wait.ConditionFunc); ok {
		r0 = rf(ctx, clusterID, nodepoolID, targetSize)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(wait.ConditionFunc)
		}
	}

	return r0
}

// PollTimeout provides a mock function with given fields:
func (_m *Client) PollTimeout() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// UpdateNodePool provides a mock function with given fields: ctx, clusterID, nodepoolID, nodepool
func (_m *Client) UpdateNodePool(ctx context.Context, clusterID string, nodepoolID string, nodepool *profitbricks.KubernetesNodePool) (*profitbricks.KubernetesNodePool, error) {
	ret := _m.Called(ctx, clusterID, nodepoolID, nodepool)

	var r0 *profitbricks.KubernetesNodePool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *profitbricks.KubernetesNodePool) *profitbricks.KubernetesNodePool); ok {
		r0 = rf(ctx, clusterID, nodepoolID, nodepool)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNodePool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *profitbricks.KubernetesNodePool) error); ok {
		r1 = rf(ctx, clusterID, nodepoolID, nodepool)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
