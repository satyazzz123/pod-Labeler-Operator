/*
Copyright 2024 satyajit Behera.

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

package controller

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/satyazzz123/pod-Labeler-Operator/api/v1alpha1"
)

// PodLabelerReconciler reconciles a PodLabeler object
type PodLabelerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.satyazzz123.online,resources=podlabelers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.satyazzz123.online,resources=podlabelers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.satyazzz123.online,resources=podlabelers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodLabeler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *PodLabelerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here

	//Fetch the PodLabeler instance or CR created
	podlabeler := &apiv1alpha1.PodLabeler{}
	err := r.Get(ctx, req.NamespacedName, podlabeler)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("PodLabeler resource not found. Ignoring since object must be deleted.")
		}
		logger.Error(err, "Failed to get PodLabler resource")
		return ctrl.Result{}, err
	}

	// List all pods matching the selector in PodLabeler Resource so that it will label pods only with the specified Selector
	podList := &v1.PodList{}
	// listOpts := []client.ListOption{
	// 	client.InNamespace(req.Namespace),
	// 	client.MatchingLabels(podlabeler.Spec.Selector),
	// }

	if err := r.List(ctx, podList, client.InNamespace(req.Namespace), client.MatchingLabels(podlabeler.Spec.Selector)); err != nil {
		logger.Error(err, "Failed to list Pods with Selectors Matching the CR")
	}

	//Now we add Labels to the pods that were Matched Above
	for _, pod := range podList.Items {
		if pod.ObjectMeta.Labels == nil {
			pod.ObjectMeta.Labels = map[string]string{} //initializes an empty Labels if no label is present

		}
		for key, value := range podlabeler.Spec.Labels { //basically it is iteratting over the map of spec in CR so that it can assign that label to above matching pod
			pod.ObjectMeta.Labels[key] = value //key from CR label and value from CR label for the attaching new label to pod

		}
		//Time to finally update the Pod
		if err := r.Update(ctx, &pod); err != nil {
			logger.Error(err, fmt.Sprintf("failed to Update Pod %s/%s labels successfully", pod.Namespace, pod.Name))
		}
		logger.Info(fmt.Sprintf("Updated Pod %s/%s labels successfully", pod.Namespace, pod.Name))
	}

	return ctrl.Result{RequeueAfter: time.Duration(15 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodLabelerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.PodLabeler{}).
		Complete(r)
}
