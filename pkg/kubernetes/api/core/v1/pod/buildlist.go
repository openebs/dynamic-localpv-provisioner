package pod

import (
	corev1 "k8s.io/api/core/v1"
)

// ListBuilder enables building an instance of
// Podlist
type ListBuilder struct {
	list    *PodList
	filters PredicateList
}

// NewListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &PodList{items: []*Pod{}}}
}

// ListBuilderForAPIList returns a instance of ListBuilder from API PodList
func ListBuilderForAPIList(pods *corev1.PodList) *ListBuilder {
	b := &ListBuilder{list: &PodList{}}
	if pods == nil {
		return b
	}
	for _, p := range pods.Items {
		p := p
		b.list.items = append(b.list.items, &Pod{object: &p})
	}
	return b
}

// ListBuilderForObjectList returns a instance of ListBuilder from API Pods
func ListBuilderForObjectList(pods ...*Pod) *ListBuilder {
	b := &ListBuilder{list: &PodList{}}
	if pods == nil {
		return b
	}
	for _, p := range pods {
		p := p
		b.list.items = append(b.list.items, p)
	}
	return b
}

// List returns the list of pod
// instances that was built by this
// builder
func (b *ListBuilder) List() *PodList {
	if len(b.filters) == 0 {
		return b.list
	}
	filtered := &PodList{}
	for _, pod := range b.list.items {
		if b.filters.all(pod) {
			filtered.items = append(filtered.items, pod)
		}
	}
	return filtered
}

// WithFilter add filters on which the pod
// has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
