package v1

import (
	calicoapi "github.com/projectcalico/libcalico-go/lib/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CalicoPolicyResourcePlural = "CalicoPolicies"

type CalicoPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              calicoapi.PolicySpec `json:"spec"`
	Status            CalicoPolicyStatus   `json:"status,omitempty"`
}

type CalicoPolicyStatus struct {
	Message string `json:"message,omitempty"`
}

type CalicoPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CalicoPolicy `json:"items"`
}
