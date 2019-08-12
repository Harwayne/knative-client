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

	"k8s.io/client-go/dynamic"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
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
		NewImporterDeleteCOCommand(p))
	return importerCmd
}

func waitForImporter(client dynamic.ResourceInterface, name string, out io.Writer, timeout int) error {
	fmt.Fprintf(out, "Waiting for importer '%s' to become ready ... ", name)
	flush(out)

	// TODO
	// err := client.WaitForTrigger(name, time.Duration(timeout)*time.Second)
	err := errors.New("waitForImporter not yet implemented")
	if err != nil {
		fmt.Fprintln(out)
		return err
	}
	fmt.Fprintln(out, "OK")
	return nil
}
