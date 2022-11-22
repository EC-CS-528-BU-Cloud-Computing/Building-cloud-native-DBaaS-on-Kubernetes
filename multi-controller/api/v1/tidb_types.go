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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
const (
	PhasePending_TiDB  = "PENDING_TiDB"
	PhaseCreating_TiDB = "CREATING_TiDB"
	PhaseRunning_TiDB  = "RUNNING_TiDB"
)

// TidbSpec defines the desired state of Tidb
type TidbSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Imagename is the name of image this Tidb instance is using
	Imagename string `json:"imagename,omitempty"`

	// HealthCheckInterval is the interval for periodic health check, if pod failed/deleted, it should recreate the pod
	HealthCheckInterval int `json:"healthcheck.interval,omitempty"`
}

// TidbStatus defines the observed state of Tidb
type TidbStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//Phase represents current phase of tidb
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tidb is the Schema for the tidbs API
type Tidb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TidbSpec   `json:"spec,omitempty"`
	Status TidbStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TidbList contains a list of Tidb
type TidbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tidb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tidb{}, &TidbList{})
}
