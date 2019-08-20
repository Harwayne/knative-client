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

package trigger

import (
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
)

// NewTriggerDeleteCommand represent 'trigger delete' command
func NewTriggerDeleteCommand(p *commands.KnParams) *cobra.Command {
	triggerDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a trigger.",
		Example: `
  # Delete trigger 't1' in the default namespace.
  kn trigger delete t1

  # Delete trigger 't2' in the 'ns1' namespace.
  kn trigger delete t2 -n ns1`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires the trigger name")
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			err = client.DeleteTrigger(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Trigger '%s' successfully deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(triggerDeleteCommand.Flags(), false)
	return triggerDeleteCommand
}
