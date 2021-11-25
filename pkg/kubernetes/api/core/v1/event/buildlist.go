/*
Copyright 2021 The OpenEBS Authors

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

package event

import (
	corev1 "k8s.io/api/core/v1"
)

type ListBuilder struct {
	list    *EventList
	filters PredicateList
}

func ListBuilderFromAPIList(events *corev1.EventList) *ListBuilder {
	b := &ListBuilder{list: &EventList{}}
	if events == nil {
		return b
	}
	for _, event := range events.Items {
		event := event
		b.list.Items = append(b.list.Items, &Event{Object: &event})
	}
	return b
}

// List returns the list of Event
// instances that was built by this
// builder
func (b *ListBuilder) List() *EventList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &EventList{}
	for _, event := range b.list.Items {
		if b.filters.all(event) {
			filtered.Items = append(filtered.Items, event)
		}
	}
	return filtered
}

// WithFilter add filters on which the Event
// has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
