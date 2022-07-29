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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cli"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"

	"github.com/spf13/cobra"
)

func CreateRootCommand(
	cfg *config.Config,
	sourceConfigLoader config.Loader,
	targetConfigLoader config.Loader,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "si-migrator",
		Short:        "The si-migrator CLI is a tool for migrating service instances from one TAS (Tanzu Application Service) to another",
		SilenceUsage: true,
	}

	// set version
	if !cli.Env.GitDirty {
		rootCmd.Version = fmt.Sprintf("%s (%s)", cli.Env.Version, cli.Env.GitSha)
	} else {
		rootCmd.Version = fmt.Sprintf("%s (%s, with local modifications)", cli.Env.Version, cli.Env.GitSha)
	}
	rootCmd.Flags().Bool("version", false, "display CLI version")
	rootCmd.PersistentFlags().BoolP("non-interactive", "n", false, "Don't ask for user input")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", cfg.Debug, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&cfg.DryRun, "dry-run", cfg.DryRun, "Display command without executing")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.Services, "services", cfg.Services, "Service types to migrate [default: all service types]")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.Instances, "instances", cfg.Instances, "Service instances to migrate [default: all service instances]")

	rootCmd.AddCommand(createCompletionCommand())

	mr, err := config.NewMigrationReader(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	addExportCommands(config.ContextWithConfig(context.Background(), cfg), rootCmd, cfg, mr, sourceConfigLoader)
	addImportCommands(config.ContextWithConfig(context.Background(), cfg), rootCmd, cfg, mr, targetConfigLoader)

	return rootCmd
}

func addExportCommands(ctx context.Context, rootCmd *cobra.Command, cfg *config.Config, mr config.MigrationReader, configLoader config.Loader) {
	reportSummary := report.NewSummary(os.Stdout)
	clientFactory := migrate.NewClientFactory(configLoader, bosh.NewClient(), om.NewClient(), cfg.Foundations.Source)
	me := cc.NewManifestExporter(cfg, clientFactory)
	factory := NewExportMigratorFactory(cfg, clientFactory)
	sf := cc.NewCloudControllerServiceFactory(clientFactory, me)
	mh := migrate.NewMigratorHelper(mr)
	e := exec.NewExecutor(
		exec.WithDryRun(cfg.DryRun),
		exec.WithDebug(cfg.Debug),
	)
	registry := migrate.NewMigratorRegistry(migrate.NewMigratorFactory(configLoader, clientFactory, mh, e, sf), mh, cfg, configLoader, clientFactory)
	sie := migrate.NewServiceInstanceExporter(cfg, clientFactory, registry, io.NewParser())
	fs := io.NewFileSystemHelper()

	exportCmd := CreateExportCommand(ctx, cfg, factory, sie, fs, reportSummary)
	exportCmd.Flags().StringSliceVar(&cfg.IncludedOrgs, "include-orgs", cfg.IncludedOrgs, "Only orgs matching the regex(es) specified will be included")
	exportCmd.Flags().StringSliceVar(&cfg.ExcludedOrgs, "exclude-orgs", cfg.ExcludedOrgs, "Any orgs matching the regex(es) specified will be excluded")
	exportCmd.PersistentFlags().StringVar(&cfg.ExportDir, "export-dir", cfg.ExportDir, "Directory where service instances will be placed or read")

	exportOrgCmd := CreateExportOrgCommand(ctx, cfg, factory, sie, fs, reportSummary)
	exportCmd.AddCommand(exportOrgCmd)

	exportSpaceCmd := CreateExportSpaceCommand(ctx, cfg, factory, sie, fs, reportSummary)
	exportSpaceCmd.Flags().StringP("org", "o", "", "Org to which the space belongs")
	err := exportSpaceCmd.MarkFlagRequired("org")
	if err != nil {
		log.Fatalln(err)
	}
	exportCmd.AddCommand(exportSpaceCmd)

	rootCmd.AddCommand(exportCmd)
}

func addImportCommands(ctx context.Context, rootCmd *cobra.Command, cfg *config.Config, mr config.MigrationReader, configLoader config.Loader) {
	reportSummary := report.NewSummary(os.Stdout)
	clientFactory := migrate.NewClientFactory(configLoader, bosh.NewClient(), om.NewClient(), cfg.Foundations.Target)
	factory := NewImportMigratorFactory(cfg, clientFactory)
	sf := cc.NewCloudControllerServiceFactory(clientFactory, nil)
	mh := migrate.NewMigratorHelper(mr)
	e := exec.NewExecutor(
		exec.WithDryRun(cfg.DryRun),
		exec.WithDebug(cfg.Debug),
		exec.WithTimeout(20*time.Minute),
	)
	registry := migrate.NewMigratorRegistry(migrate.NewMigratorFactory(configLoader, clientFactory, mh, e, sf), mh, cfg, configLoader, clientFactory)
	sii := migrate.NewServiceInstanceImporter(registry)
	fs := io.NewFileSystemHelper()

	importCmd := CreateImportCommand(ctx, cfg, factory, sii, fs, reportSummary)
	importCmd.Flags().StringSliceVar(&cfg.IncludedOrgs, "include-orgs", cfg.IncludedOrgs, "Only orgs matching the regex(es) specified will be included")
	importCmd.Flags().StringSliceVar(&cfg.ExcludedOrgs, "exclude-orgs", cfg.ExcludedOrgs, "Any orgs matching the regex(es) specified will be excluded")
	importCmd.PersistentFlags().BoolVar(&cfg.IgnoreServiceKeys, "ignore-service-keys", cfg.IgnoreServiceKeys, "Don't create any service keys on import")
	importCmd.PersistentFlags().StringVar(&cfg.ExportDir, "import-dir", cfg.ExportDir, "Directory where service instances will be placed or read")
	importCmd.PersistentFlags().StringToStringVar(&cfg.DomainsToReplace, "domains-to-replace", cfg.DomainsToReplace, "Domains to replace in any found application routes")

	importOrgCmd := CreateImportOrgCommand(ctx, cfg, factory, sii, fs, reportSummary)
	importCmd.AddCommand(importOrgCmd)

	importSpaceCmd := CreateImportSpaceCommand(ctx, cfg, factory, sii, fs, reportSummary)
	importSpaceCmd.Flags().StringP("org", "o", "", "Org to which the space belongs")
	err := importSpaceCmd.MarkFlagRequired("org")
	if err != nil {
		log.Fatalln(err)
	}
	importCmd.AddCommand(importSpaceCmd)

	rootCmd.AddCommand(importCmd)
}
