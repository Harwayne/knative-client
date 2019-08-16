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

package generic

import (
	"github.com/knative/client/pkg/kn/commands"
	hprinters "github.com/knative/client/pkg/printers"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceListHandlers adds print handlers for service list command
func ImporterListHandlers(h hprinters.PrintHandler) {
	crdColumns := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the Knative service."},
		{Name: "Age", Type: "string", Description: "Age of the service."},
	}
	h.TableHandler(crdColumns, printCRD)
	h.TableHandler(crdColumns, printCRDList)

	uColumns := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the Knative service."},
		{Name: "Age", Type: "string", Description: "Age of the service."},
		{Name: "Ready", Type: "string", Description: "Readiness of the CO, if found."},
		{Name: "Reason", Type: "string", Description: "Explanation of readiness."},
	}
	h.TableHandler(uColumns, printUnstructured)
	h.TableHandler(uColumns, printUnstructuredList)
}

// Private functions

// printKServiceList populates the knative service list table rows
func printCRDList(cl *v1beta1.CustomResourceDefinitionList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(cl.Items))
	for _, i := range cl.Items {
		r, err := printCRD(&i, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printKService populates the knative service table rows
func printCRD(c *v1beta1.CustomResourceDefinition, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := c.GetName()
	age := commands.TranslateTimestampSince(c.GetCreationTimestamp())

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: c},
	}
	row.Cells = append(row.Cells,
		name,
		age)
	return []metav1beta1.TableRow{row}, nil
}

// printKServiceList populates the knative service list table rows
func printUnstructuredList(ul *unstructured.UnstructuredList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(ul.Items))
	for _, i := range ul.Items {
		r, err := printUnstructured(&i, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printKService populates the knative service table rows
func printUnstructured(u *unstructured.Unstructured, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := u.GetName()
	age := commands.TranslateTimestampSince(u.GetCreationTimestamp())

	var ready string
	var reason string
	readyCond, err := extractReadyCondition(*u)
	if err == nil {
		// Note this is the opposite of normal Golang error handling. This only happens if there
		// isn't an error.
		ready = string(readyCond.Status)
		reason = readyCond.Reason
	}

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: u},
	}
	row.Cells = append(row.Cells,
		name,
		age,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
