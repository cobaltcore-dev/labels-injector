/*
SPDX-FileCopyrightText: Copyright 2025 SAP SE or an SAP affiliate company and cobaltcore-dev contributors
SPDX-License-Identifier: Apache-2.0

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
