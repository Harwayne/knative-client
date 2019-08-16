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

package generic

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/eventing/pkg/apis/eventing"
	"github.com/spf13/cobra"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	trueVal = true
)

func NewImporterCreateCOCommand(p *commands.KnParams) *cobra.Command {
	var editFlags EditFlags
	var waitFlags commands.WaitFlags

	importerCreateCommand := &cobra.Command{
		Use:   "create NAME --image IMAGE",
		Short: "Create an importer custom object.",
		Example: `
  # Create a importer 'mysvc' using image at dev.local/ns/image:latest
  kn importer create mysvc --image dev.local/ns/image:latest

  # Create a importer with multiple environment variables
  kn importer create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest

  # Create or replace a importer 's1' with image dev.local/ns/image:v2 using --force flag
  # if importer 's1' doesn't exist, it's just a normal create operation
  kn importer create --force s1 --image dev.local/ns/image:v2

  # Create or replace environment variables of importer 's1' using --force flag
  kn importer create --force s1 --env KEY1=NEW_VALUE1 --env NEW_KEY2=NEW_VALUE2 --image dev.local/ns/image:v1

  # Create importer 'mysvc' with port 80
  kn importer create mysvc --port 80 --image dev.local/ns/image:latest

  # Create or replace default resources of a importer 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn importer create --force s1 --image dev.local/ns/image:v1`,

		RunE: CreateCOFunc(p, &editFlags, &waitFlags),
	}
	commands.AddNamespaceFlags(importerCreateCommand.Flags(), false)
	editFlags.AddCreateFlags(importerCreateCommand, "")
	waitFlags.AddConditionWaitFlags(importerCreateCommand, 60, "Create", "importer")
	return importerCreateCommand
}

func CreateCOFunc(p *commands.KnParams, editFlags *EditFlags, waitFlags *commands.WaitFlags, options ...Option) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("'importer create-co' requires the importer CRD name as the first argument and the CO name as the second argument")
		}
		crdName := args[0]
		name := args[1]

		ns, err := p.GetNamespace(cmd)
		if err != nil {
			return err
		}

		client, crd, err := GetCRD(p, crdName)
		if err != nil {
			return err
		}
		gvr := getGVR(crd)
		gvk := getGVK(crd)

		importer, err := constructImporter(cmd, gvk, *editFlags, name, ns)
		if err != nil {
			return err
		}

		for _, option := range options {
			if err = option(importer); err != nil {
				return err
			}
		}

		nc := client.Resource(gvr).Namespace(ns)

		importerExists, err := importerExists(nc, name)
		if err != nil {
			return err
		}

		if importerExists {
			if !editFlags.ForceCreate {
				return fmt.Errorf(
					"cannot create importer '%s' in namespace '%s' "+
						"because the importer already exists and no --force option was given", name, ns)
			}
			importer, err = replaceImporter(nc, importer, cmd.OutOrStdout())
		} else {
			importer, err = createImporter(nc, importer, cmd.OutOrStdout())
		}
		if err != nil {
			return err
		}

		if !waitFlags.Async {
			out := cmd.OutOrStdout()
			timeout := time.Duration(waitFlags.TimeoutInSeconds) * time.Second
			err := waitForUnstructured(nc, importer.GetName(), out, timeout)
			if err != nil {
				return err
			}
			return nil
		}

		return nil
	}
}

type Option func(*unstructured.Unstructured) error

func WithControllingOwner(owner metav1.Object) Option {
	return func(u *unstructured.Unstructured) error {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(owner)
		if err != nil {
			return err
		}
		o := unstructured.Unstructured{
			Object: unstructuredMap,
		}
		owners := u.GetOwnerReferences()
		owners = append(owners, metav1.OwnerReference{
			APIVersion: o.GetAPIVersion(),
			Kind:       o.GetKind(),
			Name:       o.GetName(),
			UID:        o.GetUID(),
			Controller: &trueVal,
		})
		u.SetOwnerReferences(owners)
		// log.Error("WithControllingOwner.afterSet ", u.GetOwnerReferences(), " ::: I had set ", owners)

		// u.SetOwnerReferences(owners) doesn't seem to work...so do it the hard way.
		var metadata metav1.ObjectMeta
		var ok bool
		if metadata, ok = u.Object["metadata"].(metav1.ObjectMeta); !ok {
			return errors.New("not an ObjectMeta")
		}
		metadata.OwnerReferences = owners
		u.Object["metadata"] = metadata
		// log.Error("WithControllingOwner.afterManualSet ", u.GetOwnerReferences(), " ::: I had set ", owners)

		return nil
	}
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

func createImporter(client dynamic.ResourceInterface, importer *unstructured.Unstructured, out io.Writer) (*unstructured.Unstructured, error) {
	created, err := client.Create(importer, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "Importer '%s' successfully created in namespace '%s'.\n", created.GetName(), created.GetNamespace())
	return created, nil
}

func replaceImporter(client dynamic.ResourceInterface, importer *unstructured.Unstructured, out io.Writer) (*unstructured.Unstructured, error) {
	var retries = 0
	for {
		existingImporter, err := client.Get(importer.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Copy over some annotations that we want to keep around. Erase others
		copyList := []string{
			eventing.CreatorAnnotation,
			eventing.UpdaterAnnotation,
		}

		newAnnotations := make(map[string]string)
		// Do the actual copy now, but only if it's in the source annotation
		for _, k := range copyList {
			if v, ok := existingImporter.GetAnnotations()[k]; ok {
				newAnnotations[k] = v
			}
		}
		importer.SetAnnotations(newAnnotations)

		importer.SetResourceVersion(existingImporter.GetResourceVersion())
		updated, err := client.Update(importer, metav1.UpdateOptions{})
		if err != nil {
			// Retry to update when a resource version conflict exists
			if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
				retries++
				continue
			}
			return nil, err
		}
		fmt.Fprintf(out, "Importer '%s' successfully replaced in namespace '%s'.\n", updated.GetName(), updated.GetNamespace())
		return updated, nil
	}
}

func importerExists(client dynamic.ResourceInterface, name string) (bool, error) {
	_, err := client.Get(name, metav1.GetOptions{})
	if api_errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Create importer struct from provided options
func constructImporter(cmd *cobra.Command, gvk schema.GroupVersionKind, editFlags EditFlags, name string, ns string) (*unstructured.Unstructured, error) {
	m := make(map[string]interface{})
	m["metadata"] = metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}
	m["apiVersion"], m["kind"] = gvk.ToAPIVersionAndKind()

	err := editFlags.Apply(m, cmd)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{
		Object: m,
	}, nil
}
