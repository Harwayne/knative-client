## kn trigger delete

Delete a trigger.

### Synopsis

Delete a trigger.

```
kn trigger delete NAME [flags]
```

### Examples

```

  # Delete trigger 't1' in the default namespace.
  kn trigger delete t1

  # Delete trigger 't2' in the 'ns1' namespace.
  kn trigger delete t2 -n ns1
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

