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
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
)

func NewTriggerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags EditFlags
	var waitFlags commands.WaitFlags

	triggerUpdateCommand := &cobra.Command{
		Use:   "update NAME [flags]",
		Short: "Update a trigger.",
		Example: `
  # Updates a trigger 'mysvc' with new environment variables
  kn trigger update mysvc --env KEY1=VALUE1 --env KEY2=VALUE2

  # Update a trigger 'mysvc' with new port
  kn trigger update mysvc --port 80

  # Updates a trigger 'mysvc' with new requests and limits parameters
  kn trigger update mysvc --requests-cpu 500m --limits-memory 1024Mi`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the trigger name.")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			var retries = 0
			for {
				name := args[0]
				trigger, err := client.GetTrigger(name)
				if err != nil {
					return err
				}
				trigger = trigger.DeepCopy()

				err = editFlags.Apply(trigger, cmd)
				if err != nil {
					return err
				}

				_, err = client.UpdateTrigger(trigger)
				if err != nil {
					// Retry to update when a resource version conflict exists
					if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
						retries++
						continue
					}
					return err
				}

				if !waitFlags.Async {
					out := cmd.OutOrStdout()
					err := waitForTrigger(client, name, out, waitFlags.TimeoutInSeconds)
					if err != nil {
						return err
					}
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Trigger '%s' updated in namespace '%s'.\n", args[0], namespace)
				return nil
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preCheck(cmd, args)
		},
	}

	commands.AddNamespaceFlags(triggerUpdateCommand.Flags(), false)
	editFlags.AddUpdateFlags(triggerUpdateCommand)
	waitFlags.AddConditionWaitFlags(triggerUpdateCommand, 60, "Update", "trigger")
	return triggerUpdateCommand
}

func preCheck(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NFlag() == 0 {
		return errors.New(fmt.Sprintf("flag(s) not set\nUsage: %s", cmd.Use))
	}

	return nil
}
