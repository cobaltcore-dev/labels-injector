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

package v1

import (
	"context"
	"net/http"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/cobaltcore-dev/labels-injector/internal"
)

// SetupLabelsInjectorWithManager registers the webhook for Pod/bindings in the manager.
func SetupLabelsInjectorWithManager(mgr ctrl.Manager) error {
	h := &PodLabelTransferHandler{
		admission.NewDecoder(mgr.GetScheme()),
		mgr.GetClient(),
	}
	mgr.GetWebhookServer().Register("/admission--v1-pods-binding", &webhook.Admission{Handler: h})
	return nil
}

// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;patch
// +kubebuilder:webhook:path=/admission--v1-pods-binding,mutating=false,failurePolicy=ignore,sideEffects=None,groups="",resources=pods/binding,verbs=create;update,versions=v1,name=labels-injector.kvm.cloud.sap,admissionReviewVersions=v1

// PodLabelTransferHandler struct is responsible for transferring labels from the node to the pod.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PodLabelTransferHandler struct {
	admission.Decoder
	client.Client
}

func (d *PodLabelTransferHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
	log := logger.FromContext(ctx).WithName("label-injector")

	log.Info("Handling request", "reqest", request.Name)

	// check if the request is for a binding
	if request.Kind.Kind != "Binding" {
		log.Error(nil, "expected a binding request", "kind", request.Kind.Kind)
		return admission.Allowed("Not a binding request")
	}

	binding := &v1.Binding{}
	if err := d.Decode(request, binding); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if binding.Target.Kind != "Node" {
		// ignore binding for non-node targets
		log.Info("ignoring, target is not a node", "kind", binding.Target.Kind)
		return admission.Allowed("Binding is not for a node")
	}

	nodeName := binding.Target.Name
	if nodeName == "" {
		log.Info("ignoring, node name is empty", "pod", binding.Name)
		return admission.Allowed("Node name is empty")
	}

	// fetch node object
	node := &v1.Node{}
	if err := d.Get(ctx, client.ObjectKey{Name: nodeName}, node); client.IgnoreNotFound(err) != nil {
		log.Error(err, "failed to fetch node", "name", nodeName)
		return admission.Allowed("").WithWarnings("Failed to fetch node")
	}

	// fetch the pod object
	pod := &v1.Pod{}
	if err := d.Get(ctx, client.ObjectKeyFromObject(binding), pod); client.IgnoreNotFound(err) != nil {
		log.Error(err, "failed to fetch pod", "object", client.ObjectKeyFromObject(binding))
		return admission.Allowed("").WithWarnings("Failed to fetch pod")
	}

	if err := internal.TransferLabel(ctx, pod, node, d.Client); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to patch pod", "object", client.ObjectKeyFromObject(pod))
	}

	return admission.Allowed("")
}
