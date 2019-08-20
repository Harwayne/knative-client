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
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/kn/commands/importer/generic"
	"github.com/knative/client/pkg/printers"
	"github.com/spf13/cobra"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Command for printing out a description of a trigger, meant to be consumed by humans
// It will show information about the trigger itself, but also a summary
// about the associated revisions.

// Whether to print extended information
var printDetails bool

// Max length When to truncate long strings (when not "all" mode switched on)
const truncateAt = 100

var (
	crdGVK = v1beta1.SchemeGroupVersion.WithResource("customresourcedefinitions")
)

// NewTriggerDescribeCommand returns a new command for describing a trigger.
func NewImporterDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details for a given importer, including the event types it generates and its configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("no trigger name provided")
			}
			if len(args) > 1 {
				return errors.New("more than one trigger name provided")
			}
			crdName := args[0]

			_, crd, err := generic.GetCRD(p, crdName)
			if err != nil {
				return err
			}

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(&crd, cmd.OutOrStdout())
			}

			printDetails, err = cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			return describe(cmd.OutOrStdout(), crd)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

// Main action describing the trigger
func describe(w io.Writer, crd v1beta1.CustomResourceDefinition) error {
	dw := printers.NewPrefixWriter(w)

	// Trigger info
	writeCRD(dw, crd)
	dw.WriteLine()
	if err := dw.Flush(); err != nil {
		return err
	}

	// TODO parse and print event types.

	return nil
}

// Write out main trigger information. Use colors for major items.
func writeCRD(dw printers.PrefixWriter, crd v1beta1.CustomResourceDefinition) {
	dw.WriteColsLn(printers.Level0, l("Name"), crd.Name)
	dw.WriteColsLn(printers.Level0, l("Kind"), crd.Spec.Names.Kind)
	writeMapDesc(dw, printers.Level0, crd.Labels, l("Labels"), "")
	writeMapDesc(dw, printers.Level0, crd.Annotations, l("Annotations"), "")
	dw.WriteColsLn(printers.Level0, l("Age"), age(crd.CreationTimestamp.Time))
	writeEventTypes(dw, printers.Level0, crd)
	writeRequiredProperties(dw, printers.Level0, crd)
}

// ======================================================================================
// Helper functions

// Format label (extracted so that color could be added more easily to all labels)
func l(label string) string {
	return label + ":"
}

// Write a map either compact in a single line (possibly truncated) or, if printDetails is set,
// over multiple line, one line per key-value pair. The output is sorted by keys.
func writeMapDesc(dw printers.PrefixWriter, indent int, m map[string]string, label string, labelPrefix string) {
	if len(m) == 0 {
		return
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if printDetails {
		l := labelPrefix + label

		for _, key := range keys {
			dw.WriteColsLn(indent, l, key+"="+m[key])
			l = labelPrefix
		}
		return
	}

	dw.WriteColsLn(indent, label, joinAndTruncate(keys, m))
}

// Join to key=value pair, comma separated, and truncate if longer than a limit
func joinAndTruncate(sortedKeys []string, m map[string]string) string {
	ret := ""
	for _, key := range sortedKeys {
		ret += fmt.Sprintf("%s=%s, ", key, m[key])
		if len(ret) > truncateAt {
			break
		}
	}
	// cut of two latest chars
	ret = strings.TrimRight(ret, ", ")
	if len(ret) <= truncateAt {
		return ret
	}
	return ret[:truncateAt-4] + " ..."
}

func age(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t))
}

type etInfo struct {
	description string
	ceType      string
	ceSchema    string
}

func getEventTypes(crd v1beta1.CustomResourceDefinition) map[string]etInfo {
	reg, ok := crd.Spec.Validation.OpenAPIV3Schema.Properties["registry"]
	if !ok {
		return nil
	}
	et, ok := reg.Properties["eventTypes"]
	if !ok {
		return nil
	}
	m := make(map[string]etInfo)
	for n, v := range et.Properties {
		i := etInfo{
			description: v.Description,
		}
		if t, ok := v.Properties["type"]; ok {
			i.ceType = t.Pattern
		}
		if s, ok := v.Properties["schema"]; ok {
			i.ceSchema = s.Pattern
		}
		m[n] = i
	}
	return m
}

func writeIfNotEmpty(dw printers.PrefixWriter, indent int, prefix string, arg string) {
	if arg != "" {
		dw.Write(indent, "%s: %v\n", prefix, arg)
	}
}

func writeEventTypes(dw printers.PrefixWriter, indent int, crd v1beta1.CustomResourceDefinition) {
	et := getEventTypes(crd)
	dw.Write(indent, "Event Types:\n")
	for n, v := range et {
		dw.Write(indent+1, "%s\n", n)
		writeIfNotEmpty(dw, indent+2, "Description", v.description)
		writeIfNotEmpty(dw, indent+2, "Type", v.ceType)
		writeIfNotEmpty(dw, indent+2, "Schema", v.ceSchema)
	}
}

type props struct {
	required map[string]prop
	optional map[string]prop
}

type prop struct {
	t           string
	description string
}

func writeRequiredProperties(dw printers.PrefixWriter, indent int, crd v1beta1.CustomResourceDefinition) {
	props := getOpenAPIProperties(crd)
	dw.Write(indent, "Configuration:\n")
	if len(props.required) > 0 {
		dw.Write(indent+1, "Required:\n")
		writeProperties(dw, indent+2, props.required)
	}
	if len(props.optional) > 0 {
		dw.Write(indent+1, "Optional:\n")
		writeProperties(dw, indent+2, props.optional)
	}
}

func writeProperties(dw printers.PrefixWriter, indent int, props map[string]prop) {
	for n, p := range props {
		dw.Write(indent, "%s\n", n)
		dw.Write(indent+1, "Type: %s\n", p.t)
		writeIfNotEmpty(dw, indent+1, "Description", p.description)
	}
}

func getOpenAPIProperties(crd v1beta1.CustomResourceDefinition) props {
	props := props{
		required: make(map[string]prop),
		optional: make(map[string]prop),
	}

	spec, ok := crd.Spec.Validation.OpenAPIV3Schema.Properties["spec"]
	if !ok {
		return props
	}
	for n, v := range spec.Properties {
		p := prop{
			description: v.Description,
			t:           v.Type,
		}
		if contains(spec.Required, n) {
			props.required[n] = p
		} else {
			props.optional[n] = p
		}
	}
	return props
}

func contains(l []string, s string) bool {
	for _, i := range l {
		if i == s {
			return true
		}
	}
	return false
}
