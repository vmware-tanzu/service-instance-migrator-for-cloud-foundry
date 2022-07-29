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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func TestExportOrg(t *testing.T) {
	stubOrgExporter := new(fakes.FakeOrgExporter)
	type args struct {
		commandArgs []string
	}
	type fields struct {
		cfg                     *config.Config
		fsOperations            *iofakes.FakeFileSystemOperations
		reportSummary           *report.Summary
		exportMigratorFactory   cmd.ExporterFactory
		serviceInstanceMigrator migrate.ServiceInstanceExporter
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantErr    error
		beforeFunc func(fields fields)
		afterFunc  func(fields fields)
	}{
		{
			name: "executes command to export org",
			args: args{
				commandArgs: []string{"some-org"},
			},
			fields: fields{
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(exporter migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return stubOrgExporter
					},
				},
				fsOperations: &iofakes.FakeFileSystemOperations{
					IsEmptyStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{Hostname: "opsman.source.url.com"},
						Target: config.OpsManager{Hostname: "opsman.target.url.com"},
					},
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(fields fields) {
				stubOrgExporter.ExportCalls(func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
					require.Equal(t, "opsman.source.url.com", om.Hostname)
					require.Equal(t, "some-org", orgs[0])
					require.Equal(t, "/path/to/export-dir", dir)
					return nil
				})
			},
			afterFunc: func(fields fields) {
				require.Equal(t, 1, stubOrgExporter.ExportCallCount())
				require.Equal(t, 0, stubOrgExporter.ExportAllCallCount())
			},
		},
		{
			name: "non-interactive does not prompt for input",
			args: args{
				commandArgs: []string{"some-org", "--non-interactive"},
			},
			fields: fields{
				cfg: &config.Config{
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				fsOperations: &iofakes.FakeFileSystemOperations{
					MkdirStub: func(s string) error {
						return nil
					},
				},
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(exporter migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return new(fakes.FakeOrgExporter)
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(fields fields) {},
			afterFunc: func(fields fields) {
				require.Equal(t, 0, fields.fsOperations.IsEmptyCallCount())
			},
		},
		{
			name: "too many args",
			args: args{
				commandArgs: []string{"some-org", "extra-arg"},
			},
			fields: fields{
				cfg: &config.Config{
					IgnoreServiceKeys: false,
					SourceApi: config.CloudController{
						URL: "https://api.cf.example.com",
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(exporter migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return stubOrgExporter
					},
				},
			},
			wantErr: fmt.Errorf("too many arguments passed in. only the name of the org is required"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportOrgCmd := cmd.CreateExportOrgCommand(context.TODO(), tt.fields.cfg, tt.fields.exportMigratorFactory, tt.fields.serviceInstanceMigrator, tt.fields.fsOperations, tt.fields.reportSummary)
			exportOrgCmd.SetArgs(tt.args.commandArgs)
			exportOrgCmd.Flags().BoolP("non-interactive", "n", false, "Don't ask for user input")
			if tt.beforeFunc != nil {
				tt.beforeFunc(tt.fields)
			}
			err := exportOrgCmd.Execute()
			if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			tt.afterFunc(tt.fields)
		})
	}
}
