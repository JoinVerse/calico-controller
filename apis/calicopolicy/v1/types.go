package v1

import (
	calicoapi "github.com/projectcalico/libcalico-go/lib/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CalicoPolicyResourcePlural = "CalicoPolicies"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CalicoPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              calicoapi.PolicySpec `json:"spec"`
	Status            CalicoPolicyStatus   `json:"status,omitempty"`
}

type CalicoPolicyStatus struct {
	State   CalicoPolicyState `json:"state,omitempty"`
	Message string            `json:"message,omitempty"`
}

type CalicoPolicyState string

const (
	CalicoPolicyStateCreated   CalicoPolicyState = "Created"
	CalicoPolicyStateProcessed CalicoPolicyState = "Processed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CalicoPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CalicoPolicy `json:"items"`
}
