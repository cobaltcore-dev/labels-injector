package internal

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TransferLabel(ctx context.Context, pod *v1.Pod, node *v1.Node, d client.Client) error {
	var transferLabels = []string{
		"kubernetes.metal.cloud.sap/name",
		"kubernetes.metal.cloud.sap/cluster",
		"kubernetes.metal.cloud.sap/bb",
		"topology.kubernetes.io/region",
		"topology.kubernetes.io/zone",
	}

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	patch := client.MergeFrom(pod.DeepCopy())
	// transfer the labels from the node to the pod
	for _, label := range transferLabels {
		if nodeLabel, ok := node.Labels[label]; ok {
			pod.Labels[label] = nodeLabel
		}
	}
	if _, ok := node.Labels["kubernetes.metal.cloud.sap/name"]; !ok {
		// legacy cluster fallback strategy
		pod.Labels["kubernetes.metal.cloud.sap/name"] = node.Name
	}

	// patch the pod object
	if err := d.Patch(ctx, pod, patch); err != nil {
		return fmt.Errorf("patch Pod %s/%s: %w", pod.Namespace, pod.Name, err)
	}

	return nil
}
