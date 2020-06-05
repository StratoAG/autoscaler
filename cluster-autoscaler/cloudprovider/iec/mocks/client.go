// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import http "net/http"

import mock "github.com/stretchr/testify/mock"
import profitbricks "github.com/profitbricks/profitbricks-sdk-go/v5"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// DeleteKubernetesNode provides a mock function with given fields: clusterID, nodepoolID, nodeID
func (_m *Client) DeleteKubernetesNode(clusterID string, nodepoolID string, nodeID string) (*http.Header, error) {
	ret := _m.Called(clusterID, nodepoolID, nodeID)

	var r0 *http.Header
	if rf, ok := ret.Get(0).(func(string, string, string) *http.Header); ok {
		r0 = rf(clusterID, nodepoolID, nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Header)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(clusterID, nodepoolID, nodeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetKubernetesNode provides a mock function with given fields: clusterID, nodepoolID, nodeID
func (_m *Client) GetKubernetesNode(clusterID string, nodepoolID string, nodeID string) (*profitbricks.KubernetesNode, error) {
	ret := _m.Called(clusterID, nodepoolID, nodeID)

	var r0 *profitbricks.KubernetesNode
	if rf, ok := ret.Get(0).(func(string, string, string) *profitbricks.KubernetesNode); ok {
		r0 = rf(clusterID, nodepoolID, nodeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNode)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(clusterID, nodepoolID, nodeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetKubernetesNodePool provides a mock function with given fields: clusterID, nodepoolID
func (_m *Client) GetKubernetesNodePool(clusterID string, nodepoolID string) (*profitbricks.KubernetesNodePool, error) {
	ret := _m.Called(clusterID, nodepoolID)

	var r0 *profitbricks.KubernetesNodePool
	if rf, ok := ret.Get(0).(func(string, string) *profitbricks.KubernetesNodePool); ok {
		r0 = rf(clusterID, nodepoolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNodePool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(clusterID, nodepoolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListKubernetesNodePools provides a mock function with given fields: clusterID
func (_m *Client) ListKubernetesNodePools(clusterID string) (*profitbricks.KubernetesNodePools, error) {
	ret := _m.Called(clusterID)

	var r0 *profitbricks.KubernetesNodePools
	if rf, ok := ret.Get(0).(func(string) *profitbricks.KubernetesNodePools); ok {
		r0 = rf(clusterID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNodePools)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(clusterID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListKubernetesNodes provides a mock function with given fields: clusterID, nodepoolID
func (_m *Client) ListKubernetesNodes(clusterID string, nodepoolID string) (*profitbricks.KubernetesNodes, error) {
	ret := _m.Called(clusterID, nodepoolID)

	var r0 *profitbricks.KubernetesNodes
	if rf, ok := ret.Get(0).(func(string, string) *profitbricks.KubernetesNodes); ok {
		r0 = rf(clusterID, nodepoolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNodes)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(clusterID, nodepoolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateKubernetesNodePool provides a mock function with given fields: clusterID, nodepoolID, nodepool
func (_m *Client) UpdateKubernetesNodePool(clusterID string, nodepoolID string, nodepool profitbricks.KubernetesNodePool) (*profitbricks.KubernetesNodePool, error) {
	ret := _m.Called(clusterID, nodepoolID, nodepool)

	var r0 *profitbricks.KubernetesNodePool
	if rf, ok := ret.Get(0).(func(string, string, profitbricks.KubernetesNodePool) *profitbricks.KubernetesNodePool); ok {
		r0 = rf(clusterID, nodepoolID, nodepool)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*profitbricks.KubernetesNodePool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, profitbricks.KubernetesNodePool) error); ok {
		r1 = rf(clusterID, nodepoolID, nodepool)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
