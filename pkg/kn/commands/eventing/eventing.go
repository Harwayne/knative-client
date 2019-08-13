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
	"fmt"
	"io"
	"time"

	eventing_kn_v1alpha1 "github.com/knative/client/pkg/eventing/v1alpha1"
	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
)

const (
	// How often to retry in case of an optimistic lock error when replacing a service (--force)
	MaxUpdateRetries = 3
)

func NewEventingCommand(p *commands.KnParams) *cobra.Command {
	eventingCmd := &cobra.Command{
		Use:   "eventing",
		Short: "Eventing command group",
	}
	eventingCmd.AddCommand(
		NewEventingActivateCommand(p),
		NewEventingDeactivateCommand(p))
	return eventingCmd
}

func waitForBroker(client eventing_kn_v1alpha1.KnClient, name string, out io.Writer, timeout int) error {
	fmt.Fprintf(out, "Waiting for Broker '%s' to become ready ... ", name)
	flush(out)

	err := client.WaitForBroker(name, time.Duration(timeout)*time.Second)
	if err != nil {
		fmt.Fprintln(out)
		return err
	}
	fmt.Fprintln(out, "OK")
	return nil
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
