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
	gimporter "github.com/knative/client/pkg/kn/commands/importer/generic"
	"github.com/knative/eventing/pkg/apis/eventing"
	eventing_v1alpha1_api "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/spf13/cobra"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ceAttributeKey = "knimportertrigger"
)

func NewTriggerCreateCommand(p *commands.KnParams) *cobra.Command {
	var triggerEditFlags EditFlags
	var waitFlags commands.WaitFlags
	var importerEditFlags gimporter.EditFlags

	triggerCreateCommand := &cobra.Command{
		Use:   "create NAME --image IMAGE",
		Short: "Create a Trigger and optionally an importer. If an Importer is created, then the created Trigger receives events exclusively from the created importer.",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'trigger create' requires one argument, the name for the trigger")
			}
			name := args[0]

			ns, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			triggerEditFlags.CopyDuplicateImporterFlags(importerEditFlags)

			// TODO GetCRD and use its real name, not the guess given on the command line.

			createImporter := triggerEditFlags.Importer != ""
			if createImporter {
				// Artificially add the correlating CloudEvent attribute.
				ceValue := fmt.Sprintf("%s/%s:%s", triggerEditFlags.Importer, ns, name)
				if triggerEditFlags.FilterAttributes == nil {
					triggerEditFlags.FilterAttributes = make(map[string]string, 1)
				}
				triggerEditFlags.FilterAttributes[ceAttributeKey] = ceValue
				if importerEditFlags.ExtensionOverrides == nil {
					importerEditFlags.ExtensionOverrides = make(map[string]string, 1)
				}
				importerEditFlags.ExtensionOverrides[ceAttributeKey] = ceValue
			}

			triggerCreateF := triggerCreateFunc(p, &triggerEditFlags, &waitFlags)
			createdTrigger, err := triggerCreateF(cmd, []string{name})
			if err != nil {
				return err
			}

			if createImporter {
				importerCreate := gimporter.CreateCOFunc(p, &importerEditFlags, &waitFlags, gimporter.WithControllingOwner(createdTrigger))
				if err := importerCreate(cmd, []string{triggerEditFlags.Importer, name}); err != nil {
					return err
				}
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(triggerCreateCommand.Flags(), false)
	triggerEditFlags.AddCreateFlags(triggerCreateCommand)
	waitFlags.AddConditionWaitFlags(triggerCreateCommand, 60, "Create", "trigger")
	importerEditFlags.AddCreateFlags(triggerCreateCommand, "importer-")
	return triggerCreateCommand
}

func triggerCreateFunc(p *commands.KnParams, editFlags *EditFlags, waitFlags *commands.WaitFlags) func(cmd *cobra.Command, args []string) (*eventing_v1alpha1_api.Trigger, error) {
	return func(cmd *cobra.Command, args []string) (*eventing_v1alpha1_api.Trigger, error) {
		if len(args) != 1 {
			return nil, errors.New("'trigger create' requires the trigger name given as single argument")
		}
		name := args[0]
		if editFlags.SubscriberName == "" {
			return nil, errors.New("'trigger create' requires the subscriber name to run provided with the --subscriber option")
		}

		namespace, err := p.GetNamespace(cmd)
		if err != nil {
			return nil, err
		}

		trigger, err := constructTrigger(cmd, *editFlags, name, namespace)
		if err != nil {
			return nil, err
		}

		client, err := p.NewEventingClient(namespace)
		if err != nil {
			return nil, err
		}

		triggerExists, err := triggerExists(client, name, namespace)
		if err != nil {
			return nil, err
		}

		var createdTrigger *eventing_v1alpha1_api.Trigger
		if triggerExists {
			if !editFlags.ForceCreate {
				return nil, fmt.Errorf(
					"cannot create trigger '%s' in namespace '%s' "+
						"because the trigger already exists and no --force option was given", name, namespace)
			}
			createdTrigger, err = replaceTrigger(client, trigger, namespace, cmd.OutOrStdout())
		} else {
			createdTrigger, err = createTrigger(client, trigger, namespace, cmd.OutOrStdout())
		}
		if err != nil {
			return nil, err
		}

		if !waitFlags.Async {
			out := cmd.OutOrStdout()
			err := waitForTrigger(client, name, out, waitFlags.TimeoutInSeconds)
			if err != nil {
				return nil, err
			}
			return createdTrigger, nil
		}

		return createdTrigger, nil
	}
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

func createTrigger(client v1alpha1.KnClient, trigger *eventing_v1alpha1_api.Trigger, namespace string, out io.Writer) (*eventing_v1alpha1_api.Trigger, error) {
	createdTrigger, err := client.CreateTrigger(trigger)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "Trigger '%s' successfully created in namespace '%s'.\n", trigger.Name, namespace)
	return createdTrigger, nil
}

func replaceTrigger(client v1alpha1.KnClient, trigger *eventing_v1alpha1_api.Trigger, namespace string, out io.Writer) (*eventing_v1alpha1_api.Trigger, error) {
	var retries = 0
	for {
		existingTrigger, err := client.GetTrigger(trigger.Name)
		if err != nil {
			return nil, err
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
		updatedTrigger, err := client.UpdateTrigger(trigger)
		if err != nil {
			// Retry to update when a resource version conflict exists
			if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
				retries++
				continue
			}
			return nil, err
		}
		fmt.Fprintf(out, "Trigger '%s' successfully replaced in namespace '%s'.\n", trigger.Name, namespace)
		return updatedTrigger, nil
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
func constructTrigger(cmd *cobra.Command, editFlags EditFlags, name string, namespace string) (*eventing_v1alpha1_api.Trigger,
	error) {

	trigger := eventing_v1alpha1_api.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	trigger.Spec.Broker = "default"

	err := editFlags.Apply(&trigger, cmd)
	if err != nil {
		return nil, err
	}
	return &trigger, nil
}
