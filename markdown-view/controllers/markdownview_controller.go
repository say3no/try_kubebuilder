/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	viewv1 "github.com/try_kubebuilder/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// MarkdownViewReconciler reconciles a MarkdownView object
type MarkdownViewReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=view.say3no.github.io,resources=markdownviews,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=view.say3no.github.io,resources=markdownviews/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=view.say3no.github.io,resources=markdownviews/finalizers,verbs=update
// finalizers なんて subresource あったか？
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;updatepatch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MarkdownView object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MarkdownViewReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

func (r *MarkdownViewReconciler) Reconcile_get(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &deployment)
	if err != nil {
		return ctrl.Result{}, err
	}
	fmt.Printf("Got Deployment: %#v\n", deployment)
	return ctrl.Result{}, nil
}

func (r *MarkdownViewReconciler) Reconcile_list(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var services corev1.ServiceList
	err := r.List(ctx, &services, &client.ListOptions{
		Namespace:     "default",
		LabelSelector: labels.SelectorFromSet((map[string]string{"app": "sample"})),
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, svc := range services.Items {
		fmt.Println(svc.Name)
	}
	return ctrl.Result{}, nil
}

func (r *MarkdownViewReconciler) Reconcile_pagination(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	token := ""
	for i := 0; ; i++ {
		var services corev1.ServiceList
		err := r.List(ctx, &services, &client.ListOptions{
			Limit:    3,
			Continue: token,
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		fmt.Printf("Page %d:\n", i)
		for _, svc := range services.Items {
			fmt.Println(svc.Name)
		}
		fmt.Println()

		token = services.ListMeta.Continue
		if len(token) == 0 {
			return ctrl.Result{}, nil
		}
	}
}

func (r *MarkdownViewReconciler) Reconcile_createOrUpdate(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	svc := &corev1.Service{}
	svc.SetNamespace("default")
	svc.SetName("sample")

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = map[string]string{"app": "nginx"}
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(80),
			},
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		fmt.Printf("Deployment %s\n", op)
	}

	return ctrl.Result{}, nil
}

func (r *MarkdownViewReconciler) Reconcile_patchMerge(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 並行性の問題に対応する
	// patch は特定の フィールドだけを競合することなく更新できる
	// 「ほぼ」同時に pach が投げられたら、たんにキューを処理することになる？

	// 案1. `client.MergeFrom` または `client.StrategicMergeFrom` を活用する
	// .    二者は リスト要素の更新方法が異なる
	//      a. `client.MergeFrom`
	// .       * 指定した要素で上書き
	//      b. `client.StrategicMergeFrom`
	// .       * patchStrategy に応じて要素が追加,または更新
	// 案2. Server-Side Apply を利用する

	// 今回は 案1. a. `client.MergeFrom` でいく

	var dep appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	newDep := dep.DeepCopy()
	newDep.Spec.Replicas = pointer.Int32Ptr(3)
	patch := client.MergeFrom(&dep)

	err = r.Patch(ctx, newDep, patch)

	return ctrl.Result{}, err
}

// Status は subresource にしている。このとき Update や Patch によって ステータスは更新できない。
// 明示的に Status の client を用意してやる必要がある
func (r *MarkdownViewReconciler) updateStatus(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var dep appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	dep.Status.AvailableReplicas = 3
	err = r.Status().Update(ctx, &dep)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MarkdownViewReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&viewv1.MarkdownView{}).
		Complete(r)
}
