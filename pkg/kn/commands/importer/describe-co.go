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

package importer

import (
	"errors"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/printers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Command for printing out a description of a trigger, meant to be consumed by humans
// It will show information about the trigger itself, but also a summary
// about the associated revisions.

// NewTriggerDescribeCommand returns a new command for describing a trigger.
func NewImporterDescribeCOCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "describe-co NAME",
		Short: "Show details for an importer custom object.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("'importer create-co' requires the importer CRD name as the first argument and the CO name as the second argument")
			}
			crdName := args[0]
			name := args[1]

			ns, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			c, crd, err := getCRD(p, crdName)
			if err != nil {
				return err
			}

			gvr := getGVR(crd)

			co, err := c.Resource(gvr).Namespace(ns).Get(name, v1.GetOptions{})
			if err != nil {
				return err
			}

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(co, cmd.OutOrStdout())
			}

			printDetails, err = cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			return describeCO(cmd.OutOrStdout(), *co)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

// Main action describing the trigger
func describeCO(w io.Writer, co unstructured.Unstructured) error {
	dw := printers.NewPrefixWriter(w)

	// Trigger info
	writeCO(dw, co)
	dw.WriteLine()
	if err := dw.Flush(); err != nil {
		return err
	}
	return nil
}

// Write out main trigger information. Use colors for major items.
func writeCO(dw printers.PrefixWriter, co unstructured.Unstructured) {
	dw.WriteColsLn(printers.Level0, l("Name"), co.GetName())
	dw.WriteColsLn(printers.Level0, l("Kind"), co.GetKind())
	writeMapDesc(dw, printers.Level0, co.GetLabels(), l("Labels"), "")
	writeMapDesc(dw, printers.Level0, co.GetAnnotations(), l("Annotations"), "")
	dw.WriteColsLn(printers.Level0, l("Age"), age(co.GetCreationTimestamp().Time))
	printReadiness(dw, printers.Level0, co)
}

// ======================================================================================
// Helper functions

func printReadiness(dw printers.PrefixWriter, level int, u unstructured.Unstructured) {
	c, err := extractReadyCondition(u)
	if err != nil {
		return
	}
	dw.WriteColsLn(level, l("Ready"), string(c.Status))
	if len(c.Reason) > 0 {
		dw.WriteColsLn(level, l("Reason"), c.Reason)
	}
}

func extractReadyCondition(u unstructured.Unstructured) (apis.Condition, error) {
	statusI, ok := u.Object["status"]
	if !ok {
		return apis.Condition{}, errors.New("status not present")
	}
	status, ok := statusI.(map[string]interface{})
	if !ok {
		return apis.Condition{}, errors.New("status not map[string]interface{}")
	}

	s := v1beta1.Status{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(status, &s); err != nil {
		return apis.Condition{}, err
	}

	for _, c := range s.Conditions {
		if c.Type == apis.ConditionReady {
			return c, nil
		}
	}
	return apis.Condition{}, errors.New("ready condition not found")
}
