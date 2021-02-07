/*
Copyright 2021 The KubeVirt Authors.

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

package v1beta1

import (
	"time"

	v1beta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"

	scheme "kubevirt.io/client-go/generated/external-snapshotter/clientset/versioned/scheme"
)

// VolumeSnapshotsGetter has a method to return a VolumeSnapshotInterface.
// A group's client should implement this interface.
type VolumeSnapshotsGetter interface {
	VolumeSnapshots(namespace string) VolumeSnapshotInterface
}

// VolumeSnapshotInterface has methods to work with VolumeSnapshot resources.
type VolumeSnapshotInterface interface {
	Create(*v1beta1.VolumeSnapshot) (*v1beta1.VolumeSnapshot, error)
	Update(*v1beta1.VolumeSnapshot) (*v1beta1.VolumeSnapshot, error)
	UpdateStatus(*v1beta1.VolumeSnapshot) (*v1beta1.VolumeSnapshot, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.VolumeSnapshot, error)
	List(opts v1.ListOptions) (*v1beta1.VolumeSnapshotList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.VolumeSnapshot, err error)
	VolumeSnapshotExpansion
}

// volumeSnapshots implements VolumeSnapshotInterface
type volumeSnapshots struct {
	client rest.Interface
	ns     string
}

// newVolumeSnapshots returns a VolumeSnapshots
func newVolumeSnapshots(c *SnapshotV1beta1Client, namespace string) *volumeSnapshots {
	return &volumeSnapshots{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the volumeSnapshot, and returns the corresponding volumeSnapshot object, and an error if there is any.
func (c *volumeSnapshots) Get(name string, options v1.GetOptions) (result *v1beta1.VolumeSnapshot, err error) {
	result = &v1beta1.VolumeSnapshot{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("volumesnapshots").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VolumeSnapshots that match those selectors.
func (c *volumeSnapshots) List(opts v1.ListOptions) (result *v1beta1.VolumeSnapshotList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.VolumeSnapshotList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("volumesnapshots").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested volumeSnapshots.
func (c *volumeSnapshots) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("volumesnapshots").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a volumeSnapshot and creates it.  Returns the server's representation of the volumeSnapshot, and an error, if there is any.
func (c *volumeSnapshots) Create(volumeSnapshot *v1beta1.VolumeSnapshot) (result *v1beta1.VolumeSnapshot, err error) {
	result = &v1beta1.VolumeSnapshot{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("volumesnapshots").
		Body(volumeSnapshot).
		Do().
		Into(result)
	return
}

// Update takes the representation of a volumeSnapshot and updates it. Returns the server's representation of the volumeSnapshot, and an error, if there is any.
func (c *volumeSnapshots) Update(volumeSnapshot *v1beta1.VolumeSnapshot) (result *v1beta1.VolumeSnapshot, err error) {
	result = &v1beta1.VolumeSnapshot{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("volumesnapshots").
		Name(volumeSnapshot.Name).
		Body(volumeSnapshot).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *volumeSnapshots) UpdateStatus(volumeSnapshot *v1beta1.VolumeSnapshot) (result *v1beta1.VolumeSnapshot, err error) {
	result = &v1beta1.VolumeSnapshot{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("volumesnapshots").
		Name(volumeSnapshot.Name).
		SubResource("status").
		Body(volumeSnapshot).
		Do().
		Into(result)
	return
}

// Delete takes name of the volumeSnapshot and deletes it. Returns an error if one occurs.
func (c *volumeSnapshots) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("volumesnapshots").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *volumeSnapshots) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("volumesnapshots").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched volumeSnapshot.
func (c *volumeSnapshots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.VolumeSnapshot, err error) {
	result = &v1beta1.VolumeSnapshot{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("volumesnapshots").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
