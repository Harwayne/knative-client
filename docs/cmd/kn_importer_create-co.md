## kn importer create-co

Create an importer custom object.

### Synopsis

Create an importer custom object.

```
kn importer create-co NAME --image IMAGE [flags]
```

### Examples

```

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
  kn importer create --force s1 --image dev.local/ns/image:v1
```

### Options

```
      --async                       Create importer and don't wait for it to become ready.
      --broker string               Broker the Importer associates with. (default "default")
      --force                       Create importer forcefully, replaces existing importer if any.
  -h, --help                        help for create-co
  -n, --namespace string            List the requested object(s) in given namespace.
      --parameters stringToString   Parameters used in the spec of the created importer, expressed as a CSV. (default [])
      --wait-timeout int            Seconds to wait before giving up on waiting for importer to be ready. (default 60)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn importer](kn_importer.md)	 - Importer command group

