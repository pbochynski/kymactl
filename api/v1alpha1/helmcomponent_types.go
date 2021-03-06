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

// HelmComponentSpec defines the desired state of HelmComponent
type HelmComponentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the component (chart name)
	ComponentName string `json:"componentName,omitempty"`

	// Location of the chart. If not provided it is folder in the kyma resources named as the component (convention)
	ChartLocation string `json:"chartLocation,omitempty"`

	// Component version (Kyma version)
	Version string `json:"version,omitempty"`

	// Target namespace where component should be installed. If not provided: kyma-system
	Namespace string `json:"namespace,omitempty"`
}

// HelmComponentStatus defines the observed state of HelmComponent
type HelmComponentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// A list of pointers to currently running jobs.

	// +optional
	Status string `json:"status,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastReconciliation *metav1.Time `json:"lastReconciliation,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"

// HelmComponent is the Schema for the helmcomponents API
type HelmComponent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmComponentSpec   `json:"spec,omitempty"`
	Status HelmComponentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HelmComponentList contains a list of HelmComponent
type HelmComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmComponent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmComponent{}, &HelmComponentList{})
}
