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
	"os"
)

func CreateExportOrgCommand(ctx context.Context, config *Config, f ExporterFactory, e migrate.ServiceInstanceExporter, d io.FileSystemOperations, s *report.Summary) *cobra.Command {
	var exportOrg = &cobra.Command{
		Use:     "org",
		Aliases: []string{"o"},
		Short:   "Export org",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires the name of the org to export")
			}
			if args[0] == "" {
				return errors.New("requires a valid name of the org to export")
			}
			if len(args) > 1 {
				return errors.New("too many arguments passed in. only the name of the org is required")
			}
			return nil
		},
		Example: "service-instance-migrator export org sample-org",
		RunE:    exportOrg(ctx, config, f, e, d, s),
	}
	return exportOrg
}

func exportOrg(ctx context.Context, cfg *Config, f ExporterFactory, e migrate.ServiceInstanceExporter, d io.FileSystemOperations, s *report.Summary) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		nonInteractive, err := cmd.Flags().GetBool("non-interactive")
		if err != nil {
			return err
		}

		if !nonInteractive {
			if empty, err := d.IsEmpty(cfg.ExportDir); !empty {
				if err != nil {
					return err
				}
				if c, err := ConfirmYesOrNo("Export directory is not empty. Do you wish to continue?", os.Stdin); !c {
					return err
				}
			}
		}

		if err := d.Mkdir(cfg.ExportDir); err != nil {
			return fmt.Errorf("failed to create export dir: %v: %w", cfg.ExportDir, err)
		}

		defer s.Display()
		org := args[0]

		p := mpb.New(mpb.WithWidth(64))
		ctx = ContextWithProgress(ctx, p)

		if err := f.NewOrgExporter(e).Export(ContextWithSummary(ctx, s), cfg.Foundations.Source, cfg.ExportDir, org); err != nil {
			if cfclient.IsOrganizationNotFoundError(err) {
				return fmt.Errorf("organization %q could not be found", org)
			}
			return err
		}

		return nil
	}
}
