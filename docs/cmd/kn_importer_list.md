## kn importer list

List available triggers.

### Synopsis

List available triggers.

```
kn importer list [name] [flags]
```

### Examples

```

  # List all triggers
  kn trigger list

  # List all triggers in JSON output format
  kn trigger list -o json

  # List trigger 'web'
  kn trigger list web
```

### Options

```
      --all-namespaces                If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for list
  -n, --namespace string              List the requested object(s) in given namespace.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-file.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn importer](kn_importer.md)	 - Importer command group
