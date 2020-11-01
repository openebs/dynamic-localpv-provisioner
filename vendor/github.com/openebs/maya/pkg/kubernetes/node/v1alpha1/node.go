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

import corev1 "k8s.io/api/core/v1"

const (
	kubeletReady = "KubeletReady"
)

// Node holds the api's node object
type Node struct {
	object *corev1.Node
}

// NodeList holds the list of node instances
type NodeList struct {
	items []*Node
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided node instance
type Predicate func(*Node) bool

// predicateList holds the list of predicates
type predicateList []Predicate

// Builder is the builder object for Pod
type Builder struct {
	Node *Node
}

// ListBuilder enables building an instance of NodeList
type ListBuilder struct {
	list    *NodeList
	filters predicateList
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{Node: &Node{object: &corev1.Node{}}}
}

// WithAPINode builds the node
// instances based on the provided node
func (b *Builder) WithAPINode(node *corev1.Node) *Builder {
	if node == nil {
		return b
	}
	return &Builder{Node: &Node{object: node}}

}

// WithAPIList builds the list of node
// instances based on the provided
// node list
func (b *ListBuilder) WithAPIList(nodes *corev1.NodeList) *ListBuilder {
	if nodes == nil {
		return b
	}
	b.WithAPIObject(nodes.Items...)
	return b
}

// WithObject builds the list of node instances based on the provided
// node list instance
func (b *ListBuilder) WithObject(nodes ...*Node) *ListBuilder {
	b.list.items = append(b.list.items, nodes...)
	return b
}

// WithAPIObject builds the list of node instances based on node api instances
func (b *ListBuilder) WithAPIObject(nodes ...corev1.Node) *ListBuilder {
	for _, n := range nodes {
		n := n
		b.list.items = append(b.list.items, &Node{&n})
	}
	return b
}

// List returns the list of node instances that was built by this builder
func (b *ListBuilder) List() *NodeList {
	if b.filters == nil && len(b.filters) == 0 {
		return b.list
	}
	filtered := &NodeList{}
	for _, node := range b.list.items {
		node := node // Pin it
		if b.filters.all(node) {
			filtered.items = append(filtered.items, node)
		}
	}
	return filtered
}

// NewListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &NodeList{items: []*Node{}}}
}

// ToAPIList converts NodeList to API NodeList
func (n *NodeList) ToAPIList() *corev1.NodeList {
	nlist := &corev1.NodeList{}
	for _, node := range n.items {
		nlist.Items = append(nlist.Items, *node.object)
	}
	return nlist
}

// all returns true if all the predicateList
// succeed against the provided node
// instance
func (l predicateList) all(n *Node) bool {
	for _, pred := range l {
		if !pred(n) {
			return false
		}
	}
	return true
}

// IsReady retuns true if the node is in ready state
func (n *Node) IsReady() bool {
	for _, nodeCond := range n.object.Status.Conditions {
		if nodeCond.Reason == kubeletReady && nodeCond.Type == corev1.NodeReady {
			return true
		}
	}
	return false
}

// IsReady is a Predicate to filter out nodes which are in running state
func IsReady() Predicate {
	return func(n *Node) bool {
		return n.IsReady()
	}
}

// WithFilter add filters on which the node has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// TODO:
// use type and have a contains func to check the taints
// before adding a taints
//
// type TaintList []corev1.Taint
//
// func (l TaintList) Contains(given corev1.Taint) bool {
//  for _, available := range l {
//    if available.MatchTaint(&given) {
//      return true
//     }
//   }
//  return false
// }

// WithTaints add taints to the Node resource
func (b *Builder) WithTaints(taintsToAdd []corev1.Taint) *Builder {
	newTaints := append([]corev1.Taint{}, taintsToAdd...)
	oldTaints := b.Node.object.Spec.Taints

	for _, oldTaint := range oldTaints {
		oldTaint := oldTaint
		existsInNew := false
		for _, taint := range newTaints {
			taint := taint
			if taint.MatchTaint(&oldTaint) {
				existsInNew = true
				break
			}
		}
		if !existsInNew {
			newTaints = append(newTaints, oldTaint)
		}
	}
	b.Node.object.Spec.Taints = newTaints
	return b
}

// Len returns the number of items present in the NodeList
func (n *NodeList) Len() int {
	return len(n.items)
}
