## si-migrator export org

Export org

```
si-migrator export org [flags]
```

### Examples

```
service-instance-migrator export org sample-org
```

### Options

```
  -h, --help   help for org
```

### Options inherited from parent commands

```
      --debug               Enable debug logging
      --dry-run             Display command without executing
      --export-dir string   Directory where service instances will be placed or read (default "export")
      --instances strings   Service instances to migrate [default: all service instances]
  -n, --non-interactive     Don't ask for user input
      --services strings    Service types to migrate [default: all service types]
```

### SEE ALSO

* [si-migrator export](si-migrator_export.md)	 - Export service instances from an org or space.

###### Auto generated by spf13/cobra on 28-Jul-2022
