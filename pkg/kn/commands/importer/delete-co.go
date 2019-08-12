// Copyright © 2018 The Knative Authors
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
	"errors"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
)

// NewImporterDeleteCOCommand represent 'importer delete-co' command
func NewImporterDeleteCOCommand(p *commands.KnParams) *cobra.Command {
	importerDeleteCommand := &cobra.Command{
		Use:   "delete-co CRD_NAME CO_NAME",
		Short: "Delete an importer custom object.",
		Example: `
  # Delete a importer 'svc1' in default namespace
  kn importer delete svc1

  # Delete a importer 'svc2' in 'ns1' namespace
  kn importer delete svc2 -n ns1`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("requires the CRD name and the CO name")
			}
			crdName := args[0]
			name := args[1]

			ns, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			c, crd, err := getCRD(p, crdName)
			if err != nil {
				return err
			}
			gvr := getGVR(crd)

			err = c.Resource(gvr).Namespace(ns).Delete(name, nil)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Importer '%s' of type '%s' successfully deleted in namespace '%s'.\n", name, crdName, ns)
			return nil
		},
	}
	commands.AddNamespaceFlags(importerDeleteCommand.Flags(), false)
	return importerDeleteCommand
}

func getGVR(crd v1beta1.CustomResourceDefinition) schema.GroupVersionResource {
	var version string
	for _, v := range crd.Spec.Versions {
		if v.Served {
			version = v.Name
			break
		}
	}
	if version == "" {
		version = crd.Spec.Version
	}

	return schema.GroupVersionResource{
		Group:    crd.Spec.Group,
		Version:  version,
		Resource: crd.Spec.Names.Plural,
	}
}

func getGVK(crd v1beta1.CustomResourceDefinition) schema.GroupVersionKind {
	gvr := getGVR(crd)
	return schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    crd.Spec.Names.Kind,
	}
}
