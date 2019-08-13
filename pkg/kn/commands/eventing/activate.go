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

package eventing

import (
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

const (
	labelKey          = "knative-eventing-injection"
	enabledLabelValue = "enabled"
)

var (
	nsGVR = corev1.SchemeGroupVersion.WithResource("namespaces")
)

func NewEventingActivateCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags

	eventingActivateCommand := &cobra.Command{
		Use:   "activate",
		Short: "Activate Knative eventing in the given namespace.",
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
			err = addLabel(client, ns)
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

			fmt.Fprintf(cmd.OutOrStdout(), "Eventing has been activated in namespace '%s'.\n", ns.Name)
			return nil
		},
	}
	commands.AddNamespaceFlags(eventingActivateCommand.Flags(), false)
	waitFlags.AddConditionWaitFlags(eventingActivateCommand, 60, "Activate", "Eventing")
	return eventingActivateCommand
}

func getNamespace(c dynamic.Interface, name string) (corev1.Namespace, error) {
	u, err := c.Resource(nsGVR).Get(name, metav1.GetOptions{})
	if err != nil {
		return corev1.Namespace{}, err
	}

	var ns corev1.Namespace
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ns); err != nil {
		return corev1.Namespace{}, fmt.Errorf("converting unstructured: %v", err)
	}
	return ns, nil
}

func addLabel(client dynamic.Interface, ns corev1.Namespace) error {
	if ns.Labels[labelKey] == enabledLabelValue {
		return nil
	}
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels[labelKey] = enabledLabelValue
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ns)
	if err != nil {
		return err
	}
	u := &unstructured.Unstructured{
		Object: m,
	}
	_, err = client.Resource(nsGVR).Update(u, metav1.UpdateOptions{})
	return err
}
