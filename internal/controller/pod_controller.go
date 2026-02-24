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
	"math/rand"
	"time"

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

	err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		return r.ReconcileAllPods(ctx)
	}))
	if err != nil {
		log.Error(err, "unable add a runnable to the manager")
	}

	return err
}

func (r *PodReconciler) ReconcileAllPods(ctx context.Context) error {
	log := logger.FromContext(ctx, "controller", "pod").WithName("ReconcileAllPods")

	// Base interval of 5 minutes
	baseInterval := 5 * time.Minute
	// Random skew up to 1 minute (±30 seconds)
	maxSkew := 60 * time.Second

	for {
		log.Info("Reconciling all Pods")

		// List all Pods in the cluster
		podList := &v1.PodList{}
		if err := r.List(ctx, podList); err != nil {
			log.Error(err, "Failed to list Pods")
			// Continue on error, will retry on next iteration
		} else {
			for _, pod := range podList.Items {
				if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&pod)}); err != nil {
					log.Error(err, "Failed to reconcile Pod", "pod", pod.Name)
					// Continue reconciling other pods even if one fails
				}
			}
		}

		// Calculate next interval with random skew
		skew := time.Duration(rand.Int63n(int64(maxSkew))) - maxSkew/2 //nolint:gosec
		nextInterval := baseInterval + skew
		log.Info("Next reconciliation scheduled", "interval", nextInterval)

		select {
		case <-time.After(nextInterval):
			// Continue to next iteration
		case <-ctx.Done():
			log.Info("Stopping ReconcileAllPods loop")
			return ctx.Err()
		}
	}
}
