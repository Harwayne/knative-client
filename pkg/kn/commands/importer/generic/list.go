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

package generic

import (
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewTriggerListCommand represents 'kn trigger list' command
func NewImporterListCOCommand(p *commands.KnParams) *cobra.Command {
	importerListCOFlags := NewImporterListFlags()

	importerListCOCommand := &cobra.Command{
		Use:   "list [CRD_NAME]",
		Short: "List available Importer Custom Objects.",
		Example: `
  # List all importer custom objects of kind
  # apiserversources.sources.eventing.knative.dev.
  kn importer generic list apiserversources.sources.eventing.knative.dev

  # List all importer custom objects of kind 'apiserversource'.
  # 'apiserversource' can be the name, kind, singular, or plural name of the
  # CRD. If there are multiple CRDs that match 'apiserversource', then an
  # error is returned.
  kn importer generic list apiserversource

  # List all apiserversource custom objects in JSON output format.
  kn importer generic list apiserversource -o json

  # List apiserversource custom object 'cli'.
  kn trigger list apiserversource cli`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires the CRD name")
			}
			crdName := args[0]

			ns, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			c, crd, err := GetCRD(p, crdName)
			if err != nil {
				return err
			}
			gvr := getGVR(crd)

			coList, err := c.Resource(gvr).Namespace(ns).List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(coList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			printer, err := importerListCOFlags.ToPrinter()
			if err != nil {
				return err
			}

			err = printer.PrintObj(coList, cmd.OutOrStdout())
			if err != nil {
				log.Error(2)
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(importerListCOCommand.Flags(), true)
	importerListCOFlags.AddFlags(importerListCOCommand)
	return importerListCOCommand
}
