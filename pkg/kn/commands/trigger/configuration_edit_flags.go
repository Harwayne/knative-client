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

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

type triggerEditFlags struct {
	Broker           string
	FilterAttributes map[string]string
	SubscriberName   string

	ForceCreate bool
}

func (p *triggerEditFlags) AddUpdateFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.Broker, "broker", "default", "Broker the Trigger associates with.")
	command.Flags().StringToStringVar(&p.FilterAttributes, "filter", make(map[string]string), "Filter attributes, expressed as a CSV.")
	command.Flags().StringVar(&p.SubscriberName, "subscriber", "", "Name of the Knative Service that is subscribing to this Trigger.")
}

func (p *triggerEditFlags) AddCreateFlags(command *cobra.Command) {
	p.AddUpdateFlags(command)
	command.Flags().BoolVar(&p.ForceCreate, "force", false, "Create trigger forcefully, replaces existing trigger if any.")
	if err := command.MarkFlagRequired("subscriber"); err != nil {
		panic(fmt.Errorf("marking flag required: %v", err))
	}
}

func (p *triggerEditFlags) Apply(t *eventingv1alpha1.Trigger, cmd *cobra.Command) error {
	fa := eventingv1alpha1.TriggerFilterAttributes(p.FilterAttributes)
	t.Spec = eventingv1alpha1.TriggerSpec{
		Broker: p.Broker,
		Filter: &eventingv1alpha1.TriggerFilter{
			Attributes: &fa,
		},
		Subscriber: &eventingv1alpha1.SubscriberSpec{
			Ref: &v1.ObjectReference{
				APIVersion: "serving.knative.dev/v1alpha1",
				Kind:       "Service",
				Name:       p.SubscriberName,
			},
		},
	}
	return nil
}
