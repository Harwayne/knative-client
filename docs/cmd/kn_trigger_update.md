## kn trigger update

Update a trigger.

### Synopsis

Update a trigger.

```
kn trigger update NAME [flags]
```

### Examples

```

  # Updates a trigger 'mysvc' with new environment variables
  kn trigger update mysvc --env KEY1=VALUE1 --env KEY2=VALUE2

  # Update a trigger 'mysvc' with new port
  kn trigger update mysvc --port 80

  # Updates a trigger 'mysvc' with new requests and limits parameters
  kn trigger update mysvc --requests-cpu 500m --limits-memory 1024Mi
```

### Options

```
      --async                   Update trigger and don't wait for it to become ready.
      --broker string           Broker the Trigger associates with. (default "default")
      --filter stringToString   Filter attributes, expressed as a CSV. (default [])
  -h, --help                    help for update
  -n, --namespace string        List the requested object(s) in given namespace.
      --subscriber string       Name of the Knative Service that is subscribing to this Trigger.
      --wait-timeout int        Seconds to wait before giving up on waiting for trigger to be ready. (default 60)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Trigger command group

