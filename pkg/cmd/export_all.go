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
	"github.com/vbauerster/mpb/v7"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func CreateExportCommand(ctx context.Context, config *config.Config, f ExporterFactory, e migrate.ServiceInstanceExporter, fso io.FileSystemOperations, s *report.Summary) *cobra.Command {
	export := &cobra.Command{
		Use:   "export",
		Short: "Export service instances from an org or space.",
		Example: `service-instance-migrator export
service-instance-migrator export --exclude-orgs='^system$",p-*'
service-instance-migrator export --exclude-orgs='system,si-migrator-org'
service-instance-migrator export --exclude-orgs='system,si-migrator-org' --export-dir=/tmp
service-instance-migrator export --include-orgs='org1,org2' --export-dir=/tmp`,
		RunE: exportAll(ctx, config, f, e, fso, s),
	}
	return export
}

func exportAll(ctx context.Context, cfg *config.Config, f ExporterFactory, e migrate.ServiceInstanceExporter, fso io.FileSystemOperations, s *report.Summary) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		nonInteractive, err := cmd.Flags().GetBool("non-interactive")
		if err != nil {
			return err
		}

		if !nonInteractive {
			if empty, err := fso.IsEmpty(cfg.ExportDir); !empty {
				if err != nil {
					return err
				}
				if c, err := ConfirmYesOrNo("Export directory is not empty. Do you wish to continue?", os.Stdin); !c {
					return err
				}
			}
		}

		if err := fso.Mkdir(cfg.ExportDir); err != nil {
			return fmt.Errorf("failed to create export dir: %v: %w", cfg.ExportDir, err)
		}

		defer s.Display()

		p := mpb.New(mpb.WithWidth(64))
		ctx = config.ContextWithProgress(ctx, p)

		err = f.NewOrgExporter(e).ExportAll(config.ContextWithSummary(ctx, s), cfg.Foundations.Source, cfg.ExportDir)
		if err != nil {
			return err
		}

		return nil
	}
}
