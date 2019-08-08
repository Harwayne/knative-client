## kn trigger delete

Delete a trigger.

### Synopsis

Delete a trigger.

```
kn trigger delete NAME [flags]
```

### Examples

```

  # Delete a trigger 'svc1' in default namespace
  kn trigger delete svc1

  # Delete a trigger 'svc2' in 'ns1' namespace
  kn trigger delete svc2 -n ns1
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

* [kn trigger](kn_trigger.md)	 - Trigger command group

