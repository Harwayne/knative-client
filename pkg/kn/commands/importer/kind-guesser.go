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
	"fmt"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/dynamic"
)

func GuessCRDFromKind(client dynamic.Interface, kind string) (v1beta1.CustomResourceDefinition, error) {
	crdL, err := listImporterCRDs(client)
	if err != nil {
		return v1beta1.CustomResourceDefinition{}, err
	}
	kind = strings.ToLower(kind)
	found := make([]v1beta1.CustomResourceDefinition, 0, len(crdL.Items))
	for _, crd := range crdL.Items {
		for _, n := range potentialNames(crd) {
			if strings.ToLower(n) == kind {
				found = append(found, crd)
				break
			}
		}
	}
	if len(found) == 0 {
		return v1beta1.CustomResourceDefinition{}, fmt.Errorf("no Importer CRD found with kind %q", kind)
	} else if len(found) == 1 {
		return *found[0].DeepCopy(), nil
	}
	return v1beta1.CustomResourceDefinition{}, fmt.Errorf("multiple matching Importer CRDs found with kind %q, use the full CRD name instead", kind)
}

func potentialNames(crd v1beta1.CustomResourceDefinition) []string {
	n := []string{
		crd.Spec.Names.Kind,
		crd.Spec.Names.Plural,
		crd.Spec.Names.Singular,
	}
	n = append(n, crd.Spec.Names.ShortNames...)
	return n
}
