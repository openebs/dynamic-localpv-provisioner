/*
Copyright 2019 The OpenEBS Authors

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	scheme "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// UpgradeResultsGetter has a method to return a UpgradeResultInterface.
// A group's client should implement this interface.
type UpgradeResultsGetter interface {
	UpgradeResults(namespace string) UpgradeResultInterface
}

// UpgradeResultInterface has methods to work with UpgradeResult resources.
type UpgradeResultInterface interface {
	Create(ctx context.Context, upgradeResult *v1alpha1.UpgradeResult, opts v1.CreateOptions) (*v1alpha1.UpgradeResult, error)
	Update(ctx context.Context, upgradeResult *v1alpha1.UpgradeResult, opts v1.UpdateOptions) (*v1alpha1.UpgradeResult, error)
	UpdateStatus(ctx context.Context, upgradeResult *v1alpha1.UpgradeResult, opts v1.UpdateOptions) (*v1alpha1.UpgradeResult, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.UpgradeResult, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.UpgradeResultList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.UpgradeResult, err error)
	UpgradeResultExpansion
}

// upgradeResults implements UpgradeResultInterface
type upgradeResults struct {
	client rest.Interface
	ns     string
}

// newUpgradeResults returns a UpgradeResults
func newUpgradeResults(c *OpenebsV1alpha1Client, namespace string) *upgradeResults {
	return &upgradeResults{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the upgradeResult, and returns the corresponding upgradeResult object, and an error if there is any.
func (c *upgradeResults) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.UpgradeResult, err error) {
	result = &v1alpha1.UpgradeResult{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("upgraderesults").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of UpgradeResults that match those selectors.
func (c *upgradeResults) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.UpgradeResultList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.UpgradeResultList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("upgraderesults").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested upgradeResults.
func (c *upgradeResults) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("upgraderesults").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a upgradeResult and creates it.  Returns the server's representation of the upgradeResult, and an error, if there is any.
func (c *upgradeResults) Create(ctx context.Context, upgradeResult *v1alpha1.UpgradeResult, opts v1.CreateOptions) (result *v1alpha1.UpgradeResult, err error) {
	result = &v1alpha1.UpgradeResult{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("upgraderesults").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(upgradeResult).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a upgradeResult and updates it. Returns the server's representation of the upgradeResult, and an error, if there is any.
func (c *upgradeResults) Update(ctx context.Context, upgradeResult *v1alpha1.UpgradeResult, opts v1.UpdateOptions) (result *v1alpha1.UpgradeResult, err error) {
	result = &v1alpha1.UpgradeResult{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("upgraderesults").
		Name(upgradeResult.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(upgradeResult).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *upgradeResults) UpdateStatus(ctx context.Context, upgradeResult *v1alpha1.UpgradeResult, opts v1.UpdateOptions) (result *v1alpha1.UpgradeResult, err error) {
	result = &v1alpha1.UpgradeResult{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("upgraderesults").
		Name(upgradeResult.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(upgradeResult).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the upgradeResult and deletes it. Returns an error if one occurs.
func (c *upgradeResults) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("upgraderesults").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *upgradeResults) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("upgraderesults").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched upgradeResult.
func (c *upgradeResults) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.UpgradeResult, err error) {
	result = &v1alpha1.UpgradeResult{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("upgraderesults").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
