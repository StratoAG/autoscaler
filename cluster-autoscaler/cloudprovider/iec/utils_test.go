package iec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestUtils_ToProviderID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := "1-2-3-4"
		want := "ionos://1-2-3-4"
		got := toProviderId(in)
		assert.Equal(t, want, got)
	})
}

func TestUtils_ToNodeId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := "ionos://1-2-3-4"
		want := "1-2-3-4"
		got := toNodeId(in)
		assert.Equal(t, want, got)
	})
}

func TestUtils_ToInstances(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := profitbricks.KubernetesNodes{
			Items: []profitbricks.KubernetesNode{
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
			},
		}
		want := []cloudprovider.Instance{
			{
				Id:     "ionos://1",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceRunning,
				},
			},
			{
				Id:     "ionos://2",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			},
			{
				Id:     "ionos://3",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			},
			{
				Id:     "ionos://4",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			},
		}
		got := toInstances(&in)
		assert.Equal(t, want, got)
	})
}

func TestUtils_ToInstance(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := profitbricks.KubernetesNode{
			ID:         "1",
			Metadata:   &profitbricks.Metadata{
				State: profitbricks.StateAvailable,
			},
		}
		want := cloudprovider.Instance{
			Id:     "ionos://1",
			Status: &cloudprovider.InstanceStatus{
				State:     cloudprovider.InstanceRunning,
			},
		}
		got := toInstance(in)
		assert.Equal(t, want, got)
	})
}

func TestUtils_ToInstanceStatus(t *testing.T) {
	tests := []struct {

		in,name string
		want *cloudprovider.InstanceStatus
	}{
		{
			name: "success, ionos server available",
			in: profitbricks.StateAvailable,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		}, {
			name: "success, ionos server busy",
			in: profitbricks.StateBusy,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			name: "success, ionos server unkown",
			in: profitbricks.StateUnknown,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			name: "success, ionos server inactive",
			in: profitbricks.StateInactive,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toInstanceStatus(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("Fail, unknown node state", func(t *testing.T) {
		want := &cloudprovider.InstanceStatus{
			State:     0,
			ErrorInfo: &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    iecErrorCode,
				ErrorMessage: "Unknown node state: wrong_state",
			},
		}
		got := toInstanceStatus("wrong_state")
		assert.Equal(t, want, got)
	})
}