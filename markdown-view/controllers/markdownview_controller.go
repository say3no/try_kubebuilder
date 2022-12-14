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

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
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
	logger := log.FromContext(ctx)

	var mdView viewv1.MarkdownView
	err := r.Get(ctx, req.NamespacedName, &mdView)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "unable to get MarkdownView", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	// 削除済み
	if !mdView.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// reconcile の config map をとって
	err = r.reconcileConfigMap(ctx, mdView)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileDeployment(ctx, mdView)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileService(ctx, mdView)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, mdView)
}

func (r *MarkdownViewReconciler) reconcileDeployment(ctx context.Context, mdView viewv1.MarkdownView) error {
	logger := log.FromContext(ctx)

	depName := "viewer-" + mdView.Name
	viewerImage := "peaceiris/mdbook:latest" // default, if spec is not defined
	if len(mdView.Spec.ViewerImage) != 0 {
		viewerImage = mdView.Spec.ViewerImage
	}

	dep := appsv1apply.Deployment(depName, mdView.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "mdbook",
			"app.kubernetes.io/instance":   mdView.Name,
			"app.kubernetes.io/created-by": "markdown-view-controller",
		}).
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(mdView.Spec.Replicas).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(map[string]string{
				"app.kubernetes.io/name":       "mdbook",
				"app.kubernetes.io/instance":   mdView.Name,
				"app.kubernetes.io/created-by": "markdown-view-controller",
			})).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(map[string]string{
					"app.kubernetes.io/name":       "mdbook",
					"app.kubernetes.io/instance":   mdView.Name,
					"app.kubernetes.io/created-by": "markdown-view-controller",
				}).
				WithSpec(corev1apply.PodSpec().
					WithContainers(corev1apply.Container().
						WithName("mdbook").
						WithImage(viewerImage).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommand("mdbook").
						WithArgs("serve", "--hostname", "0.0.0.0").
						WithVolumeMounts(corev1apply.VolumeMount().
							WithName("markdowns").
							WithMountPath("/book/src"),
						).
						WithPorts(corev1apply.ContainerPort().
							WithName("http").
							WithProtocol(corev1.ProtocolTCP).
							WithContainerPort(3000),
						).
						WithLivenessProbe(corev1apply.Probe().
							WithHTTPGet(corev1apply.HTTPGetAction().
								WithPort(intstr.FromString("http")).
								WithPath("/").
								WithScheme(corev1.URISchemeHTTP),
							),
						).
						WithReadinessProbe(corev1apply.Probe().
							WithHTTPGet(corev1apply.HTTPGetAction().
								WithPort(intstr.FromString("http")).
								WithPath("/").
								WithScheme(corev1.URISchemeHTTP),
							),
						),
					).
					WithVolumes(corev1apply.Volume().
						WithName("markdowns").
						WithConfigMap(corev1apply.ConfigMapVolumeSource().
							WithName("markdowns-" + mdView.Name),
						),
					),
				),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dep)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: depName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, "markdown-view-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "markdown-view-controller",
		Force:        pointer.Bool(true),
	})

	if err != nil {
		logger.Error(err, "unable to create or update Deployment")
		return err
	}
	logger.Info("reconcile Deployment successfully", "name", mdView.Name)
	return nil
}

func (r *MarkdownViewReconciler) reconcileService(ctx context.Context, mdView viewv1.MarkdownView) error {
	logger := log.FromContext(ctx)
	svcName := "viewer-" + mdView.Name

	svc := corev1apply.Service(svcName, mdView.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "mdbook",
			"app.kubernetes.io/instance":   mdView.Name,
			"app.kubernetes.io/created-by": "markdown-view-controller",
		}).
		WithSpec(corev1apply.ServiceSpec().
			WithSelector(map[string]string{
				"app.kubernetes.io/name":       "mdbook",
				"app.kubernetes.io/instance":   mdView.Name,
				"app.kubernetes.io/created-by": "markdown-view-controller",
			}).
			WithType(corev1.ServiceTypeClusterIP).
			WithPorts(corev1apply.ServicePort().
				WithProtocol(corev1.ProtocolTCP).
				WithPort(80).
				WithTargetPort(intstr.FromInt(3000)),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svc)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current corev1.Service
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: svcName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := corev1apply.ExtractService(&current, "markdown-view-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(svc, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "markdown-view-controller",
		Force:        pointer.Bool(true),
	})
	if err != nil {
		logger.Error(err, "unable to create or update Service")
		return err
	}

	logger.Info("reconcile Service successfully", "name", mdView.Name)
	return nil
}

func (r *MarkdownViewReconciler) reconcileConfigMap(ctx context.Context, mdView viewv1.MarkdownView) error {
	logger := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(mdView.Namespace)
	cm.SetName("markdowns-" + mdView.Name)

	// CreateOrUpdate -> Get after create or update
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		for name, content := range mdView.Spec.Markdowns {
			cm.Data[name] = content
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "unable to create or update ConfigMap")
		return err
	}
	if op != controllerutil.OperationResultNone {
		logger.Info("reconcile COnfigMap successfully", "op", op)
	}
	return nil
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
func (r *MarkdownViewReconciler) updateStatus(ctx context.Context, mdView viewv1.MarkdownView) (ctrl.Result, error) {
	var dep appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: "viewer-" + mdView.Name}, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	var status viewv1.MarkdownViewStatus
	if dep.Status.AvailableReplicas == 0 {
		status = viewv1.MarkdownViewNotReady
	} else if dep.Status.AvailableReplicas == mdView.Spec.Replicas {
		status = viewv1.MarkdownViewHealthy
	} else {
		status = viewv1.MarkdownViewAvailable
	}

	if mdView.Status != status {
		mdView.Status = status
		err = r.Status().Update(ctx, &mdView)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if mdView.Status != status {
		mdView.Status = status
		err = r.Status().Update(ctx, &mdView)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *MarkdownViewReconciler) Reconcile_deleteWithPreConditions(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deploy appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &deploy)
	if err != nil {
		return ctrl.Result{}, err
	}
	uid := deploy.GetUID()
	resourceVersion := deploy.GetResourceVersion()
	cond := metav1.Preconditions{
		UID:             &uid,
		ResourceVersion: &resourceVersion,
	}
	// preconditions で uid レベルでの validation を行い uniqness を担保
	err = r.Delete(ctx, &deploy, &client.DeleteOptions{Preconditions: &cond})
	return ctrl.Result{}, err
}

func (r *MarkdownViewReconciler) Reconcile_deleteAllOfDeployment(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	err := r.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("default"))
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MarkdownViewReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&viewv1.MarkdownView{}).
		Complete(r)
}
