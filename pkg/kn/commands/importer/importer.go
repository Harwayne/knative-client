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
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	// How often to retry in case of an optimistic lock error when replacing a service (--force)
	MaxUpdateRetries = 3
)

func NewImporterCommand(p *commands.KnParams) *cobra.Command {
	importerCmd := &cobra.Command{
		Use: "importer",
		Aliases: []string{
			"importers",
			"source",
			"sources",
		},
		Short: "Importer command group",
	}
	importerCmd.AddCommand(
		NewImporterListCommand(p),
		NewImporterDescribeCommand(p),
		NewImporterCreateCOCommand(p),
		NewImporterDeleteCOCommand(p),
		NewImporterDescribeCOCommand(p),
		NewImporterListCOCommand(p))
	return importerCmd
}

func waitForUnstructured(client dynamic.ResourceInterface, name string, out io.Writer, timeout time.Duration) error {
	fmt.Fprintf(out, "Waiting for importer %q to become ready ... ", name)
	flush(out)

	t := time.After(timeout)

	time.Sleep(2 * time.Second)
	for {
		select {
		case <-t:
			fmt.Fprintln(out, "Timed out waiting for the importer to get ready")
			return errors.New("time out waiting for the importer to get ready")
		default:
			ready, err := getImporterReady(client, name)
			if err != nil {
				fmt.Fprintln(out, "Error getting importer ", err)
				return err
			}
			if ready {
				fmt.Fprintln(out, "OK")
				return nil
			}
		}
	}
}

func getImporterReady(client dynamic.ResourceInterface, name string) (bool, error) {
	u, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		if err.Error() == "status not present" {
			// Give extra time for the status to become present.
			return false, nil
		}
		return false, err
	}
	c, err := extractReadyCondition(*u)
	if err != nil {
		return false, err
	}
	if c.Status == corev1.ConditionTrue {
		return true, nil
	}
	return false, nil
}
