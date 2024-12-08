/*
Copyright 2024.

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

// Auth contains authentication credentials
type Auth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Registry contains container registry information
type Registry struct {
	ImageUrl string `json:"imageUrl,omitempty"`
	Auth     *Auth  `json:"auth,omitempty"`
}

// Repository contains git repository information
type Repository struct {
	RepoUrl string `json:"repoUrl,omitempty"`
	Branch  string `json:"branch,omitempty"`
	Auth    *Auth  `json:"auth,omitempty"`
}

// Image contains image build configuration
type Image struct {
	Type     string      `json:"type,omitempty"`
	Repo     *Repository `json:"repo,omitempty"`
	Registry *Registry   `json:"registry,omitempty"`
}

// RelatedImage defines the image build configuration
type RelatedImage struct {
	// Builder specifies the ImageBuild object name
	Builder string `json:"builder,omitempty"`

	// Namespace specifies the ImageBuild object namespace
	// +kubebuilder:default:="default"
	Namespace string `json:"namespace,omitempty"`

	// Version specifies the ImageBuild version
	Version string `json:"version,omitempty"`

	// Digest contains the image digest information
	Digest string `json:"digest,omitempty"`
}

// TaskDefineSpec defines the desired state of TaskDefine
type TaskDefineSpec struct {
	// IDLType specifies the type of the IDL (1-同步器，2-api，3-UI)
	IdlType string `json:"idlType,omitempty"`

	// IDLCode contains the IDL id
	IdlCode string `json:"idlCode,omitempty"`

	// IDLName specifies the name of the IDL
	IdlName string `json:"idlName,omitempty"`

	// IDLVersion specifies the version of the IDL
	IdlVersion string `json:"idlVersion,omitempty"`

	// RelatedImage specifies the image build configuration
	RelatedImage RelatedImage `json:"relatedImage,omitempty"`

	// Definition contains user-defined task configuration
	Definition string `json:"definition,omitempty"`
}

// ImageStatus contains image information from ImageBuild
type ImageStatus struct {
	// Repository contains the full repository URL
	Repository string `json:"repository,omitempty"`

	// Tag contains the image tag
	Tag string `json:"tag,omitempty"`

	// Digest contains the image digest value
	Digest string `json:"digest,omitempty"`

	// PullString contains the complete image pull string
	PullString string `json:"pullString,omitempty"`
}

// BuildStatus contains build-related status information
type BuildStatus struct {
	// Image contains the full image reference
	Image string `json:"image,omitempty"`

	// Time is the timestamp of the build
	Time string `json:"time,omitempty"`

	// Repo contains the repository URL
	Repo string `json:"repo,omitempty"`

	// Branch contains the git branch
	Branch string `json:"branch,omitempty"`
}

// TaskDefineStatus defines the observed state of TaskDefine
type TaskDefineStatus struct {
	// Image contains information from the associated ImageBuild
	Image *ImageStatus `json:"image,omitempty"`

	// LastUpdateTime is the timestamp of the last image info update
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`

	// Build contains information about the latest build
	Build *BuildStatus `json:"build,omitempty"`

	// LastUpdated is the timestamp of the last status update
	LastUpdated string `json:"lastUpdated,omitempty"`

	// State represents the current state of the TaskDefine
	// +kubebuilder:validation:Enum=Pending;Building;Ready;Failed;Invalid
	State string `json:"state,omitempty"`

	// Message provides additional information about the current state
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=td
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// TaskDefine is the Schema for the taskdefines API
type TaskDefine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskDefineSpec   `json:"spec,omitempty"`
	Status TaskDefineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TaskDefineList contains a list of TaskDefine
type TaskDefineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TaskDefine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TaskDefine{}, &TaskDefineList{})
}
