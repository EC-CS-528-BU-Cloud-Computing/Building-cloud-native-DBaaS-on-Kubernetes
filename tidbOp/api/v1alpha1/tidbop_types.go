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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	PhasePending  = "PENDING"
	PhaseCreating = "CREATING"
	PhaseRunning  = "RUNNING"
)

// TidbopSpec defines the desired state of Tidbop
type TidbopSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Imagename is the name of image this Tidb instance is using
	Imagename string `json:"imagename,omitempty"`

	// HealthCheckInterval is the interval for periodic health check, if pod failed/deleted, it should recreate the pod
	HealthCheckInterval int `json:"healthcheck.interval,omitempty"`
}

// TidbopStatus defines the observed state of Tidbop
type TidbopStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tidbop is the Schema for the tidbops API
type Tidbop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TidbopSpec   `json:"spec,omitempty"`
	Status TidbopStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TidbopList contains a list of Tidbop
type TidbopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tidbop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tidbop{}, &TidbopList{})
}
