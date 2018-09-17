/*
Copyright 2018 The Knative Authors

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

	"github.com/knative/pkg/apis/duck"
)

// Targettable is the schema for the targettable portion of the payload
type Targettable struct {
	// TODO(vaikas): Give me a schema!
	Field string `json:"field,omitempty"`
}

// Implementations can verify that they implement Targettable via:
var _ = duck.VerifyType(&Target{}, &Targettable{})

// Targettable is an Implementable "duck type".
var _ duck.Implementable = (*Targettable)(nil)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Target is a skeleton type wrapping Targettable in the manner we expect
// resource writers defining compatible resources to embed it.  We will
// typically use this type to deserialize Targettable ObjectReferences and
// access the Targettable data.  This is not a real resource.
type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TargetStatus `json:"status"`
}

// TargetStatus shows how we expect folks to embed Targettable in
// their Status field.
type TargetStatus struct {
	Targettable *Targettable `json:"targettable,omitempty"`
}

// In order for Targettable to be Implementable, Target must be Populatable.
var _ duck.Populatable = (*Target)(nil)

// GetFullType implements duck.Implementable
func (_ *Targettable) GetFullType() duck.Populatable {
	return &Target{}
}

// Populate implements duck.Populatable
func (t *Target) Populate() {
	t.Status.Targettable = &Targettable{
		// Populate ALL fields
		Field: "this is not empty",
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TargetList is a list of Target resources
type TargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Target `json:"items"`
}
