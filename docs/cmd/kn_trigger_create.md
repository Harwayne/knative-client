## kn trigger create

Create a trigger.

### Synopsis

Create a trigger.

```
kn trigger create NAME --image IMAGE [flags]
```

### Examples

```

  # Create a trigger 'mysvc' using image at dev.local/ns/image:latest
  kn trigger create mysvc --image dev.local/ns/image:latest

  # Create a trigger with multiple environment variables
  kn trigger create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest

  # Create or replace a trigger 's1' with image dev.local/ns/image:v2 using --force flag
  # if trigger 's1' doesn't exist, it's just a normal create operation
  kn trigger create --force s1 --image dev.local/ns/image:v2

  # Create or replace environment variables of trigger 's1' using --force flag
  kn trigger create --force s1 --env KEY1=NEW_VALUE1 --env NEW_KEY2=NEW_VALUE2 --image dev.local/ns/image:v1

  # Create trigger 'mysvc' with port 80
  kn trigger create mysvc --port 80 --image dev.local/ns/image:latest

  # Create or replace default resources of a trigger 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn trigger create --force s1 --image dev.local/ns/image:v1
```

### Options

```
      --async                   Create trigger and don't wait for it to become ready.
      --broker string           Broker the Trigger associates with. (default "default")
      --filter stringToString   Filter attributes, expressed as a CSV. (default [])
      --force                   Create trigger forcefully, replaces existing trigger if any.
  -h, --help                    help for create
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

