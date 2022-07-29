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

func TestImportOrg(t *testing.T) {
	stubOrgImporter := &fakes.FakeOrgImporter{
		ImportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
			return nil
		},
	}
	type args struct {
		org string
	}
	type fields struct {
		cfg                     *config.Config
		reportSummary           *report.Summary
		importMigratorFactory   *fakes.FakeImporterFactory
		serviceInstanceImporter migrate.ServiceInstanceImporter
		fsOperations            *iofakes.FakeFileSystemOperations
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantErr    bool
		beforeFunc func()
		afterFunc  func()
	}{
		{
			name: "executes command to import org",
			args: args{
				org: "some-org",
			},
			fields: fields{
				cfg: &config.Config{
					ExportDir: "/path/to/import-dir",
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
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				importMigratorFactory: &fakes.FakeImporterFactory{
					NewOrgImporterStub: func(importer migrate.ServiceInstanceImporter) cmd.OrgImporter {
						return stubOrgImporter
					},
				},
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
			},
			wantErr: false,
			beforeFunc: func() {
				stubOrgImporter.ImportCalls(func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
					require.Equal(t, "opsman.target.url.com", om.Hostname)
					require.Equal(t, "/path/to/import-dir", dir)
					require.Equal(t, "some-org", orgs[0])
					return nil
				})
			},
			afterFunc: func() {
				require.Equal(t, 1, stubOrgImporter.ImportCallCount())
				require.Equal(t, 0, stubOrgImporter.ImportAllCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importOrgCmd := cmd.CreateImportOrgCommand(context.TODO(), tt.fields.cfg, tt.fields.importMigratorFactory, tt.fields.serviceInstanceImporter, tt.fields.fsOperations, tt.fields.reportSummary)
			importOrgCmd.SetArgs([]string{tt.args.org})
			tt.beforeFunc()
			if err := importOrgCmd.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.afterFunc()
		})
	}
}

func TestImportOrgFlags(t *testing.T) {
	type args struct {
		config                  *config.Config
		reportSummary           *report.Summary
		importMigratorFactory   *fakes.FakeImporterFactory
		serviceInstanceImporter migrate.ServiceInstanceImporter
		commandArgs             []string
		fsOperations            *iofakes.FakeFileSystemOperations
	}
	tests := []struct {
		name      string
		args      args
		wantErr   error
		want      *config.Config
		afterFunc func(*testing.T, *config.Config, *config.Config)
	}{
		{
			name: "ignore_service_keys from config is set when flag is not given",
			args: args{
				config: &config.Config{
					IgnoreServiceKeys: true,
					TargetApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				commandArgs:   []string{"some-org"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				importMigratorFactory: &fakes.FakeImporterFactory{
					NewOrgImporterStub: func(importer migrate.ServiceInstanceImporter) cmd.OrgImporter {
						return &fakes.FakeOrgImporter{
							ImportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
			},
			want: &config.Config{
				IgnoreServiceKeys: true,
				TargetApi: config.CloudController{
					URL:      "https://api.cf.example.com",
					Username: "some-user",
					Password: "some-password",
				},
			},
			afterFunc: func(t *testing.T, expected *config.Config, actual *config.Config) {
				require.Equal(t, expected, actual)
			},
		},
		{
			name: "ignore-service-keys flag overrides services from config",
			args: args{
				config: &config.Config{
					IgnoreServiceKeys: false,
					TargetApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				importMigratorFactory: &fakes.FakeImporterFactory{
					NewOrgImporterStub: func(importer migrate.ServiceInstanceImporter) cmd.OrgImporter {
						return &fakes.FakeOrgImporter{
							ImportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"some-org", "--ignore-service-keys"},
			},
			want: &config.Config{
				IgnoreServiceKeys: true,
				TargetApi: config.CloudController{
					URL:      "https://api.cf.example.com",
					Username: "some-user",
					Password: "some-password",
				},
			},
			afterFunc: func(t *testing.T, expected *config.Config, actual *config.Config) {
				require.Equal(t, expected, actual)
			},
		},
		{
			name: "too many args",
			args: args{
				config: &config.Config{
					IgnoreServiceKeys: false,
					TargetApi: config.CloudController{
						URL: "https://api.cf.example.com",
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				importMigratorFactory: &fakes.FakeImporterFactory{
					NewOrgImporterStub: func(importer migrate.ServiceInstanceImporter) cmd.OrgImporter {
						return &fakes.FakeOrgImporter{
							ImportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"some-org", "extra-arg"},
			},
			wantErr: fmt.Errorf("too many arguments passed in. only the name of the org is required"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importOrgCommand := cmd.CreateImportOrgCommand(context.TODO(), tt.args.config, tt.args.importMigratorFactory, tt.args.serviceInstanceImporter, tt.args.fsOperations, tt.args.reportSummary)
			importOrgCommand.SetArgs(tt.args.commandArgs)
			importOrgCommand.PersistentFlags().BoolVar(&tt.args.config.IgnoreServiceKeys, "ignore-service-keys", tt.args.config.IgnoreServiceKeys, "Don't create any service keys on import")
			err := importOrgCommand.Execute()
			if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			tt.afterFunc(t, tt.want, tt.args.config)
		})
	}
}
