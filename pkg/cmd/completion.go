/*
 *  Copyright 2022 VMware, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func createCompletionCommand() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

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
	`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  matchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return fmt.Errorf("Powershell completions not yet supported")
			}
			return nil
		},
	}

	return completionCmd
}

func matchAll(checks ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, check := range checks {
			if err := check(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
