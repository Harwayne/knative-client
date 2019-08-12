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
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

type importerEditFlags struct {
	Broker     string
	Parameters map[string]string

	ForceCreate bool
}

func (p *importerEditFlags) AddUpdateFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.Broker, "broker", "default", "Broker the Importer associates with.")
	command.Flags().StringToStringVar(&p.Parameters, "parameters", make(map[string]string), "Parameters used in the spec of the created importer, expressed as a CSV.")
}

func (p *importerEditFlags) AddCreateFlags(command *cobra.Command) {
	p.AddUpdateFlags(command)
	command.Flags().BoolVar(&p.ForceCreate, "force", false, "Create importer forcefully, replaces existing importer if any.")
}

func (p *importerEditFlags) Apply(m map[string]interface{}, cmd *cobra.Command) error {
	spec := make(map[string]interface{})
	for k, v := range p.Parameters {
		spec[k] = v
	}
	spec["sink"] = v1.ObjectReference{
		APIVersion: "eventing.knative.dev/v1alpha1",
		Kind:       "Broker",
		Name:       p.Broker,
	}
	m["spec"] = spec
	return nil
}
