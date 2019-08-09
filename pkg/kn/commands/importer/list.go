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

	"k8s.io/apimachinery/pkg/labels"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
)

// NewTriggerListCommand represents 'kn trigger list' command
func NewImporterListCommand(p *commands.KnParams) *cobra.Command {
	triggerListFlags := NewTriggerListFlags()

	triggerListCommand := &cobra.Command{
		Use:   "list [name]",
		Short: "List available triggers.",
		Example: `
  # List all triggers
  kn trigger list

  # List all triggers in JSON output format
  kn trigger list -o json

  # List trigger 'web'
  kn trigger list web`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.NewDynamicClient()
			if err != nil {
				return err
			}
			crdList, err := getCRDInfo(args, client)
			if err != nil {
				return err
			}
			if len(crdList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			printer, err := triggerListFlags.ToPrinter()
			if err != nil {
				return err
			}

			err = printer.PrintObj(crdList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(triggerListCommand.Flags(), true)
	triggerListFlags.AddFlags(triggerListCommand)
	return triggerListCommand
}

func getCRDInfo(args []string, client dynamic.Interface) (*unstructured.UnstructuredList, error) {
	var (
		cl  *unstructured.UnstructuredList
		err error
	)
	switch len(args) {
	case 0:
		cl, err = client.Resource(crdGVK).List(v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"knative.dev/crd-install": "true",
			}).String(),
		})
	default:
		return nil, fmt.Errorf("'kn service list' accepts maximum 1 argument")
	}
	return cl, err
}
