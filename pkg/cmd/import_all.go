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

	"github.com/spf13/cobra"
	. "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func CreateImportCommand(ctx context.Context, config *Config, f ImporterFactory, i migrate.ServiceInstanceImporter, fso io.FileSystemOperations, s *report.Summary) *cobra.Command {
	export := &cobra.Command{
		Use:   "import",
		Short: "Import service instances from an org or space.",
		Example: `service-instance-migrator import
service-instance-migrator import --exclude-orgs='^system$",p-*'
service-instance-migrator import --exclude-orgs='system,p-spring-cloud-services'
service-instance-migrator import --exclude-orgs='system,p-spring-cloud-services' --import-dir=/tmp
service-instance-migrator import --exclude-orgs='system,p-spring-cloud-services' --import-dir=/tmp --ignore-service-keys
service-instance-migrator import --include-orgs='org1,org2' --import-dir=/tmp`,
		RunE: importAll(ctx, config, f, i, fso, s),
	}
	return export
}

func importAll(ctx context.Context, cfg *Config, f ImporterFactory, i migrate.ServiceInstanceImporter, fso io.FileSystemOperations, s *report.Summary) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if exists, _ := fso.Exists(cfg.ExportDir); !exists {
			return fmt.Errorf("import directory %q does not exist", cfg.ExportDir)
		}

		defer s.Display()

		p := mpb.New(mpb.WithWidth(64))
		ctx = ContextWithProgress(ctx, p)

		if err := f.NewOrgImporter(i).ImportAll(ContextWithSummary(ctx, s), cfg.Foundations.Target, cfg.ExportDir); err != nil {
			return fmt.Errorf("failed to import services: %w", err)
		}

		return nil
	}
}
