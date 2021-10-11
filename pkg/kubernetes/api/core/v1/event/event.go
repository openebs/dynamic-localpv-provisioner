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
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Event holds the API's Event object
type Event struct {
	Object *corev1.Event
}

// EventList holds the list of API Event instances
type EventList struct {
	Items []*Event
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// Predicate defines an abstraction
// to determine conditional checks
// against the provided Event instance
type Predicate func(*Event) bool

// all returns true if all the predicates
// succeed against the provided Event
// instance
func (l PredicateList) all(event *Event) bool {
	for _, pred := range l {
		if !pred(event) {
			return false
		}
	}
	return true
}

// Len returns the number of items present in the EventList
// Note: Required for AscendingSortByLastTimestamp()
// and DescendingSortByLastTimestamp()
func (eventList *EventList) Len() int {
	return len(eventList.Items)
}

// Swap swaps two *Event instances in the EventList
// Note: Required for AscendingSortByLastTimestamp()
// and DescendingSortByLastTimestamp()
func (eventList *EventList) Swap(i, j int) {
	eventList.Items[i], eventList.Items[j] = eventList.Items[j], eventList.Items[i]
}

// Checks if the object.LastTimestamp for the Event
// instance at 'i' index was earlier than the one
// for the Event at 'j' index
// Note: Required for AscendingSortByLastTimestamp()
// and DescendingSortByLastTimestamp()
func (eventList *EventList) Less(i, j int) bool {
	return eventList.Items[i].Object.LastTimestamp.Time.Before(eventList.Items[j].Object.LastTimestamp.Time)
}

// Sorts Events in the EventList starting with
// the least recent one
func (eventList *EventList) LatestLastSort() *EventList {
	sort.Stable(eventList)
	return eventList
}

// Sorts Events in the EventList starting with
// the most recent one
func (eventList *EventList) LatestFirstSort() *EventList {
	sort.Stable(sort.Reverse(eventList))
	return eventList
}

// isBdcEvent() returns 'true' if the Kind field
// of the InvoledObject struct is BlockDeviceClaim
func (event *Event) isBdcEvent() bool {
	return event.Object.InvolvedObject.Kind == "BlockDeviceClaim"
}

// Returns a predicate from isBdcEvent() function
func IsBdcEvent() Predicate {
	return func(event *Event) bool {
		return event.isBdcEvent()
	}
}

// isBdEvent() returns 'true' if the Kind field
// of the InvoledObject struct is BlockDevice
func (event *Event) isBdEvent() bool {
	return event.Object.InvolvedObject.Kind == "BlockDevice"
}

// Returns a predicate from isBdEvent() function
func IsBdEvent() Predicate {
	return func(event *Event) bool {
		return event.isBdEvent()
	}
}

// isPodEvent() returns 'true' if the Kind field
// of the InvoledObject struct is Pod
func (event *Event) isPodEvent() bool {
	return event.Object.InvolvedObject.Kind == "Pod"
}

// Returns a predicate from isPodEvent() function
func IsPodEvent() Predicate {
	return func(event *Event) bool {
		return event.isPodEvent()
	}
}

// Queries the Reason field in the InvolvedObject
// instance in the corev1.Event
// for an exact match of the argument string
func (event *Event) hasReason(reason string) bool {
	return event.Object.Reason == reason
}

// Returns a predicate from hasReason() function
func HasReason(reason string) Predicate {
	return func(event *Event) bool {
		return event.hasReason(reason)
	}
}

// Queries the Message field in the InvolvedObject
// instance in the corev1.Event
// for substring that matches the argument string
func (event *Event) hasStringInMessage(substr string) bool {
	return strings.Contains(event.Object.Message, substr)
}

// Returns a predicate from hasMessage() function
func HasStringInMessage(substr string) Predicate {
	return func(event *Event) bool {
		return event.hasStringInMessage(substr)
	}
}

// Queries the Type field in the InvolvedObject
// instance in the corev1.Event
// for an exact match of the argument string
func (event *Event) isType(typeVal string) bool {
	return event.Object.Type == typeVal
}

// Returns a predicate from isType() function
func IsType(typeVal string) Predicate {
	return func(event *Event) bool {
		return event.isType(typeVal)
	}
}
