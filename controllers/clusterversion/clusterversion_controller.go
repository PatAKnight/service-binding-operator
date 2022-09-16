/*
Copyright 2021 Red Hat OpenShift Data Foundation.
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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	configv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/redhat-developer/service-binding-operator/consoleplugin"
)

// ClusterVersionReconciler reconciles a ClusterVersion object
type ClusterVersionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Namespace string
}

//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ClusterVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := log.FromContext(ctx)
	instance := configv1.ClusterVersion{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, &instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcilePluginResources()
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.Add(manager.RunnableFunc(func(context.Context) error {
		return r.reconcilePluginResources()
	}))
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1.ClusterVersion{}).
		Complete(r)
}

func (r *ClusterVersionReconciler) reconcilePluginResources() error {
	namespace := r.Namespace
	log := log.FromContext(context.TODO())

	pluginDeployment := consoleplugin.Deployment(namespace)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, pluginDeployment, func() error {
		return nil
	})
	if err != nil {
		return err
	} else {
		log.Info("Deployment successfully reconciled", "operation", op)
	}
	
	err = r.reconcileService(pluginDeployment, namespace, log)
	if err != nil {
		return err
	}
	err = r.reconcileConfigMap(pluginDeployment, namespace, log)
	if err != nil {
		return err
	}
	err = r.reconcileConsolePlugin(namespace, log)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterVersionReconciler) reconcileService(pluginDeployment *appsv1.Deployment, namespace string, log logr.Logger) error {
	pluginService := consoleplugin.Service(namespace)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, pluginService, func() error {
		return controllerutil.SetControllerReference(pluginDeployment, pluginService, r.Scheme)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else {
		log.Info("Service successfully reconciled", "operation", op)
	}
	return nil
}

func (r *ClusterVersionReconciler) reconcileConfigMap(pluginDeployment *appsv1.Deployment, namespace string, log logr.Logger) error {
	pluginConfigMap := consoleplugin.ConfigMap(namespace)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, pluginConfigMap, func() error {
		return controllerutil.SetControllerReference(pluginDeployment, pluginConfigMap, r.Scheme)
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else {
		log.Info("ConfigMap successfully reconciled", "operation", op)
	}
	return nil
}

func (r *ClusterVersionReconciler) reconcileConsolePlugin(namespace string, log logr.Logger) error {
	consolePlugin := consoleplugin.ConsolePlugin(namespace)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, consolePlugin, func() error {
		return nil
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else {
		log.Info("Console Plugin successfully reconciled", "operation", op)
	}
	return nil
}