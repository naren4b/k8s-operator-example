package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageArraySpec defines the desired state of ImageArray
type ImageArraySpec struct {
    Images []string `json:"images"`
}

// ImageArrayStatus defines the observed state of ImageArray
type ImageArrayStatus struct {
    // Status fields
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ImageArray is the Schema for the imagearrays API
type ImageArray struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   ImageArraySpec   `json:"spec,omitempty"`
    Status ImageArrayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImageArrayList contains a list of ImageArray
type ImageArrayList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []ImageArray `json:"items"`
}

func init() {
    SchemeBuilder.Register(&ImageArray{}, &ImageArrayList{})
}
