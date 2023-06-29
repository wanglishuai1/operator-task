/*
Copyright The Kubernetes Authors.

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

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "github.com/shenyisyn/myci/pkg/apis/task/v1alpha1"
	taskv1alpha1 "github.com/shenyisyn/myci/pkg/client/applyconfiguration/task/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTasks implements TaskInterface
type FakeTasks struct {
	Fake *FakeApiV1alpha1
	ns   string
}

var tasksResource = v1alpha1.SchemeGroupVersion.WithResource("tasks")

var tasksKind = v1alpha1.SchemeGroupVersion.WithKind("Task")

// Get takes name of the task, and returns the corresponding task object, and an error if there is any.
func (c *FakeTasks) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Task, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(tasksResource, c.ns, name), &v1alpha1.Task{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Task), err
}

// List takes label and field selectors, and returns the list of Tasks that match those selectors.
func (c *FakeTasks) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TaskList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(tasksResource, tasksKind, c.ns, opts), &v1alpha1.TaskList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.TaskList{ListMeta: obj.(*v1alpha1.TaskList).ListMeta}
	for _, item := range obj.(*v1alpha1.TaskList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested tasks.
func (c *FakeTasks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(tasksResource, c.ns, opts))

}

// Create takes the representation of a task and creates it.  Returns the server's representation of the task, and an error, if there is any.
func (c *FakeTasks) Create(ctx context.Context, task *v1alpha1.Task, opts v1.CreateOptions) (result *v1alpha1.Task, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(tasksResource, c.ns, task), &v1alpha1.Task{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Task), err
}

// Update takes the representation of a task and updates it. Returns the server's representation of the task, and an error, if there is any.
func (c *FakeTasks) Update(ctx context.Context, task *v1alpha1.Task, opts v1.UpdateOptions) (result *v1alpha1.Task, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(tasksResource, c.ns, task), &v1alpha1.Task{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Task), err
}

// Delete takes name of the task and deletes it. Returns an error if one occurs.
func (c *FakeTasks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(tasksResource, c.ns, name, opts), &v1alpha1.Task{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTasks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(tasksResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.TaskList{})
	return err
}

// Patch applies the patch and returns the patched task.
func (c *FakeTasks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Task, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(tasksResource, c.ns, name, pt, data, subresources...), &v1alpha1.Task{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Task), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied task.
func (c *FakeTasks) Apply(ctx context.Context, task *taskv1alpha1.TaskApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Task, err error) {
	if task == nil {
		return nil, fmt.Errorf("task provided to Apply must not be nil")
	}
	data, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	name := task.Name
	if name == nil {
		return nil, fmt.Errorf("task.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(tasksResource, c.ns, *name, types.ApplyPatchType, data), &v1alpha1.Task{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Task), err
}
