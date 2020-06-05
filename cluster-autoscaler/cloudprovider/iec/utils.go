package iec

import (
	"fmt"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"github.com/profitbricks/profitbricks-sdk-go/v5"
)

const (
	iecProviderIDPrefix = "ionos://"
	iecErrorCode        = "no-code-iec"
)

// toProviderId converts plain node id to a node.spec.ProviderId
func toProviderId(nodeID string) string {
	return fmt.Sprintf("%s%s", iecProviderIDPrefix, nodeID)
}

// toNodeId converts a node.spec.ProviderId to plain node id
func toNodeId(providerID string) string {
	return strings.TrimPrefix(providerID, iecProviderIDPrefix)
}

// toInstances converts a slice of *korev2.NodeResult to an array of cloudprovider.Instance
func toInstances(nodes *profitbricks.KubernetesNodes) []cloudprovider.Instance {
	instances := make([]cloudprovider.Instance, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		instances = append(instances, toInstance(node))
	}
	return instances
}

// to Instance converts a given *korev2.NodeResult to a cloudprovider.Instance
func toInstance(node profitbricks.KubernetesNode) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderId(fmt.Sprint(node.ID)),
		Status: toInstanceStatus(node.Metadata.State),
	}
}

// toInstanceStatus converts the given korev2.NodeLifecycleState to a cloudprovider.InstanceStatus
func toInstanceStatus(nodeState string) *cloudprovider.InstanceStatus {
	st := &cloudprovider.InstanceStatus{}
	switch nodeState {
	case profitbricks.StateUnknown:
		fallthrough
	case profitbricks.StateInactive:
		fallthrough
	case profitbricks.StateBusy:
		st.State = cloudprovider.InstanceCreating
	case profitbricks.StateAvailable:
		st.State = cloudprovider.InstanceRunning
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    iecErrorCode,
			ErrorMessage: fmt.Sprintf("Unknown node state: %s", nodeState),
		}
	}
	return st
}
