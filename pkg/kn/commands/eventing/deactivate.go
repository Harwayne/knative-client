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

package eventing

import (
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
)

func NewEventingDeactivateCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags

	eventingActivateCommand := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate Knative eventing in the given namespace.",
		Long: `
Deactivate Knative eventing in the given namespace. This does not delete
anything currently running. Rather it will stop reconciling that namespace. So
if anything is altered or deleted, nothing will put it back into a working
state.`,
		Example: `
  # Activate Knative eventing in the namespace that kn is associated with.
  kn eventing activate

  # Activate Knative eventing in the namespace 'foo'
  kn eventing activate --namespace foo`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("no argument are expected")
			}
			namespaceName, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewDynamicClient()
			if err != nil {
				return err
			}

			ns, err := getNamespace(client, namespaceName)
			if err != nil {
				return err
			}
			err = removeLabel(client, ns)
			if err != nil {
				return err
			}

			if !waitFlags.Async {
				eventingClient, err := p.NewEventingClient(ns.Name)
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				err = waitForBroker(eventingClient, "default", out, waitFlags.TimeoutInSeconds)
				if err != nil {
					return err
				}
				return nil
			}

			msg :=
				`Eventing has been deactivated in namespace %q. Nothing has been removed, so
anything already working will continue to work. However, if any pieces are
altered or deleted, they will not be fixed.`
			fmt.Fprintf(cmd.OutOrStdout(), msg, ns.Name)
			return nil
		},
	}
	commands.AddNamespaceFlags(eventingActivateCommand.Flags(), false)
	waitFlags.AddConditionWaitFlags(eventingActivateCommand, 60, "Activate", "Eventing")
	return eventingActivateCommand
}

func removeLabel(client dynamic.Interface, ns corev1.Namespace) error {
	if ns.Labels[labelKey] != enabledLabelValue {
		return nil
	}
	delete(ns.Labels, labelKey)
	return updateNS(client, ns)
}
