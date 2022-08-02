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

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	iofakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	migratefakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func TestExportSpace(t *testing.T) {
	type args struct {
		config                  *config.Config
		commandArgs             []string
		fsOperations            *iofakes.FakeFileSystemOperations
		reportSummary           *report.Summary
		exportMigratorFactory   *fakes.FakeExporterFactory
		serviceInstanceMigrator *migratefakes.FakeServiceInstanceExporter
	}
	tests := []struct {
		name       string
		args       args
		wantErr    error
		beforeFunc func(*args) (*fakes.FakeSpaceExporter, *iofakes.FakeFileSystemOperations)
		afterFunc  func(*config.Config, *fakes.FakeSpaceExporter, *iofakes.FakeFileSystemOperations)
	}{
		{
			name: "executes command to export space",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{Hostname: "opsman.source.url.com"},
						Target: config.OpsManager{Hostname: "opsman.target.url.com"},
					},
				},
				fsOperations: &iofakes.FakeFileSystemOperations{
					IsEmptyStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				commandArgs:   []string{"space", "some-space", "-o", "some-org"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(args *args) (*fakes.FakeSpaceExporter, *iofakes.FakeFileSystemOperations) {
				spaceExporterStub := new(fakes.FakeSpaceExporter)
				args.exportMigratorFactory = &fakes.FakeExporterFactory{
					NewSpaceExporterStub: func(exporter migrate.ServiceInstanceExporter) cmd.SpaceExporter {
						return spaceExporterStub
					},
				}
				spaceExporterStub.ExportCalls(func(ctx context.Context, om config.OpsManager, dir string, org, space string) error {
					require.Equal(t, "opsman.source.url.com", om.Hostname)
					require.Equal(t, "/path/to/export-dir", dir)
					require.Equal(t, "some-org", org)
					require.Equal(t, "some-space", space)
					return nil
				})
				return spaceExporterStub, nil
			},
			afterFunc: func(cfg *config.Config, spaceExporterStub *fakes.FakeSpaceExporter, w *iofakes.FakeFileSystemOperations) {
				require.Equal(t, 1, spaceExporterStub.ExportCallCount())
				require.False(t, cfg.Debug)
				require.False(t, cfg.DryRun)
			},
		},
		{
			name: "non-interactive does not prompt for input",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				commandArgs: []string{"space", "some-space", "-o", "some-org", "--non-interactive"},
				fsOperations: &iofakes.FakeFileSystemOperations{
					MkdirStub: func(s string) error {
						return nil
					},
				},
				exportMigratorFactory: &fakes.FakeExporterFactory{
					NewSpaceExporterStub: func(exporter migrate.ServiceInstanceExporter) cmd.SpaceExporter {
						return &fakes.FakeSpaceExporter{}
					},
				},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(args *args) (*fakes.FakeSpaceExporter, *iofakes.FakeFileSystemOperations) {
				return new(fakes.FakeSpaceExporter), args.fsOperations
			},
			afterFunc: func(c *config.Config, s *fakes.FakeSpaceExporter, w *iofakes.FakeFileSystemOperations) {
				require.Equal(t, 0, w.IsEmptyCallCount())
			},
		},
		{
			name: "command line flags override config values",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				fsOperations: &iofakes.FakeFileSystemOperations{
					IsEmptyStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				commandArgs:   []string{"space", "some-space", "-o", "some-org", "--debug", "--dry-run", "--export-dir", "/diff/path/to/export-dir"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			beforeFunc: func(args *args) (*fakes.FakeSpaceExporter, *iofakes.FakeFileSystemOperations) {
				spaceExporter := new(fakes.FakeSpaceExporter)
				spaceExporter.ExportStub = func(ctx context.Context, om config.OpsManager, dir string, org, space string) error {
					require.Equal(t, "/diff/path/to/export-dir", dir)
					return nil
				}
				args.exportMigratorFactory = &fakes.FakeExporterFactory{
					NewSpaceExporterStub: func(exporter migrate.ServiceInstanceExporter) cmd.SpaceExporter {
						return spaceExporter
					},
				}
				args.fsOperations.MkdirStub = func(s string) error {
					require.Equal(t, "/diff/path/to/export-dir", s)
					return nil
				}
				return spaceExporter, nil
			},
			afterFunc: func(cfg *config.Config, spaceExporterStub *fakes.FakeSpaceExporter, w *iofakes.FakeFileSystemOperations) {
				require.Equal(t, "/diff/path/to/export-dir", cfg.ExportDir)
				require.True(t, cfg.Debug)
				require.True(t, cfg.DryRun)
			},
		},
		{
			name: "too many args",
			args: args{
				config: &config.Config{
					ExportDir: "/path/to/export-dir",
					SourceApi: config.CloudController{
						URL:      "https://api.cf.example.com",
						Username: "some-user",
						Password: "some-password",
					},
				},
				fsOperations: &iofakes.FakeFileSystemOperations{
					IsEmptyStub: func(s string) (bool, error) {
						return true, nil
					},
				},
				commandArgs:   []string{"space", "some-space", "-o", "some-org", "extra-arg"},
				reportSummary: report.NewSummary(&bytes.Buffer{}),
			},
			wantErr: fmt.Errorf("too many arguments passed in. only the name of the org is required"),
			beforeFunc: func(args *args) (*fakes.FakeSpaceExporter, *iofakes.FakeFileSystemOperations) {
				return new(fakes.FakeSpaceExporter), args.fsOperations
			},
			afterFunc: func(c *config.Config, s *fakes.FakeSpaceExporter, w *iofakes.FakeFileSystemOperations) {
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spaceExporterStub, fsOperationsStub := tt.beforeFunc(&tt.args)
			cfg := tt.args.config
			rootCmd := &cobra.Command{}
			rootCmd.PersistentFlags().StringVar(&cfg.ExportDir, "export-dir", cfg.ExportDir, "Directory where service instances will be placed or read")
			rootCmd.PersistentFlags().BoolP("non-interactive", "n", false, "Don't ask for user input")
			exportSpaceCmd := cmd.CreateExportSpaceCommand(context.TODO(), tt.args.config, tt.args.exportMigratorFactory, tt.args.serviceInstanceMigrator, tt.args.fsOperations, tt.args.reportSummary)
			exportSpaceCmd.Flags().StringP("org", "o", "", "Org to which the space belongs")
			exportSpaceCmd.Flags().BoolP("non-interactive", "n", false, "Don't ask for user input")
			rootCmd.AddCommand(exportSpaceCmd)
			rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", cfg.Debug, "Enable debug logging")
			rootCmd.PersistentFlags().BoolVar(&cfg.DryRun, "dry-run", cfg.DryRun, "Display command without executing")
			rootCmd.SetArgs(tt.args.commandArgs)
			err := rootCmd.Execute()
			if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			tt.afterFunc(cfg, spaceExporterStub, fsOperationsStub)
		})
	}
}
