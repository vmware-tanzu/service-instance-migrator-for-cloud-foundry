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
	"errors"
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v7"
	. "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func CreateImportSpaceCommand(ctx context.Context, config *Config, f ImporterFactory, i migrate.ServiceInstanceImporter, fso io.FileSystemOperations, summary *report.Summary) *cobra.Command {
	var importSpace = &cobra.Command{
		Use:     "space",
		Aliases: []string{"s"},
		Short:   "Import space",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires the name of the space to import")
			}
			if len(args) > 1 {
				return errors.New("too many arguments passed in. only the name of the space is required")
			}
			return nil
		},
		Example: `service-instance-migrator import space sample-space -o sample-org
service-instance-migrator import space sample-space --org sample-org --import-dir=./export --debug
service-instance-migrator import space sample-space --org sample-org --import-dir=./export --services="sqlserver,credhub"
service-instance-migrator import space sample-space --org sample-org --import-dir=./export --instances="sql-test,mysqldb"
`,
		RunE: importSpace(ctx, config, f, i, fso, summary),
	}
	return importSpace
}

func importSpace(ctx context.Context, cfg *Config, f ImporterFactory, i migrate.ServiceInstanceImporter, fso io.FileSystemOperations, summary *report.Summary) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		space := args[0]
		org, err := cmd.Flags().GetString("org")
		if err != nil {
			return err
		}

		if exists, _ := fso.Exists(cfg.ExportDir); !exists {
			return fmt.Errorf("import directory %q does not exist", cfg.ExportDir)
		}

		defer summary.Display()

		p := mpb.New(mpb.WithWidth(64))
		ctx = ContextWithProgress(ctx, p)

		if err := f.NewSpaceImporter(i).Import(ContextWithSummary(ctx, summary), cfg.Foundations.Target, cfg.ExportDir, org, space); err != nil {
			if cfclient.IsOrganizationNotFoundError(err) {
				return fmt.Errorf("organization %q could not be found", org)
			}
			if cfclient.IsSpaceNotFoundError(err) {
				return fmt.Errorf("space %q could not be found", space)
			}
			return err
		}

		return nil
	}
}
