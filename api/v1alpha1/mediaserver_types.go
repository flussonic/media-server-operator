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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MediaServerVolume struct {
	// Volume and corresponding VolumeMount name
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// VolumeMount mountPath
	MountPath string `json:"mountPath" protobuf:"bytes,1,opt,name=mountPath"`

	// volumeSource represents the location and type of the mounted volume.
	// If not specified, the Volume is implied to be an EmptyDir.
	// This implied behavior is deprecated and will be removed in a future version.
	corev1.VolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

// MediaServerSpec defines the desired state of MediaServer
type MediaServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MediaServer. Edit mediaserver_types.go to remove/update
	Image string `json:"image,omitempty"`
	// (Optional) PodEnvVariables is a slice of environment variables that are added to the pods
	// Default: (empty list)
	// +optional
	PodEnvVariables []corev1.EnvVar `json:"env,omitempty"`
	// (Optional) node selector for placing pods with Media Server instances. They are deployed
	// as DaemonSet, one per each streaming node. Here you can select which nodes to use for streaming
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
	// (Optional) if you decide to refuse from Ingress usage for streaming (good idea to refuse) and
	// stream directly from each Pod to the wild Internet, use this HostPort option to expose port
	HostPort int32 `json:"hostPort,omitempty" protobuf:"varint,2,opt,name=hostPort"`
	// (Optional) only for debug to access admin UI of each Media Server
	AdminHostPort int32 `json:"adminHostPort,omitempty" protobuf:"varint,2,opt,name=adminHostPort"`
	// (Optional) place additional files to /etc/flussonic/flussonic.conf.d to configure Media Server
	ConfigExtra map[string]string `json:"configExtra,omitempty" protobuf:"bytes,2,rep,name=configExtra"`
	// (Optiona) additionally mounted volumes
	Volumes []MediaServerVolume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,2,rep,name=volumes"`
}

// MediaServerStatus defines the observed state of MediaServer
type MediaServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MediaServer is the Schema for the mediaservers API
type MediaServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MediaServerSpec   `json:"spec,omitempty"`
	Status MediaServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MediaServerList contains a list of MediaServer
type MediaServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MediaServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MediaServer{}, &MediaServerList{})
}
