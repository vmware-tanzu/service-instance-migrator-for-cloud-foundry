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

package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	iofakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	migratefakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func TestImportSpace(t *testing.T) {
	spaceImporterStub := &fakes.FakeSpaceImporter{
		ImportStub: func(ctx context.Context, om config.OpsManager, dir string, org, space string) error {
			return nil
		},
	}
	type args struct {
		config          *config.Config
		commandArgs     []string
		reportSummary   *report.Summary
		importer        *migratefakes.FakeServiceInstanceImporter
		migratorFactory *fakes.FakeImporterFactory
		fsOperations    *iofakes.FakeFileSystemOperations
	}
	tests := []struct {
		name       string
		args       args
		wantErr    error
		beforeFunc func()
		afterFunc  func()
	}{
		{
			name: "executes command to import space",
			args: args{
				config: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{Hostname: "opsman.source.url.com"},
						Target: config.OpsManager{Hostname: "opsman.target.url.com"},
					},
					TargetApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
					ExportDir: "/path/to/import-dir",
				},
				migratorFactory: &fakes.FakeImporterFactory{
					NewSpaceImporterStub: func(importer migrate.ServiceInstanceImporter) cmd.SpaceImporter {
						return spaceImporterStub
					},
				},
				commandArgs:   []string{"some-space", "-o", "some-org"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
			},
			beforeFunc: func() {
				spaceImporterStub.ImportCalls(func(ctx context.Context, om config.OpsManager, dir string, org, space string) error {
					require.Equal(t, "opsman.target.url.com", om.Hostname)
					require.Equal(t, "some-org", org)
					require.Equal(t, "some-space", space)
					require.Equal(t, "/path/to/import-dir", dir)
					return nil
				})
			},
			afterFunc: func() {
				require.Equal(t, 1, spaceImporterStub.ImportCallCount())
			},
		},
		{
			name: "too many args",
			args: args{
				config: &config.Config{
					TargetApi: config.CloudController{
						URL: "https://api.cf.example.com",
					},
					ExportDir: "/path/to/import-dir",
				},
				migratorFactory: &fakes.FakeImporterFactory{
					NewSpaceImporterStub: func(importer migrate.ServiceInstanceImporter) cmd.SpaceImporter {
						return spaceImporterStub
					},
				},
				commandArgs:   []string{"some-space", "-o", "some-org", "extra-arg"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
			},
			wantErr: fmt.Errorf("too many arguments passed in. only the name of the org is required"),
			beforeFunc: func() {
			},
			afterFunc: func() {
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importSpaceCmd := cmd.CreateImportSpaceCommand(context.TODO(), tt.args.config, tt.args.migratorFactory, tt.args.importer, tt.args.fsOperations, tt.args.reportSummary)
			importSpaceCmd.Flags().StringP("org", "o", "", "Org to which the space belongs")
			importSpaceCmd.SetArgs(tt.args.commandArgs)
			tt.beforeFunc()
			err := importSpaceCmd.Execute()
			if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			tt.afterFunc()
		})
	}
}
