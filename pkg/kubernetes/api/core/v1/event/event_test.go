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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsBdcEventPredicate(t *testing.T) {
	tests := map[string]struct {
		availableInvolvedObjectKind string
		isBDCKind                   bool
	}{
		"Test1": {"BlockDeviceClaim", true},
		"Test2": {"Pod", false},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeEvent := &Event{Object: &corev1.Event{InvolvedObject: corev1.ObjectReference{Kind: test.availableInvolvedObjectKind}}}
			ok := IsBdcEvent()(fakeEvent)
			if ok != test.isBDCKind {
				t.Fatalf("Test %v failed, Expected %s but got %v", name, "BlockDeviceClaim", fakeEvent.Object.InvolvedObject.Kind)
			}
		})
	}
}

func TestIsBdEventPredicate(t *testing.T) {
	tests := map[string]struct {
		availableInvolvedObjectKind string
		isBDKind                    bool
	}{
		"Test1": {"BlockDevice", true},
		"Test2": {"Pod", false},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeEvent := &Event{Object: &corev1.Event{InvolvedObject: corev1.ObjectReference{Kind: test.availableInvolvedObjectKind}}}
			ok := IsBdEvent()(fakeEvent)
			if ok != test.isBDKind {
				t.Fatalf("Test %v failed, Expected %s but got %v", name, "BlockDevice", fakeEvent.Object.InvolvedObject.Kind)
			}
		})
	}
}

func TestIsPodEventPredicate(t *testing.T) {
	tests := map[string]struct {
		availableInvolvedObjectKind string
		isPodKind                   bool
	}{
		"Test1": {"BlockDeviceClaim", false},
		"Test2": {"Pod", true},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeEvent := &Event{Object: &corev1.Event{InvolvedObject: corev1.ObjectReference{Kind: test.availableInvolvedObjectKind}}}
			ok := IsPodEvent()(fakeEvent)
			if ok != test.isPodKind {
				t.Fatalf("Test %v failed, Expected %s but got %v", name, "Pod", fakeEvent.Object.InvolvedObject.Kind)
			}
		})
	}
}

func TestIsTypePredicate(t *testing.T) {
	tests := map[string]struct {
		availableType string
		checkForType  string
		isType        bool
	}{
		"Test1": {"Normal", "Normal", true},
		"Test2": {"Normal", "Warning", false},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeEvent := &Event{Object: &corev1.Event{Type: test.availableType}}
			ok := IsType(test.checkForType)(fakeEvent)
			if ok != test.isType {
				t.Fatalf("Test %v failed, Expected %s but got %v", name, test.checkForType, fakeEvent.Object.Type)
			}
		})
	}
}

func TestHasReasonPredicate(t *testing.T) {
	tests := map[string]struct {
		availableReason string
		checkForReason  string
		hasReason       bool
	}{
		"Test1": {"BlockDeviceReleased", "BlockDeviceReleased", true},
		"Test2": {"BlockDeviceClaimBound", "BlockDeviceClaimed", false},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeEvent := &Event{&corev1.Event{Reason: test.availableReason}}
			ok := HasReason(test.checkForReason)(fakeEvent)
			if ok != test.hasReason {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableReason, fakeEvent.Object.Reason)
			}
		})
	}
}

func TestStringInMessagePredicate(t *testing.T) {
	tests := map[string]struct {
		availableMessage string
		checkForString   string
		hasString        bool
	}{
		"Test1": {"Released from BDC: bdc-pvc-080b002e-c345-4a3d-88ab-324c41d41819", "bdc-pvc-080b002e-c345-4a3d-88ab-324c41d41819", true},
		"Test2": {"Released from BDC: bdc-pvc-080b002e-c345-4a3d-88ab-324c41d42000", "bdc-pvc-080b002e-c345-4a3d-88ab-324c41d41819", false},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeEvent := &Event{&corev1.Event{Message: test.availableMessage}}
			ok := HasStringInMessage(test.checkForString)(fakeEvent)
			if ok != test.hasString {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableMessage, fakeEvent.Object.Message)
			}
		})
	}
}

func TestLatestFirstSort(t *testing.T) {
	tests := map[string]struct {
		availableEvents []*Event
		sortedOrder     []string
	}{
		"Test1": {
			availableEvents: []*Event{
				{
					Object: &corev1.Event{
						ObjectMeta:    metav1.ObjectMeta{Name: "Event 1"},
						LastTimestamp: metav1.Unix(500, 0),
					},
				},
				{
					Object: &corev1.Event{
						ObjectMeta:    metav1.ObjectMeta{Name: "Event 2"},
						LastTimestamp: metav1.Unix(600, 0),
					},
				},
				{
					Object: &corev1.Event{
						ObjectMeta:    metav1.ObjectMeta{Name: "Event 3"},
						LastTimestamp: metav1.Unix(400, 0),
					},
				},
			},
			sortedOrder: []string{"Event 2", "Event 1", "Event 3"},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			availableEventList := &EventList{Items: test.availableEvents}
			resultEventList := availableEventList.LatestFirstSort()
			var resultOrder []string
			for _, event := range resultEventList.Items {
				resultOrder = append(resultOrder, event.Object.Name)
			}
			if !reflect.DeepEqual(resultOrder, test.sortedOrder) {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.sortedOrder, resultOrder)
			}
		})
	}
}

func TestLatestLastSort(t *testing.T) {
	tests := map[string]struct {
		availableEvents []*Event
		sortedOrder     []string
	}{
		"Test1": {
			availableEvents: []*Event{
				{
					Object: &corev1.Event{
						ObjectMeta:    metav1.ObjectMeta{Name: "Event 1"},
						LastTimestamp: metav1.Unix(500, 0),
					},
				},
				{
					Object: &corev1.Event{
						ObjectMeta:    metav1.ObjectMeta{Name: "Event 2"},
						LastTimestamp: metav1.Unix(600, 0),
					},
				},
				{
					Object: &corev1.Event{
						ObjectMeta:    metav1.ObjectMeta{Name: "Event 3"},
						LastTimestamp: metav1.Unix(400, 0),
					},
				},
			},
			sortedOrder: []string{"Event 3", "Event 1", "Event 2"},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			availableEventList := &EventList{Items: test.availableEvents}
			resultEventList := availableEventList.LatestLastSort()
			var resultOrder []string
			for _, event := range resultEventList.Items {
				resultOrder = append(resultOrder, event.Object.Name)
			}
			if !reflect.DeepEqual(resultOrder, test.sortedOrder) {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.sortedOrder, resultOrder)
			}
		})
	}
}
