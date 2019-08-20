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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/printers"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck/v1beta1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	knv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// Command for printing out a description of a trigger, meant to be consumed by humans
// It will show information about the trigger itself, but also a summary
// about the associated revisions.

// Whether to print extended information
var printDetails bool

// Max length When to truncate long strings (when not "all" mode switched on)
const truncateAt = 100

// NewTriggerDescribeCommand returns a new command for describing a trigger.
func NewTriggerDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details for the named trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("no trigger name provided")
			}
			if len(args) > 1 {
				return errors.New("more than one trigger name provided")
			}
			triggerName := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			trigger, err := client.GetTrigger(triggerName)
			if err != nil {
				return err
			}

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(trigger, cmd.OutOrStdout())
			}

			printDetails, err = cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			return describe(cmd.OutOrStdout(), trigger)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

// Main action describing the trigger
func describe(w io.Writer, trigger *v1alpha1.Trigger) error {
	dw := printers.NewPrefixWriter(w)

	// Trigger info
	writeTrigger(dw, trigger)
	dw.WriteLine()
	if err := dw.Flush(); err != nil {
		return err
	}

	// Condition info
	writeConditions(dw, trigger)
	if err := dw.Flush(); err != nil {
		return err
	}

	return nil
}

// Write out main trigger information. Use colors for major items.
func writeTrigger(dw printers.PrefixWriter, trigger *v1alpha1.Trigger) {
	dw.WriteColsLn(printers.Level0, l("Name"), trigger.Name)
	dw.WriteColsLn(printers.Level0, l("Namespace"), trigger.Namespace)
	writeMapDesc(dw, printers.Level0, trigger.Labels, l("Labels"), "")
	writeMapDesc(dw, printers.Level0, trigger.Annotations, l("Annotations"), "")
	dw.WriteColsLn(printers.Level0, l("Age"), age(trigger.CreationTimestamp.Time))
}

// Print out a table with conditions. Use green for 'ok', and red for 'nok' if color is enabled
func writeConditions(dw printers.PrefixWriter, trigger *v1alpha1.Trigger) {
	dw.WriteColsLn(printers.Level0, l("Conditions"))
	maxLen := getMaxTypeLen(toConditions(trigger.Status.Conditions))
	formatHeader := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	formatRow := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	dw.Write(printers.Level1, formatHeader, "OK", "TYPE", "AGE", "REASON")
	for _, condition := range toConditions(trigger.Status.Conditions) {
		ok := formatStatus(condition.Status)
		reason := condition.Reason
		if printDetails && reason != "" {
			reason = fmt.Sprintf("%s (%s)", reason, condition.Message)
		}
		dw.Write(printers.Level1, formatRow, ok, formatConditionType(condition), age(condition.LastTransitionTime.Inner.Time), reason)
	}
}

func toConditions(conditions knv1beta1.Conditions) v1beta1.Conditions {
	j, err := json.Marshal(conditions)
	if err != nil {
		panic(fmt.Errorf("converting conditions: %v", err))
	}
	c := v1beta1.Conditions{}
	err = json.Unmarshal(j, &c)
	if err != nil {
		panic(fmt.Errorf("unmarshaling conditions: %v", err))
	}
	return c
}

// ======================================================================================
// Helper functions

// Format label (extracted so that color could be added more easily to all labels)
func l(label string) string {
	return label + ":"
}

// Used for conditions table to do own formatting for the table,
// as the tabbed writer doesn't work nicely with colors
func getMaxTypeLen(conditions v1beta1.Conditions) int {
	max := 0
	for _, condition := range conditions {
		if len(condition.Type) > max {
			max = len(condition.Type)
		}
	}
	return max
}

// Color the type of the conditions
func formatConditionType(condition apis.Condition) string {
	return string(condition.Type)
}

// Status in ASCII format
func formatStatus(status corev1.ConditionStatus) string {
	switch status {
	case v1.ConditionTrue:
		return "++"
	case v1.ConditionFalse:
		return "--"
	default:
		return ""
	}
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
