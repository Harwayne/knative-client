// Copyright © 2019 The Knative Authors
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
	"encoding/json"
	"fmt"
	"time"

	"github.com/knative/client/pkg/eventing"
	"github.com/knative/client/pkg/wait"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	client_v1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	"github.com/knative/pkg/apis"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis/duck/v1beta1"
)

// Kn interface to Eventing. All methods are relative to the
// namespace specified during construction
type KnClient interface {
	GetTrigger(name string) (*v1alpha1.Trigger, error)

	ListTriggers(opts ...ListConfig) (*v1alpha1.TriggerList, error)

	// Create a new service
	CreateTrigger(service *v1alpha1.Trigger) error

	// Update the given service
	UpdateTrigger(service *v1alpha1.Trigger) error

	// Delete a service by name
	DeleteTrigger(name string) error

	// Wait for a service to become ready, but not longer than provided timeout
	WaitForTrigger(name string, timeout time.Duration) error

	// Wait for a service to become ready, but not longer than provided timeout
	WaitForBroker(name string, timeout time.Duration) error
}

type listConfigCollector struct {
	// Labels to filter on
	Labels labels.Set

	// Fields to filter on
	Fields fields.Set
}

// Config function for builder pattern
type ListConfig func(config *listConfigCollector)

type ListConfigs []ListConfig

// add selectors to a list options
func (opts ListConfigs) toListOptions() v1.ListOptions {
	listConfig := listConfigCollector{labels.Set{}, fields.Set{}}
	for _, f := range opts {
		f(&listConfig)
	}
	options := v1.ListOptions{}
	if len(listConfig.Fields) > 0 {
		options.FieldSelector = listConfig.Fields.String()
	}
	if len(listConfig.Labels) > 0 {
		options.LabelSelector = listConfig.Labels.String()
	}
	return options
}

// Filter list on the provided name
func WithName(name string) ListConfig {
	return func(lo *listConfigCollector) {
		lo.Fields["metadata.name"] = name
	}
}

// // Filter on the service name
// func WithTrigger(trigger string) ListConfig {
// 	return func(lo *listConfigCollector) {
// 		lo.Labels[api_eventing.TriggerLabelKey] = trigger
// 	}
// }

type kneClient struct {
	client    client_v1alpha1.EventingV1alpha1Interface
	namespace string
}

// Create a new client facade for the provided namespace
func NewKnEventingClient(client client_v1alpha1.EventingV1alpha1Interface, namespace string) KnClient {
	return &kneClient{
		client:    client,
		namespace: namespace,
	}
}

// Get a trigger by its unique name
func (cl *kneClient) GetTrigger(name string) (*v1alpha1.Trigger, error) {
	trigger, err := cl.client.Triggers(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = eventing.UpdateGroupVersionKind(trigger, v1alpha1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

// List triggers
func (cl *kneClient) ListTriggers(config ...ListConfig) (*v1alpha1.TriggerList, error) {
	triggerList, err := cl.client.Triggers(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, err
	}
	triggerListNew := triggerList.DeepCopy()
	err = updateEventingGVK(triggerListNew)
	if err != nil {
		return nil, err
	}

	triggerListNew.Items = make([]v1alpha1.Trigger, len(triggerList.Items))
	for idx, trigger := range triggerList.Items {
		triggerClone := trigger.DeepCopy()
		err := updateEventingGVK(triggerClone)
		if err != nil {
			return nil, err
		}
		triggerListNew.Items[idx] = *triggerClone
	}
	return triggerListNew, nil
}

// Create a new trigger
func (cl *kneClient) CreateTrigger(trigger *v1alpha1.Trigger) error {
	_, err := cl.client.Triggers(cl.namespace).Create(trigger)
	if err != nil {
		return err
	}
	return updateEventingGVK(trigger)
}

// Update the given trigger
func (cl *kneClient) UpdateTrigger(trigger *v1alpha1.Trigger) error {
	_, err := cl.client.Triggers(cl.namespace).Update(trigger)
	if err != nil {
		return err
	}
	return updateEventingGVK(trigger)
}

// Delete a trigger by name
func (cl *kneClient) DeleteTrigger(triggerName string) error {
	return cl.client.Triggers(cl.namespace).Delete(
		triggerName,
		&v1.DeleteOptions{},
	)
}

// Wait for a trigger to become ready, but not longer than provided timeout
func (cl *kneClient) WaitForTrigger(name string, timeout time.Duration) error {
	waitForReady := newTriggerWaitForReady(cl.client.Triggers(cl.namespace).Watch)
	return waitForReady.Wait(name, timeout)
}

// Wait for a Broker to become ready, but not longer than provided timeout
func (cl *kneClient) WaitForBroker(name string, timeout time.Duration) error {
	waitForReady := newBrokerWaitForReady(cl.client.Brokers(cl.namespace).Watch)
	return waitForReady.Wait(name, timeout)
}

// update with the v1alpha1 group + version
func updateEventingGVK(obj runtime.Object) error {
	return eventing.UpdateGroupVersionKind(obj, v1alpha1.SchemeGroupVersion)
}

// Create wait arguments for a Knative service which can be used to wait for
// a create/update options to be finished
// Can be used by `service_create` and `service_update`, hence this extra file
func newTriggerWaitForReady(watch wait.WatchFunc) wait.WaitForReady {
	return wait.NewWaitForReadyIgnoreGeneration(
		"trigger",
		watch,
		triggerConditionExtractor)
}


func triggerConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
	t, ok := obj.(*v1alpha1.Trigger)
	if !ok {
		return nil, fmt.Errorf("%v is not a trigger", obj)
	}
	return ToServingConditions(t.Status.Conditions)
}
// Create wait arguments for a Knative service which can be used to wait for
// a create/update options to be finished
// Can be used by `service_create` and `service_update`, hence this extra file
func newBrokerWaitForReady(watch wait.WatchFunc) wait.WaitForReady {
	return wait.NewWaitForReadyIgnoreGeneration(
		"broker",
		watch,
		brokerConditionExtractor)
}


func brokerConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
	b, ok := obj.(*v1alpha1.Broker)
	if !ok {
		return nil, fmt.Errorf("%v is not a trigger", obj)
	}
	return ToServingConditions(b.Status.Conditions)
}

func ToServingConditions(conditions v1beta1.Conditions) (apis.Conditions, error) {
	// TODO do this right.
	j, err := json.Marshal(conditions)
	if err != nil {
		return nil, err
	}
	c := apis.Conditions{}
	err = json.Unmarshal(j, &c)
	return c, err
}
