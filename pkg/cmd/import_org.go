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

func CreateImportOrgCommand(ctx context.Context, config *Config, f ImporterFactory, i migrate.ServiceInstanceImporter, fso io.FileSystemOperations, s *report.Summary) *cobra.Command {
	var importOrg = &cobra.Command{
		Use:     "org",
		Aliases: []string{"o"},
		Short:   "Import org",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires the name of the org to import")
			}
			if args[0] == "" {
				return errors.New("requires a valid name of the org to import")
			}
			if len(args) > 1 {
				return errors.New("too many arguments passed in. only the name of the org is required")
			}
			return nil
		},
		Example: "service-instance-migrator import org sample-org",
		RunE:    importOrg(ctx, config, f, i, fso, s),
	}
	return importOrg
}

func importOrg(ctx context.Context, cfg *Config, f ImporterFactory, i migrate.ServiceInstanceImporter, fso io.FileSystemOperations, s *report.Summary) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		org := args[0]

		if exists, _ := fso.Exists(cfg.ExportDir); !exists {
			return fmt.Errorf("import directory %q does not exist", cfg.ExportDir)
		}

		defer s.Display()

		p := mpb.New(mpb.WithWidth(64))
		ctx = ContextWithProgress(ctx, p)

		if err := f.NewOrgImporter(i).Import(ContextWithSummary(ctx, s), cfg.Foundations.Target, cfg.ExportDir, org); err != nil {
			if cfclient.IsOrganizationNotFoundError(err) {
				return fmt.Errorf("organization %q could not be found", org)
			}
			return fmt.Errorf("failed to import services: %w", err)
		}

		return nil
	}
}
