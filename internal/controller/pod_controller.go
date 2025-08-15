/*
SPDX-FileCopyrightText: Copyright 2024 SAP SE or an SAP affiliate company and cobaltcore-dev contributors
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, LibVirtVersion 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/cobaltcore-dev/labels-injector/internal"
)

type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx, "controller", "pod").WithName("Reconcile")

	pod := &v1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); client.IgnoreNotFound(err) != nil {
		// ignore not found errors, could be deleted
		return ctrl.Result{}, err
	}

	log.Info("Reconciling Pod", "pod", req.NamespacedName)

	// Fetch Node object of the pod
	node := &v1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: pod.Spec.NodeName}, node); client.IgnoreNotFound(err) != nil {
		// ignore not found errors, could be deleted
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, internal.TransferLabel(ctx, pod, node, r.Client)
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := logger.FromContext(context.Background(), "controller", "pod")
	log.Info("Setting up Pod controller with manager")

	err := mgr.Add(manager.RunnableFunc(func(context.Context) error {
		return r.ReconcileAllPods()
	}))
	if err != nil {
		log.Error(err, "unable add a runnable to the manager")
	}

	return err
}

func (r *PodReconciler) ReconcileAllPods() error {
	log := logger.FromContext(context.Background(), "controller", "pod").WithName("ReconcileAllPods")

	log.Info("Reconciling all Pods")

	// List all Pods in the cluster
	podList := &v1.PodList{}
	if err := r.List(context.Background(), podList); err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&pod)}); err != nil {
			log.Error(err, "Failed to reconcile Pod", "pod", pod.Name)
			return err
		}
	}

	return nil
}
