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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"os"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra/doc"
)

func main() {
	cfg, err := config.NewDefaultConfig()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	rootCmd := cmd.CreateRootCommand(
		cfg,
		migrate.NewConfigLoader(cfg, nil, cmd.NoopPropertiesProvider{}),
		migrate.NewConfigLoader(cfg, nil, cmd.NoopPropertiesProvider{}),
	)
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err := doc.GenMarkdownTree(rootCmd, path+"/docs/"); err != nil {
		panic(err)
	}
}
