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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

// NewTriggerListCommand represents 'kn trigger list' command
func NewImporterListCOCommand(p *commands.KnParams) *cobra.Command {
	importerListCOFlags := NewImporterListFlags()

	importerListCOCommand := &cobra.Command{
		Use:   "list-co [name]",
		Short: "List available Importer Custom Objects.",
		Example: `
  # List all triggers
  kn trigger list

  # List all triggers in JSON output format
  kn trigger list -o json

  # List trigger 'web'
  kn trigger list web`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires the CRD name")
			}
			crdName := args[0]

			ns, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			c, crd, err := getCRD(p, crdName)
			if err != nil {
				return err
			}
			gvr := getGVR(crd)

			coList, err := c.Resource(gvr).Namespace(ns).List(v1.ListOptions{})
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
