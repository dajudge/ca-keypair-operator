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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CaKeyPairSpec defines the desired state of CaKeyPair
type CaKeyPairSpec struct {
	// KeySize specifies the key size in bits of the generated keypair
	KeySize int32 `json:"keySize,omitempty"`

	// SecretName specifies the name of secret to store the generated keypair in
	SecretName string `json:"secretName,omitempty"`

	// Subject the subject for the CA cert
	Subject CaKeyPairSubject `json:"subject,omitempty"`

	// The common name
	CommonName string `json:"commonName"`
}

// CaKeyPairStatus defines the observed state of CaKeyPair
type CaKeyPairStatus struct {
	// Reference to the secret
	Secret corev1.ObjectReference `json:"active,omitempty"`
}

// CaKeyPairSubject defines the full X509 name specification
type CaKeyPairSubject struct {
	// Organizations
	Organizations []string `json:"organizations,omitempty"`

	// Countries
	Countries []string `json:"countries,omitempty"`

	// Organizational units
	OrganizationalUnits []string `json:"organizationalUnits,omitempty"`

	// Localities
	Localities []string `json:"localities,omitempty"`

	// Provinces
	Provices []string `json:"provices,omitempty"`

	// Street addresses
	StreetAddresses []string `json:"streetAddresses,omitempty"`

	// Postal codes
	PostalCodes []string `json:"postalCodes,omitempty"`

	// Serial number
	SerialNumber string `json:"serialNumber,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CaKeyPair is the Schema for the cakeypairs API
type CaKeyPair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaKeyPairSpec   `json:"spec,omitempty"`
	Status CaKeyPairStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CaKeyPairList contains a list of CaKeyPair
type CaKeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaKeyPair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaKeyPair{}, &CaKeyPairList{})
}
