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

package main

import (
	"fmt"
	"os"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"

	"github.com/spf13/cobra"
)

func main() {
	defer recoverConfigLoader()
	cfg, err := config.NewDefaultConfig()
	if err != nil {
		log.Fatalln(err)
	}

	commandExists := func(args []string) (*cobra.Command, bool) {
		noopCmd := cmd.CreateRootCommand(
			cfg,
			migrate.NewConfigLoader(cfg, nil, cmd.NoopPropertiesProvider{}),
			migrate.NewConfigLoader(cfg, nil, cmd.NoopPropertiesProvider{}),
		)
		if len(args) > 0 {
			for _, c := range noopCmd.Commands() {
				if c.Name() == args[1] {
					return noopCmd, true
				}
			}
			return noopCmd, false
		}
		return noopCmd, false
	}

	if noopRoot, ok := commandExists(os.Args); !ok {
		// execute the cobra help/error message
		if err := noopRoot.Execute(); err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}

	mr, err := config.NewMigrationReader(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	if err := cmd.CreateRootCommand(
		cfg,
		migrate.NewConfigLoader(cfg, mr, om.NewPropertiesProvider(cfg, cfg.Foundations.Source, om.NewClientFactory(), bosh.NewClientFactory(), credhub.NewClientFactory())),
		migrate.NewConfigLoader(cfg, mr, om.NewPropertiesProvider(cfg, cfg.Foundations.Target, om.NewClientFactory(), bosh.NewClientFactory(), credhub.NewClientFactory())),
	).Execute(); err != nil {
		log.Fatalln(err)
	}
}

// Failure to load config can cause a panic. Print the panic message and show path to logs
func recoverConfigLoader() {
	if r := recover(); r != nil {
		_, _ = fmt.Fprintln(os.Stderr, r)
		fmt.Printf("\nOpen %s to view logs\n", log.Logger.Path)
	}
}
