/*
Copyright 2023.

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
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gcpsecretv1alpha1 "tutorial.kubebuilder.io/project/api/v1alpha1"
)

// AccessTokenReconciler reconciles a AccessToken object
type AccessTokenReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gcpsecret.secret.arpan.io,resources=accesstokens,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gcpsecret.secret.arpan.io,resources=accesstokens/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gcpsecret.secret.arpan.io,resources=accesstokens/finalizers,verbs=update
//+kubebuilder:rbac:groups=v1,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AccessToken object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *AccessTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("GCPAcessToken", req.Name, "NameSpace", req.Namespace)
	token := &gcpsecretv1alpha1.AccessToken{}

	// find all object of this type
	err := r.Get(ctx, req.NamespacedName, token)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Controller Resources Are Not Avialable", "GCPACESSTOKEN", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to List Details Check details please")
		return ctrl.Result{}, err
	}
	// check if the required secrets Are already in place and needs and update

	secrets := &v1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Namespace: token.Spec.TargetNamespace, Name: token.Spec.TargetName}, secrets)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Secret Need to be create on", "Name", token.Spec.TargetName, "Namespace", token.Spec.TargetNamespace, "Operator", token.Spec.Namespace)
			sec := r.SecretForOperator(token)
			err := r.Create(ctx, sec)
			if err != nil {
				log.Error(err, "Unable to create secret", "Name", token.Spec.TargetName, "NameSpace", token.Spec.Namespace)
			}
			return ctrl.Result{Requeue: true}, nil
		}
	} else if err != nil {
		log.Error(err, "Unable To Get The Details")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AccessTokenReconciler) SecretForOperator(token *gcpsecretv1alpha1.AccessToken) *v1.Secret {
	ls := map[string]string{"operator": token.Name, "OperatorNamespace": token.Namespace}
	immutable := false
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: token.Spec.TargetName, Namespace: token.Spec.TargetNamespace, Labels: ls},
		Immutable:  &immutable,
		Data:       nil,
		StringData: map[string]string{
			"username": "my-username",
			"password": "my-password",
		},
	}
	return secret
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccessTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gcpsecretv1alpha1.AccessToken{}).
		Owns(&v1.Secret{}).
		Complete(r)
}
