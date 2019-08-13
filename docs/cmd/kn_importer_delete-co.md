## kn importer delete-co

Delete an importer custom object.

### Synopsis

Delete an importer custom object.

```
kn importer delete-co CRD_NAME CO_NAME [flags]
```

### Examples

```

  # Delete a importer 'svc1' in default namespace
  kn importer delete svc1

  # Delete a importer 'svc2' in 'ns1' namespace
  kn importer delete svc2 -n ns1
```

### Options

```
  -h, --help               help for delete-co
  -n, --namespace string   List the requested object(s) in given namespace.
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn importer](kn_importer.md)	 - Importer command group

