// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Service is a wrapper over service api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type Service struct {
	object *corev1.Service
}

// ServiceList is a wrapper over service api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type ServiceList struct {
	items []*Service
}

// Len returns the number of items present
// in the ServiceList
func (s *ServiceList) Len() int {
	return len(s.items)
}

// ToAPIList converts ServiceList to API ServiceList
func (s *ServiceList) ToAPIList() *corev1.ServiceList {
	slist := &corev1.ServiceList{}
	for _, service := range s.items {
		slist.Items = append(slist.Items, *service.object)
	}
	return slist
}

type serviceBuildOption func(*Service)

// NewForAPIObject returns a new instance of Service
func NewForAPIObject(obj *corev1.Service,
	opts ...serviceBuildOption) *Service {
	s := &Service{object: obj}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided service instance
type Predicate func(*Service) bool

// IsNil returns true if the Service instance
// is nil
func (s *Service) IsNil() bool {
	return s.object == nil
}

// IsNil is predicate to filter out nil Service
// instances
func IsNil() Predicate {
	return func(s *Service) bool {
		return s.IsNil()
	}
}

// ContainsName is filter function to filter service's
// based on the name
func ContainsName(name string) Predicate {
	return func(s *Service) bool {
		return strings.Contains(s.object.GetName(), name)
	}
}
