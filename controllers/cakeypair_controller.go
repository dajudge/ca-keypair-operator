/*
Copyright 2021 The CA-KeyPair-Operator Authors.

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
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"math/big"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cakeypairsv1alpha1 "cakeypair-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// CaKeyPairReconciler reconciles a CaKeyPair object
type CaKeyPairReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

// +kubebuilder:rbac:groups=cakeypairs.dajudge.com,resources=cakeypairs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cakeypairs.dajudge.com,resources=cakeypairs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *CaKeyPairReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cakeypair", req.NamespacedName)

	var caKeyPair cakeypairsv1alpha1.CaKeyPair
	if err := r.Get(ctx, req.NamespacedName, &caKeyPair); err != nil {
		log.Error(err, "Failed to load CA key pair")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("CA key pair loaded", "data", caKeyPair)

	if r.SecretRenamed(caKeyPair) {
		oldSecretName := types.NamespacedName{Namespace: caKeyPair.Namespace, Name: caKeyPair.Status.Secret.Name}
		if err := r.DeleteExistingSecret(ctx, oldSecretName, log); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		log.Info("Deleted existing secret for CA key pair", "secret", oldSecretName)
	}

	newSecret, err := r.GetOrCreateSecret(caKeyPair, ctx, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	if caKeyPair.Status.Secret.Name != newSecret.Name {
		if err := r.UpdateKeyPairStatus(caKeyPair, newSecret, ctx, log); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		log.Info("Updated status of CA key pair")
	} else {
		log.Info("CA key pair already has correct status")
	}

	return ctrl.Result{}, nil
}

func (r *CaKeyPairReconciler) GetOrCreateSecret(
	caKeyPair cakeypairsv1alpha1.CaKeyPair,
	ctx context.Context,
	log logr.Logger,
) (corev1.Secret, error) {
	newSecretName := types.NamespacedName{Namespace: caKeyPair.Namespace, Name: caKeyPair.Spec.SecretName}
	secretExists, existingSecret, err := r.LoadExistingSecret(ctx, newSecretName, log)
	if err != nil {
		return corev1.Secret{}, err
	}
	if secretExists {
		log.Info("Secret for CA key pair already exists", "secret", newSecretName)
		return existingSecret, nil
	}

	newSecret, err := r.CreateNewKeyPair(caKeyPair, log, ctx)
	if err != nil {
		return corev1.Secret{}, err
	}
	log.Info("Created new secret for CA key pair", "secret", newSecretName)
	return newSecret, nil
}

func (r *CaKeyPairReconciler) UpdateKeyPairStatus(
	caKeyPair cakeypairsv1alpha1.CaKeyPair,
	newSecret v1.Secret,
	ctx context.Context,
	log logr.Logger,
) error {
	caKeyPair.Status.Secret = corev1.ObjectReference{
		Kind:            newSecret.Kind,
		Namespace:       newSecret.Namespace,
		Name:            newSecret.Name,
		UID:             newSecret.UID,
		APIVersion:      newSecret.APIVersion,
		ResourceVersion: newSecret.ResourceVersion,
	}
	if err := r.Status().Update(ctx, &caKeyPair); err != nil {
		log.Error(err, "Failed to update status of CA key pair")
		return err
	}
	return nil
}

func (r *CaKeyPairReconciler) LoadExistingSecret(
	ctx context.Context,
	newSecretName types.NamespacedName,
	log logr.Logger,
) (bool, v1.Secret, error) {
	var existingSecret v1.Secret
	if err := r.Get(ctx, newSecretName, &existingSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return false, v1.Secret{}, nil
		}
		log.Error(err, "Error locating existing secret", "secret", newSecretName)
		return false, v1.Secret{}, err
	} else {
		return true, existingSecret, nil
	}
}

func (r *CaKeyPairReconciler) CreateNewKeyPair(
	caKeyPair cakeypairsv1alpha1.CaKeyPair,
	log logr.Logger,
	ctx context.Context,
) (v1.Secret, error) {
	newSecret, err := r.InitNewKeyPairSecret(caKeyPair, log)
	if err != nil {
		log.Error(err, "Failed initialize key pair")
		return v1.Secret{}, err
	}
	if err := r.Create(ctx, &newSecret); err != nil {
		newSecretName := types.NamespacedName{Namespace: newSecret.Namespace, Name: newSecret.Namespace}
		log.Error(err, "Failed to create secret for CA key pair", "secret", newSecretName)
		return v1.Secret{}, err
	}
	return newSecret, nil
}

func (r *CaKeyPairReconciler) InitNewKeyPairSecret(
	caKeyPair cakeypairsv1alpha1.CaKeyPair,
	log logr.Logger,
) (v1.Secret, error) {
	keyBuffer, certBuffer, err := r.InitNewKeyPair(caKeyPair, log)
	if err != nil {
		return v1.Secret{}, err
	}
	ref := metav1.OwnerReference{
		APIVersion: caKeyPair.APIVersion,
		Kind:       caKeyPair.Kind,
		Name:       caKeyPair.Name,
		UID:        caKeyPair.UID,
	}
	newSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            caKeyPair.Spec.SecretName,
			Namespace:       caKeyPair.Namespace,
			OwnerReferences: []metav1.OwnerReference{ref},
		},
		Data: map[string][]byte{
			"tls.key": keyBuffer.Bytes(),
			"tls.crt": certBuffer.Bytes(),
		},
		Type: v1.SecretTypeOpaque,
	}
	return newSecret, nil
}

func (r *CaKeyPairReconciler) InitNewKeyPair(
	caKeyPair cakeypairsv1alpha1.CaKeyPair,
	log logr.Logger,
) (*bytes.Buffer, *bytes.Buffer, error) {
	start := makeTimestamp()
	priv, err := rsa.GenerateKey(rand.Reader, int(caKeyPair.Spec.KeySize))
	if err != nil {
		log.Error(err, "Failed to generate private key for key pair")
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:         caKeyPair.Spec.CommonName,
			Organization:       caKeyPair.Spec.Subject.Organizations,
			Country:            caKeyPair.Spec.Subject.Countries,
			OrganizationalUnit: caKeyPair.Spec.Subject.OrganizationalUnits,
			Locality:           caKeyPair.Spec.Subject.Localities,
			Province:           caKeyPair.Spec.Subject.Provices,
			StreetAddress:      caKeyPair.Spec.Subject.StreetAddresses,
			PostalCode:         caKeyPair.Spec.Subject.PostalCodes,
			SerialNumber:       caKeyPair.Spec.Subject.SerialNumber,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Error(err, "Failed to create certificate for CA key pair")
		return nil, nil, err
	}
	keyBuffer := &bytes.Buffer{}
	_ = pem.Encode(keyBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	certBuffer := &bytes.Buffer{}
	_ = pem.Encode(certBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	end := makeTimestamp()
	log.Info("New key pair initialized", "millis", end-start)
	return keyBuffer, certBuffer, nil
}

func (r *CaKeyPairReconciler) SecretRenamed(keyPair cakeypairsv1alpha1.CaKeyPair) bool {
	return keyPair.Status.Secret.Name != keyPair.Spec.SecretName && len(keyPair.Status.Secret.Name) > 0
}

func (r *CaKeyPairReconciler) DeleteExistingSecret(
	ctx context.Context,
	oldSecretName types.NamespacedName,
	log logr.Logger,
) error {
	var oldSecret v1.Secret
	log.Info("Ensuring nonexistence of old secret", "oldSecretName", oldSecretName)
	if err := r.Get(ctx, oldSecretName, &oldSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Old secret does not exist, anyway")
			return nil
		} else {
			log.Error(err, "Failed to locate existing secret", "secret", oldSecretName)
			return err
		}
	}
	log.Info("Old secret is present", "oldSecretName", oldSecretName)
	if err := r.Delete(ctx, &oldSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to delete existing secret", "secret", oldSecretName)
			return err
		}
	}
	log.Info("Old secret deleted", "oldSecret", oldSecretName)
	return nil
}

func (r *CaKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cakeypairsv1alpha1.CaKeyPair{}).
		Complete(r)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
