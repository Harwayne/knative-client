## kn importer generic create

Create an importer custom object.

### Synopsis

Create an importer custom object.

```
kn importer generic create CRD_NAME CO_NAME [flags]
```

### Examples

```

  # Create a CronJobSource importer named 'my-imp' which sends its events to
  # the default Broker.
  kn importer generic create CronJobSource my-imp

  TODO: Add more
```

### Options

```
      --async                                Create importer and don't wait for it to become ready.
      --broker string                        Broker the Importer associates with. (default "default")
      --event-types strings                  Comma separated list of event types.
      --extension-overrides stringToString   CloudEvent extension attribute overrides. (default [])
      --force                                Create importer forcefully, replaces existing importer if any.
  -h, --help                                 help for create
  -n, --namespace string                     List the requested object(s) in given namespace.
      --parameters stringToString            Parameters used in the spec of the created importer, expressed as a CSV. (default [])
      --secret specField=secretName:key      Secret to inject into the spec as a SecretKeySelector. In the form specField=secretName:key. Which will set `spec.specField` to a SecretKeySelector. (default =:)
      --wait-timeout int                     Seconds to wait before giving up on waiting for importer to be ready. (default 60)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn importer generic](kn_importer_generic.md)	 - Generic Importer command group

