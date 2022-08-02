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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	iofakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func TestImport(t *testing.T) {
	type args struct {
		config                  *config.Config
		commandArgs             []string
		importDir               string
		reportSummary           *report.Summary
		importMigratorFactory   *fakes.FakeImporterFactory
		serviceInstanceImporter migrate.ServiceInstanceImporter
		fsOperations            *iofakes.FakeFileSystemOperations
		orgImporter             *fakes.FakeOrgImporter
	}
	tests := []struct {
		name       string
		args       args
		beforeFunc func(args args)
		afterFunc  func(args args)
	}{
		{
			name: "import all orgs",
			args: args{
				orgImporter:           new(fakes.FakeOrgImporter),
				importMigratorFactory: new(fakes.FakeImporterFactory),
				fsOperations:          new(iofakes.FakeFileSystemOperations),
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
				},
				commandArgs:   []string{"--import-dir", "/path/to/import-dir"},
				importDir:     "/path/to/import-dir",
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(args args) {
				args.fsOperations.ExistsStub = func(s string) (bool, error) {
					return true, nil
				}
				args.importMigratorFactory.NewOrgImporterStub = func(importer migrate.ServiceInstanceImporter) cmd.OrgImporter {
					return args.orgImporter
				}
				args.orgImporter.ImportStub = func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
					return nil
				}
				args.orgImporter.ImportAllCalls(func(ctx context.Context, om config.OpsManager, dir string) error {
					require.Equal(t, "opsman.target.url.com", om.Hostname)
					require.Equal(t, "/path/to/import-dir", dir)
					return nil
				})
			},
			afterFunc: func(args args) {
				require.Equal(t, 0, args.orgImporter.ImportCallCount())
				require.Equal(t, 1, args.orgImporter.ImportAllCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importCmd := cmd.CreateImportCommand(context.TODO(), tt.args.config, tt.args.importMigratorFactory, tt.args.serviceInstanceImporter, tt.args.fsOperations, tt.args.reportSummary)
			importCmd.PersistentFlags().StringVar(&tt.args.config.ExportDir, "import-dir", tt.args.config.ExportDir, "Directory where service instances will be placed or read")
			importCmd.SetArgs(tt.args.commandArgs)

			tt.beforeFunc(tt.args)
			err := importCmd.Execute()
			if err != nil {
				t.Errorf("importAll() = %v", err)
			}
			tt.afterFunc(tt.args)
		})
	}
}

func TestImportFlags(t *testing.T) {
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
			name: "export_dir from config is set when flag is not given",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/import",
					TargetApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				commandArgs:   []string{},
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
				ExportDir: "/path/to/import",
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
			name: "import-dir flag overrides export_dir from config",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/import",
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
				commandArgs: []string{"--import-dir", "/overridden/path/to/import"},
			},
			want: &config.Config{
				ExportDir: "/overridden/path/to/import",
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
				commandArgs:   []string{},
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
				commandArgs: []string{"--ignore-service-keys"},
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
			name: "exclude_orgs from config is set when flag is not given",
			args: args{
				commandArgs: []string{},
				config: &config.Config{
					ExcludedOrgs: []string{"some-org"},
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
			},
			want: &config.Config{
				ExcludedOrgs: []string{"some-org"},
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
			name: "exclude-orgs flag overrides exclude_orgs from config",
			args: args{
				commandArgs: []string{"--exclude-orgs", "some-org"},
				config: &config.Config{
					ExcludedOrgs: []string{"some-other-org"},
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
			},
			want: &config.Config{
				ExcludedOrgs: []string{"some-org"},
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
			name: "include_orgs from config is set when flag is not given",
			args: args{
				commandArgs: []string{},
				config: &config.Config{
					IncludedOrgs: []string{"some-org"},
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
			},
			want: &config.Config{
				IncludedOrgs: []string{"some-org"},
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
			name: "include-orgs flag overrides include_orgs from config",
			args: args{
				config: &config.Config{
					IncludedOrgs: []string{"some-other-org"},
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
				commandArgs: []string{"--include-orgs", "some-org"},
			},
			want: &config.Config{
				IncludedOrgs: []string{"some-org"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importCmd := cmd.CreateImportCommand(context.TODO(), tt.args.config, tt.args.importMigratorFactory, tt.args.serviceInstanceImporter, tt.args.fsOperations, tt.args.reportSummary)
			importCmd.SetArgs(tt.args.commandArgs)
			importCmd.PersistentFlags().StringVar(&tt.args.config.ExportDir, "import-dir", tt.args.config.ExportDir, "Directory where service instances will be placed or read")
			importCmd.PersistentFlags().BoolVar(&tt.args.config.IgnoreServiceKeys, "ignore-service-keys", tt.args.config.IgnoreServiceKeys, "Don't create any service keys on import")
			importCmd.Flags().StringSliceVar(&tt.args.config.IncludedOrgs, "include-orgs", tt.args.config.IncludedOrgs, "Only orgs matching the regex(es) specified will be included")
			importCmd.Flags().StringSliceVar(&tt.args.config.ExcludedOrgs, "exclude-orgs", tt.args.config.ExcludedOrgs, "Any orgs matching the regex(es) specified will be excluded")
			err := importCmd.Execute()
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
