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

package importer

import (
	"encoding/json"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	hprinters "github.com/knative/client/pkg/printers"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis/duck/v1beta1"
)

// ServiceListHandlers adds print handlers for service list command
func ServiceListHandlers(h hprinters.PrintHandler) {
	kServiceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the Knative service."},
		{Name: "Generation", Type: "integer", Description: "Sequence number of 'Generation' of the service that was last processed by the controller."},
		{Name: "Age", Type: "string", Description: "Age of the service."},
		{Name: "Conditions", Type: "string", Description: "Conditions describing statuses of service components."},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the service."},
		{Name: "Reason", Type: "string", Description: "Reason for non-ready condition of the service."},
		{Name: "Subscriber", Type: "string", Description: "Subscriber's URL."},
	}
	h.TableHandler(kServiceColumnDefinitions, printTrigger)
	h.TableHandler(kServiceColumnDefinitions, printTriggerList)
}

// Private functions

// printKServiceList populates the knative service list table rows
func printTriggerList(kServiceList *eventingv1alpha1.TriggerList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(kServiceList.Items))
	for _, ksvc := range kServiceList.Items {
		r, err := printTrigger(&ksvc, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printKService populates the knative service table rows
func printTrigger(kService *eventingv1alpha1.Trigger, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := kService.Name
	generation := kService.Status.ObservedGeneration
	age := commands.TranslateTimestampSince(kService.CreationTimestamp)
	conditions := commands.ConditionsValue(toConditions(kService.Status.Conditions))
	ready := commands.ReadyCondition(toConditions(kService.Status.Conditions))
	reason := commands.NonReadyConditionReason(toConditions(kService.Status.Conditions))
	subscriberURI := kService.Status.SubscriberURI

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: kService},
	}
	row.Cells = append(row.Cells,
		name,
		generation,
		age,
		conditions,
		ready,
		reason,
		subscriberURI)
	return []metav1beta1.TableRow{row}, nil
}

func toConditions(conditions v1beta1.Conditions) duckv1beta1.Conditions {
	j, err := json.Marshal(conditions)
	if err != nil {
		panic(fmt.Errorf("marshalling json: %v", err))
	}
	c := duckv1beta1.Conditions{}
	err = json.Unmarshal(j, &c)
	if err != nil {
		panic(fmt.Errorf("unmarshalling json: %v", err))
	}
	return c
}
