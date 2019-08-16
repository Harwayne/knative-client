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

	gimporter "github.com/knative/client/pkg/kn/commands/importer/generic"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
)

const (
	ceAttributeKey = "knimportertrigger"
)

func NewTriggerCreateWithImporterCommand(p *commands.KnParams) *cobra.Command {
	var triggerEditFlags EditFlags
	var waitFlags commands.WaitFlags
	var importerEditFlags gimporter.EditFlags

	triggerCreateCommand := &cobra.Command{
		Use:   "create-with-importer NAME --image IMAGE",
		Short: "Create a Trigger and an importer. The created Trigger receives events exclusively from the created importer.",
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
			if len(args) != 2 {
				return errors.New("'trigger create' requires two arguments, the first is the importer CRD, the second is the name for both the trigger and the importer")
			}
			importerCRD := args[0]
			name := args[1]

			ns, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			triggerEditFlags.CopyDuplicateImporterFlags(importerEditFlags)

			// Artificially add the correlating CloudEvent attribute.
			ceValue := fmt.Sprintf("%s/%s:%s", importerCRD, ns, name)
			if triggerEditFlags.FilterAttributes == nil {
				triggerEditFlags.FilterAttributes = make(map[string]string, 1)
			}
			triggerEditFlags.FilterAttributes[ceAttributeKey] = ceValue
			if importerEditFlags.ExtensionOverrides == nil {
				importerEditFlags.ExtensionOverrides = make(map[string]string, 1)
			}
			importerEditFlags.ExtensionOverrides[ceAttributeKey] = ceValue

			triggerCreateF := triggerCreateFunc(p, &triggerEditFlags, &waitFlags)
			createdTrigger, err := triggerCreateF(cmd, []string{name})
			if err != nil {
				return err
			}

			importerCreate := gimporter.CreateCOFunc(p, &importerEditFlags, &waitFlags, gimporter.WithControllingOwner(createdTrigger))
			if err := importerCreate(cmd, []string{importerCRD, name}); err != nil {
				return err
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(triggerCreateCommand.Flags(), false)
	triggerEditFlags.AddCreateFlags(triggerCreateCommand, false)
	waitFlags.AddConditionWaitFlags(triggerCreateCommand, 60, "Create", "trigger")
	importerEditFlags.AddCreateFlags(triggerCreateCommand)
	return triggerCreateCommand
}
