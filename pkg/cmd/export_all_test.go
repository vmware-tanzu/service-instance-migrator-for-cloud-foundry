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

func TestExport(t *testing.T) {
	type args struct {
		commandArgs             []string
		fsOperations            *iofakes.FakeFileSystemOperations
		config                  *config.Config
		exportMigratorFactory   *fakes.FakeExporterFactory
		serviceInstanceMigrator migrate.ServiceInstanceExporter
		reportSummary           *report.Summary
		orgExporter             *fakes.FakeOrgExporter
	}
	tests := []struct {
		name       string
		args       args
		beforeFunc func(args args)
		afterFunc  func(args args)
	}{
		{
			name: "export all orgs",
			args: args{
				fsOperations:          new(iofakes.FakeFileSystemOperations),
				orgExporter:           new(fakes.FakeOrgExporter),
				exportMigratorFactory: new(fakes.FakeExporterFactory),
				commandArgs:           []string{},
				config: &config.Config{
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
			beforeFunc: func(args args) {
				args.fsOperations.IsEmptyStub = func(s string) (bool, error) {
					return true, nil
				}
				args.orgExporter.ExportAllCalls(func(ctx context.Context, om config.OpsManager, dir string) error {
					require.Equal(t, "opsman.source.url.com", om.Hostname)
					require.Equal(t, "/path/to/export-dir", dir)
					return nil
				})
				args.exportMigratorFactory.NewOrgExporterStub = func(exporter migrate.ServiceInstanceExporter) cmd.OrgExporter {
					return args.orgExporter
				}
			},
			afterFunc: func(args args) {
				require.Equal(t, 1, args.fsOperations.MkdirCallCount())
				require.Equal(t, 1, args.exportMigratorFactory.NewOrgExporterCallCount())
				require.Equal(t, 0, args.orgExporter.ExportCallCount())
				require.Equal(t, 1, args.orgExporter.ExportAllCallCount())
			},
		},
		{
			name: "non-interactive does not prompt for input",
			args: args{
				fsOperations:          new(iofakes.FakeFileSystemOperations),
				orgExporter:           new(fakes.FakeOrgExporter),
				exportMigratorFactory: new(fakes.FakeExporterFactory),
				commandArgs:           []string{"--non-interactive"},
				config: &config.Config{
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(args args) {
				args.fsOperations.MkdirStub = func(s string) error {
					return nil
				}
				args.exportMigratorFactory.NewOrgExporterStub = func(exporter migrate.ServiceInstanceExporter) cmd.OrgExporter {
					return args.orgExporter
				}
			},
			afterFunc: func(args args) {
				require.Equal(t, 0, args.fsOperations.IsEmptyCallCount())
				require.Equal(t, 1, args.fsOperations.MkdirCallCount())
				require.Equal(t, 1, args.exportMigratorFactory.NewOrgExporterCallCount())
				require.Equal(t, 0, args.orgExporter.ExportCallCount())
				require.Equal(t, 1, args.orgExporter.ExportAllCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.args.config
			exportCmd := cmd.CreateExportCommand(context.TODO(), cfg, tt.args.exportMigratorFactory, tt.args.serviceInstanceMigrator, tt.args.fsOperations, tt.args.reportSummary)
			exportCmd.PersistentFlags().BoolP("non-interactive", "n", false, "Don't ask for user input")
			exportCmd.PersistentFlags().StringVar(&cfg.ExportDir, "export-dir", cfg.ExportDir, "Directory where service instances will be placed or read")
			exportCmd.Flags().BoolP("non-interactive", "n", false, "Don't ask for user input")
			exportCmd.SetArgs(tt.args.commandArgs)

			tt.beforeFunc(tt.args)
			err := exportCmd.Execute()
			if err != nil {
				t.Errorf("exportAll() = %v", err)
			}
			tt.afterFunc(tt.args)
		})
	}
}

func TestExportFlags(t *testing.T) {
	type args struct {
		config                  *config.Config
		reportSummary           *report.Summary
		exportMigratorFactory   *fakes.FakeExporterFactory
		serviceInstanceExporter migrate.ServiceInstanceExporter
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
					ExportDir: "/path/to/export",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				commandArgs:   []string{"-n"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
				fsOperations: &iofakes.FakeFileSystemOperations{
					ExistsStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(importer migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return &fakes.FakeOrgExporter{
							ExportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
			},
			want: &config.Config{
				ExportDir: "/path/to/export",
				SourceApi: config.CloudController{
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
			name: "export-dir flag overrides export_dir from config",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/export",
					SourceApi: config.CloudController{
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
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(importer migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return &fakes.FakeOrgExporter{
							ExportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"--export-dir", "/overridden/path/to/export", "-n"},
			},
			want: &config.Config{
				ExportDir: "/overridden/path/to/export",
				SourceApi: config.CloudController{
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
				config: &config.Config{
					ExcludedOrgs: []string{"some-org"},
					SourceApi: config.CloudController{
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
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(importer migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return &fakes.FakeOrgExporter{
							ExportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"-n"},
			},
			want: &config.Config{
				ExcludedOrgs: []string{"some-org"},
				SourceApi: config.CloudController{
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
				config: &config.Config{
					ExcludedOrgs: []string{"some-other-org"},
					SourceApi: config.CloudController{
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
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(importer migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return &fakes.FakeOrgExporter{
							ExportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"--exclude-orgs", "some-org", "-n"},
			},
			want: &config.Config{
				ExcludedOrgs: []string{"some-org"},
				SourceApi: config.CloudController{
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
				config: &config.Config{
					IncludedOrgs: []string{"some-org"},
					SourceApi: config.CloudController{
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
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(importer migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return &fakes.FakeOrgExporter{
							ExportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"-n"},
			},
			want: &config.Config{
				IncludedOrgs: []string{"some-org"},
				SourceApi: config.CloudController{
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
					SourceApi: config.CloudController{
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
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewOrgExporterStub: func(importer migrate.ServiceInstanceExporter) cmd.OrgExporter {
						return &fakes.FakeOrgExporter{
							ExportStub: func(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
								return nil
							},
						}
					},
				},
				commandArgs: []string{"--include-orgs", "some-org", "-n"},
			},
			want: &config.Config{
				IncludedOrgs: []string{"some-org"},
				SourceApi: config.CloudController{
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
			exportCmd := cmd.CreateExportCommand(context.TODO(), tt.args.config, tt.args.exportMigratorFactory, tt.args.serviceInstanceExporter, tt.args.fsOperations, tt.args.reportSummary)
			exportCmd.SetArgs(tt.args.commandArgs)
			exportCmd.Flags().StringSliceVar(&tt.args.config.IncludedOrgs, "include-orgs", tt.args.config.IncludedOrgs, "Only orgs matching the regex(es) specified will be included")
			exportCmd.Flags().StringSliceVar(&tt.args.config.ExcludedOrgs, "exclude-orgs", tt.args.config.ExcludedOrgs, "Any orgs matching the regex(es) specified will be excluded")
			exportCmd.PersistentFlags().StringVar(&tt.args.config.ExportDir, "export-dir", tt.args.config.ExportDir, "Directory where service instances will be placed or read")
			exportCmd.PersistentFlags().BoolP("non-interactive", "n", false, "Don't ask for user input")
			err := exportCmd.Execute()
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
