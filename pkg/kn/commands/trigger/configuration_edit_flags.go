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
	"fmt"

	gimporter "github.com/knative/client/pkg/kn/commands/importer/generic"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

type EditFlags struct {
	Broker           string
	FilterAttributes map[string]string
	SubscriberName   string
	Importer         string

	ForceCreate bool
}

func (p *EditFlags) AddUpdateFlags(command *cobra.Command) {
	command.Flags().StringToStringVar(&p.FilterAttributes, "filter", make(map[string]string), "Filter attributes, expressed as a CSV.")
	command.Flags().StringVar(&p.SubscriberName, "subscriber", "", "Name of the Knative Service that is subscribing to this Trigger.")
	command.Flags().StringVar(&p.Importer, "importer", "", "Name of the Importer CRD to create pointing at this Trigger.")
}

func (p *EditFlags) AddCreateFlags(command *cobra.Command) {
	p.AddUpdateFlags(command)
	if err := command.MarkFlagRequired("subscriber"); err != nil {
		panic(fmt.Errorf("marking flag required: %v", err))
	}
}

func (p *EditFlags) CopyDuplicateImporterFlags(flags gimporter.EditFlags) {
	p.Broker = flags.Broker
	p.ForceCreate = flags.ForceCreate
}

func (p *EditFlags) Apply(t *eventingv1alpha1.Trigger, _ *cobra.Command) error {
	fa := eventingv1alpha1.TriggerFilterAttributes(p.FilterAttributes)
	t.Spec = eventingv1alpha1.TriggerSpec{
		Broker: p.Broker,
		Filter: &eventingv1alpha1.TriggerFilter{
			Attributes: &fa,
		},
		Subscriber: &eventingv1alpha1.SubscriberSpec{
			Ref: &corev1.ObjectReference{
				APIVersion: "serving.knative.dev/v1alpha1",
				Kind:       "Service",
				Name:       p.SubscriberName,
			},
		},
	}
	return nil
}
