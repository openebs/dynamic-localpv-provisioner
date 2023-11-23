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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func fakeNodeListEmpty(noOfNodes int) *corev1.NodeList {
	list := &corev1.NodeList{}
	for i := 0; i < noOfNodes; i++ {
		node := corev1.Node{}
		list.Items = append(list.Items, node)
	}
	return list
}

func fakeNodeListAPI(nodeNames []string) *corev1.NodeList {
	if len(nodeNames) == 0 {
		return nil
	}
	list := &corev1.NodeList{}
	for _, name := range nodeNames {
		node := corev1.Node{}
		node.SetName(name)
		list.Items = append(list.Items, node)
	}
	return list
}

func fakeNodeInstances(nodes map[string]corev1.NodeConditionType) []*Node {
	nlist := []*Node{}
	for k := range nodes {
		n := &corev1.Node{}
		fakeNodeCondition := corev1.NodeCondition{
			Reason: kubeletReady,
			Type:   nodes[k],
		}
		n.SetName(k)
		n.Status.Conditions = append(n.Status.Conditions, fakeNodeCondition)
		nlist = append(nlist, &Node{n})
	}
	return nlist
}

func TestListBuilderFuncWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availableNodes    []string
		expectedNodeCount int
	}{
		"Node set 1":  {[]string{}, 0},
		"Node set 2":  {[]string{"node1"}, 1},
		"Node set 3":  {[]string{"node1", "node2"}, 2},
		"Node set 4":  {[]string{"node1", "node2", "node3"}, 3},
		"Node set 5":  {[]string{"node1", "node2", "node3", "node4"}, 4},
		"Node set 6":  {[]string{"node1", "node2", "node3", "node4", "node5"}, 5},
		"Node set 7":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6"}, 6},
		"Node set 8":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7"}, 7},
		"Node set 9":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8"}, 8},
		"Node set 10": {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8", "node9"}, 9},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeNodeListAPI(mock.availableNodes))
			if mock.expectedNodeCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderFuncWithEmptyNodeList(t *testing.T) {
	tests := map[string]struct {
		nodeCount, expectedNodeCount int
	}{
		"Two nodes": {5, 5},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeNodeListEmpty(mock.nodeCount))
			if mock.expectedNodeCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availableNodes    []string
		expectedNodeCount int
	}{
		"Node set 2":  {[]string{"node1"}, 1},
		"Node set 3":  {[]string{"node1", "node2"}, 2},
		"Node set 4":  {[]string{"node1", "node2", "node3"}, 3},
		"Node set 5":  {[]string{"node1", "node2", "node3", "node4"}, 4},
		"Node set 6":  {[]string{"node1", "node2", "node3", "node4", "node5"}, 5},
		"Node set 7":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6"}, 6},
		"Node set 8":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7"}, 7},
		"Node set 9":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8"}, 8},
		"Node set 10": {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8", "node9"}, 9},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIObject(fakeNodeListAPI(mock.availableNodes).Items...)
			if mock.expectedNodeCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availableNodes    []string
		expectedNodeCount int
	}{
		"Node set 1":  {[]string{}, 0},
		"Node set 2":  {[]string{"node1"}, 1},
		"Node set 3":  {[]string{"node1", "node2"}, 2},
		"Node set 4":  {[]string{"node1", "node2", "node3"}, 3},
		"Node set 5":  {[]string{"node1", "node2", "node3", "node4"}, 4},
		"Node set 6":  {[]string{"node1", "node2", "node3", "node4", "node5"}, 5},
		"Node set 7":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6"}, 6},
		"Node set 8":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7"}, 7},
		"Node set 9":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8"}, 8},
		"Node set 10": {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8", "node9"}, 9},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeNodeListAPI(mock.availableNodes)).List().ToAPIList()
			if mock.expectedNodeCount != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeCount, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availableNodes map[string]corev1.NodeConditionType
		filteredNodes  []string
		filters        predicateList
	}{
		"Nodes Set 1": {
			availableNodes: map[string]corev1.NodeConditionType{"Node 1": corev1.NodeReady},
			filteredNodes:  []string{"Node 1"},
			filters:        predicateList{IsReady()},
		},
		"Nodes Set 2": {
			availableNodes: map[string]corev1.NodeConditionType{"Node 1": corev1.NodeReady, "Node 2": corev1.NodeReady},
			filteredNodes:  []string{"Node 1", "Node 2"},
			filters:        predicateList{IsReady()},
		},
		"Nodes Set 3": {
			availableNodes: map[string]corev1.NodeConditionType{"Node 1": corev1.NodeDiskPressure, "Node 2": corev1.NodeMemoryPressure},
			filteredNodes:  []string{},
			filters:        predicateList{IsReady()},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			nl := &ListBuilder{list: &NodeList{items: fakeNodeInstances(mock.availableNodes)}, filters: mock.filters}
			list := nl.List()
			if len(list.items) != len(mock.filteredNodes) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredNodes), len(list.items))
			}
		})
	}
}

func TestWithTaints(t *testing.T) {
	node := &corev1.Node{
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	cases := []struct {
		name           string
		taintsToAdd    []corev1.Taint
		expectedTaints []corev1.Taint
		expectedErr    bool
	}{
		{
			name:           "no changes with taints",
			taintsToAdd:    []corev1.Taint{},
			expectedTaints: node.Spec.Taints,
			expectedErr:    false,
		},
		{
			name: "add new taint",
			taintsToAdd: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			expectedTaints: append([]corev1.Taint{{Key: "foo_1",
				Effect: corev1.TaintEffectNoExecute}},
				node.Spec.Taints...),
			expectedErr: false,
		},
	}

	for _, c := range cases {
		b := NewBuilder().WithAPINode(node).WithTaints(c.taintsToAdd)
		if !reflect.DeepEqual(c.expectedTaints, b.Node.object.Spec.Taints) {
			t.Errorf("[%s] expect to see taint list %#v, but got: %#v",
				c.name, c.expectedTaints,
				b.Node.object.Spec.Taints)
		}
	}
}
