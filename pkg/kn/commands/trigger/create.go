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
	"io"

	"github.com/knative/client/pkg/eventing/v1alpha1"
	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/eventing/pkg/apis/eventing"
	eventing_v1alpha1_api "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/spf13/cobra"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTriggerCreateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags triggerEditFlags
	var waitFlags commands.WaitFlags

	triggerCreateCommand := &cobra.Command{
		Use:   "create NAME --image IMAGE",
		Short: "Create a trigger.",
		Example: `
  # Create a trigger 'mysvc' using image at dev.local/ns/image:latest
  kn trigger create mysvc --image dev.local/ns/image:latest

  # Create a trigger with multiple environment variables
  kn trigger create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest

  # Create or replace a trigger 's1' with image dev.local/ns/image:v2 using --force flag
  # if trigger 's1' doesn't exist, it's just a normal create operation
  kn trigger create --force s1 --image dev.local/ns/image:v2

  # Create or replace environment variables of trigger 's1' using --force flag
  kn trigger create --force s1 --env KEY1=NEW_VALUE1 --env NEW_KEY2=NEW_VALUE2 --image dev.local/ns/image:v1

  # Create trigger 'mysvc' with port 80
  kn trigger create mysvc --port 80 --image dev.local/ns/image:latest

  # Create or replace default resources of a trigger 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn trigger create --force s1 --image dev.local/ns/image:v1`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'trigger create' requires the trigger name given as single argument")
			}
			name := args[0]
			if editFlags.SubscriberName == "" {
				return errors.New("'trigger create' requires the subscriber name to run provided with the --subscriber option")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			trigger, err := constructTrigger(cmd, editFlags, name, namespace)
			if err != nil {
				return err
			}

			client, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			triggerExists, err := triggerExists(client, name, namespace)
			if err != nil {
				return err
			}

			if triggerExists {
				if !editFlags.ForceCreate {
					return fmt.Errorf(
						"cannot create trigger '%s' in namespace '%s' "+
							"because the trigger already exists and no --force option was given", name, namespace)
				}
				err = replaceTrigger(client, trigger, namespace, cmd.OutOrStdout())
			} else {
				err = createTrigger(client, trigger, namespace, cmd.OutOrStdout())
			}
			if err != nil {
				return err
			}

			if !waitFlags.Async {
				out := cmd.OutOrStdout()
				err := waitForTrigger(client, name, out, waitFlags.TimeoutInSeconds)
				if err != nil {
					return err
				}
				return showUrl(client, name, namespace, out)
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(triggerCreateCommand.Flags(), false)
	editFlags.AddCreateFlags(triggerCreateCommand)
	waitFlags.AddConditionWaitFlags(triggerCreateCommand, 60, "Create", "trigger")
	return triggerCreateCommand
}

// Duck type for writers having a flush
type flusher interface {
	Flush() error
}

func flush(out io.Writer) {
	if flusher, ok := out.(flusher); ok {
		flusher.Flush()
	}
}

func createTrigger(client v1alpha1.KnClient, trigger *eventing_v1alpha1_api.Trigger, namespace string, out io.Writer) error {
	err := client.CreateTrigger(trigger)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Trigger '%s' successfully created in namespace '%s'.\n", trigger.Name, namespace)
	return nil
}

func replaceTrigger(client v1alpha1.KnClient, trigger *eventing_v1alpha1_api.Trigger, namespace string, out io.Writer) error {
	var retries = 0
	for {
		existingTrigger, err := client.GetTrigger(trigger.Name)
		if err != nil {
			return err
		}

		// Copy over some annotations that we want to keep around. Erase others
		copyList := []string{
			eventing.CreatorAnnotation,
			eventing.UpdaterAnnotation,
		}

		// If the target Annotation doesn't exist, create it even if
		// we don't end up copying anything over so that we erase all
		// existing annotations
		if trigger.Annotations == nil {
			trigger.Annotations = map[string]string{}
		}

		// Do the actual copy now, but only if it's in the source annotation
		for _, k := range copyList {
			if v, ok := existingTrigger.Annotations[k]; ok {
				trigger.Annotations[k] = v
			}
		}

		trigger.ResourceVersion = existingTrigger.ResourceVersion
		err = client.UpdateTrigger(trigger)
		if err != nil {
			// Retry to update when a resource version conflict exists
			if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
				retries++
				continue
			}
			return err
		}
		fmt.Fprintf(out, "Trigger '%s' successfully replaced in namespace '%s'.\n", trigger.Name, namespace)
		return nil
	}
}

func triggerExists(client v1alpha1.KnClient, name string, namespace string) (bool, error) {
	_, err := client.GetTrigger(name)
	if api_errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Create trigger struct from provided options
func constructTrigger(cmd *cobra.Command, editFlags triggerEditFlags, name string, namespace string) (*eventing_v1alpha1_api.Trigger,
	error) {

	trigger := eventing_v1alpha1_api.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	// TODO: Should it always be `runLatest` ?
	trigger.Spec.Broker = "default"

	err := editFlags.Apply(&trigger, cmd)
	if err != nil {
		return nil, err
	}
	return &trigger, nil
}

func showUrl(client v1alpha1.KnClient, triggerName string, namespace string, out io.Writer) error {
	_, err := client.GetTrigger(triggerName)
	if err != nil {
		return fmt.Errorf("cannot fetch trigger '%s' in namespace '%s' for extracting the URL: %v", triggerName, namespace, err)
	}

	//	url := trigger.Status.URL.String()
	//	if url == "" {
	//		url = trigger.Status.DeprecatedDomain
	//	}
	url := "triggers-dont-have-urls"
	fmt.Fprintln(out, "\nTrigger URL:")
	fmt.Fprintf(out, "%s\n", url)
	return nil
}
