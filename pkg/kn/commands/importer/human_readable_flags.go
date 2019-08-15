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
	"github.com/knative/client/pkg/kn/commands"
	hprinters "github.com/knative/client/pkg/printers"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceListHandlers adds print handlers for service list command
func ImporterListHandlers(h hprinters.PrintHandler) {
	kServiceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the Knative service."},
		{Name: "Age", Type: "string", Description: "Age of the service."},
	}
	h.TableHandler(kServiceColumnDefinitions, printTrigger)
	h.TableHandler(kServiceColumnDefinitions, printTriggerList)
}

// Private functions

// printKServiceList populates the knative service list table rows
func printTriggerList(cl *v1beta1.CustomResourceDefinitionList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(cl.Items))
	for _, i := range cl.Items {
		r, err := printTrigger(&i, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printKService populates the knative service table rows
func printTrigger(c *v1beta1.CustomResourceDefinition, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
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
