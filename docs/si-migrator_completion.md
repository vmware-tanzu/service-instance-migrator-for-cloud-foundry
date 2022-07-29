## si-migrator completion

Generate completion script

### Synopsis

To load completions:

	Bash:

	  $ source <(yourprogram completion bash)

	  # To load completions for each session, execute once:
	  # Linux:
	  $ yourprogram completion bash > /etc/bash_completion.d/yourprogram
	  # macOS:
	  $ yourprogram completion bash > /usr/local/etc/bash_completion.d/yourprogram

	Zsh:

	  # If shell completion is not already enabled in your environment,
	  # you will need to enable it.  You can execute the following once:

	  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

	  # To load completions for each session, execute once:
	  $ yourprogram completion zsh > "${fpath[1]}/_yourprogram"

	  # You will need to start a new shell for this setup to take effect.

	fish:

	  $ yourprogram completion fish | source

	  # To load completions for each session, execute once:
	  $ yourprogram completion fish > ~/.config/fish/completions/yourprogram.fish

	PowerShell:

	  PS> yourprogram completion powershell | Out-String | Invoke-Expression

	  # To load completions for every new session, run:
	  PS> yourprogram completion powershell > yourprogram.ps1
	  # and source this file from your PowerShell profile.
	

```
si-migrator completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --debug               Enable debug logging
      --dry-run             Display command without executing
      --instances strings   Service instances to migrate [default: all service instances]
  -n, --non-interactive     Don't ask for user input
      --services strings    Service types to migrate [default: all service types]
```

### SEE ALSO

* [si-migrator](si-migrator.md)	 - The si-migrator CLI is a tool for migrating service instances from one TAS (Tanzu Application Service) to another

###### Auto generated by spf13/cobra on 28-Jul-2022