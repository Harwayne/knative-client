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

package trigger

import (
	"fmt"

	eventing_kn_v1alpha1 "github.com/knative/client/pkg/eventing/v1alpha1"
	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/spf13/cobra"
)

// NewTriggerListCommand represents 'kn trigger list' command
func NewTriggerListCommand(p *commands.KnParams) *cobra.Command {
	triggerListFlags := NewTriggerListFlags()

	triggerListCommand := &cobra.Command{
		Use:   "list [name]",
		Short: "List available triggers.",
		Example: `
  # List all triggers.
  kn trigger list

  # List all triggers in JSON output format.
  kn trigger list -o json

  # List trigger 'web'.
  kn trigger list web`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}
			triggerList, err := getTriggerInfo(args, client)
			if err != nil {
				return err
			}
			if len(triggerList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			printer, err := triggerListFlags.ToPrinter()
			if err != nil {
				return err
			}

			err = printer.PrintObj(triggerList, cmd.OutOrStdout())
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

func getTriggerInfo(args []string, client eventing_kn_v1alpha1.KnClient) (*v1alpha1.TriggerList, error) {
	var (
		tl  *v1alpha1.TriggerList
		err error
	)
	switch len(args) {
	case 0:
		tl, err = client.ListTriggers()
	case 1:
		tl, err = client.ListTriggers(eventing_kn_v1alpha1.WithName(args[0]))
	default:
		return nil, fmt.Errorf("'kn service list' accepts maximum 1 argument")
	}
	return tl, err
}
