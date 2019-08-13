## kn eventing deactivate

Deactivate Knative eventing in the given namespace.

### Synopsis


Deactivate Knative eventing in the given namespace. This does not delete
anything currently running. Rather it will stop reconciling that namespace. So
if anything is altered or deleted, nothing will put it back into a working
state.

```
kn eventing deactivate [flags]
```

### Examples

```

  # Activate Knative eventing in the namespace that kn is associated with.
  kn eventing activate

  # Activate Knative eventing in the namespace 'foo'
  kn eventing activate --namespace foo
```

### Options

```
      --async              Activate Eventing and don't wait for it to become ready.
  -h, --help               help for deactivate
  -n, --namespace string   List the requested object(s) in given namespace.
      --wait-timeout int   Seconds to wait before giving up on waiting for Eventing to be ready. (default 60)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn eventing](kn_eventing.md)	 - Eventing command group

