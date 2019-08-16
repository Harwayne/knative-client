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

package generic

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
)

type EditFlags struct {
	Broker             string
	Parameters         map[string]string
	EventTypes         []string
	ExtensionOverrides map[string]string
	Secret             secret

	ForceCreate bool
}

type secret struct {
	set       bool
	specField string
	name      string
	key       string
}

var _ pflag.Value = (*secret)(nil)

func (p *EditFlags) addUpdateFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.Broker, "broker", "default", "Broker the Importer associates with.")
	command.Flags().StringToStringVar(&p.Parameters, "parameters", make(map[string]string), "Parameters used in the spec of the created importer, expressed as a CSV.")
	command.Flags().StringSliceVar(&p.EventTypes, "eventTypes", []string{}, "Comma separated list of event types.")
	command.Flags().StringToStringVar(&p.ExtensionOverrides, "extensionOverrides", make(map[string]string), "CloudEvent extension attribute overrides.")
	command.Flags().Var(&p.Secret, "secret", "Secret to inject into the spec as a SecretKeySelector. In the form `specField=secretName:key`. Which will set `spec.specField` to a SecretKeySelector.")
}

func (p *EditFlags) AddCreateFlags(command *cobra.Command) {
	p.addUpdateFlags(command)
	command.Flags().BoolVar(&p.ForceCreate, "force", false, "Create importer forcefully, replaces existing importer if any.")
}

func (p *EditFlags) Apply(m map[string]interface{}, cmd *cobra.Command) error {
	spec := make(map[string]interface{})
	for k, v := range p.Parameters {
		spec[k] = v
	}
	spec["sink"] = v1.ObjectReference{
		APIVersion: "eventing.knative.dev/v1alpha1",
		Kind:       "Broker",
		Name:       p.Broker,
	}
	if len(p.EventTypes) > 0 {
		spec["eventTypes"] = p.EventTypes
	}
	if len(p.ExtensionOverrides) > 0 {
		spec["ceOverrides"] = map[string]interface{}{
			"extensions": p.ExtensionOverrides,
		}
	}
	if p.Secret.set {
		spec[p.Secret.specField] = v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: p.Secret.name,
			},
			Key: p.Secret.key,
		}
	}
	m["spec"] = spec
	return nil
}

func (s secret) String() string {
	return fmt.Sprintf("%s=%s:%s", s.specField, s.name, s.key)
}

func (s secret) Set(f string) error {
	i := strings.Index(f, "=")
	if i < 0 {
		return fmt.Errorf("did not match the expected syntax (missing '=') %q", f)
	}
	s.specField = f[:i]
	if len(f) < i+1 {
		return fmt.Errorf("did not match expected synatax (nothing after '=') %q", f)
	}
	sks := f[i+1:]
	i = strings.Index(sks, ":")
	if i < 0 {
		return fmt.Errorf("did not match the expected syntax (missing ':') %q", f)
	} else if len(sks) < i+1 {
		return fmt.Errorf("did not match expected syntax (nothing after ':') %q", f)
	}
	s.name = sks[:i]
	s.key = sks[i+1:]

	if s.specField == "" {
		return fmt.Errorf("empty specField %q", f)
	}
	if s.name == "" {
		return fmt.Errorf("empty secret name %q", f)
	}
	if s.key == "" {
		return fmt.Errorf("empty secret key %q", f)
	}
	s.set = true
	return nil
}

func (s secret) Type() string {
	return "secret"
}
