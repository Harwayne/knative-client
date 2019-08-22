## kn importer generic delete

Delete an importer custom object.

### Synopsis

Delete an importer custom object.

```
kn importer generic delete CRD_NAME CO_NAME [flags]
```

### Examples

```

  # Delete the importer 'imp1' of type
  # 'apiserversources.sources.eventing.knative.dev' in the default namespace.
  kn importer delete apiserversources.sources.eventing.knative.dev imp1

  # Delete the importer 'imp1' of type 'apiserversources' in the default
  # namespace. 'apiserversource' can be the name, kind, singular, or plural name
  # of the CRD. If there are multiple CRDs that match 'apiserversource', then an
  # error is returned.
  kn importer delete apiserversources imp1
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   List the requested object(s) in given namespace.
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn importer generic](kn_importer_generic.md)	 - Generic Importer command group

