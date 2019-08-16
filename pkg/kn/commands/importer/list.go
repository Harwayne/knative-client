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

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/kn/commands/importer/generic"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

// NewTriggerListCommand represents 'kn trigger list' command
func NewImporterListCommand(p *commands.KnParams) *cobra.Command {
	triggerListFlags := generic.NewImporterListFlags()

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
			if len(args) != 0 {
				return fmt.Errorf("'kn service list' accepts maximum 1 argument")
			}
			client, err := p.NewDynamicClient()
			if err != nil {
				return err
			}
			crdList, err := generic.ListImporterCRDs(client)
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

			err = printer.PrintObj(&crdList, cmd.OutOrStdout())
			if err != nil {
				log.Error(2)
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(triggerListCommand.Flags(), true)
	triggerListFlags.AddFlags(triggerListCommand)
	return triggerListCommand
}
