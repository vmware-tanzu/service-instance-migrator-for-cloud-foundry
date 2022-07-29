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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	configfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"os"
	"path/filepath"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	boshfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"

	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootFlags(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	type args struct {
		config             *config.Config
		command            *cobra.Command
		commandArgs        []string
		bcf                bosh.ClientFactory
		mr                 *configfakes.FakeMigrationReader
		sourceConfigLoader config.Loader
		targetConfigLoader config.Loader
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		want      *config.Config
		afterFunc func(*testing.T, *config.Config, *config.Config)
	}{
		{
			name: "services from config is set when flag is not given",
			args: args{
				bcf: FakeBoshClientFactory([]string{"192.168.12.24"}),
				config: &config.Config{
					ConfigDir: filepath.Join(cwd, "testdata"),
					Services:  []string{"me"},
				},
				mr:                 new(configfakes.FakeMigrationReader),
				sourceConfigLoader: new(configfakes.FakeLoader),
				targetConfigLoader: new(configfakes.FakeLoader),
				commandArgs:        []string{"fake"},
				command:            NewFakeCommand(),
			},
			want: &config.Config{
				ConfigDir: filepath.Join(cwd, "testdata"),
				Services:  []string{"me"},
			},
			afterFunc: func(t *testing.T, expected *config.Config, actual *config.Config) {
				require.Equal(t, expected, actual)
			},
		},
		{
			name: "services flag overrides services from config",
			args: args{
				bcf: FakeBoshClientFactory([]string{"192.168.12.24"}),
				config: &config.Config{
					ConfigDir: filepath.Join(cwd, "testdata"),
					Services:  []string{"not-me"},
				},
				sourceConfigLoader: new(configfakes.FakeLoader),
				targetConfigLoader: new(configfakes.FakeLoader),
				commandArgs:        []string{"fake", "--services", "only-me"},
				command:            NewFakeCommand(),
			},
			want: &config.Config{
				ConfigDir: filepath.Join(cwd, "testdata"),
				Services:  []string{"only-me"},
			},
			afterFunc: func(t *testing.T, expected *config.Config, actual *config.Config) {
				require.Equal(t, expected, actual)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := cmd.CreateRootCommand(tt.args.config, tt.args.sourceConfigLoader, tt.args.targetConfigLoader)
			rootCmd.AddCommand(tt.args.command)
			rootCmd.SetArgs(tt.args.commandArgs)
			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("rootCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.afterFunc(t, tt.want, tt.args.config)
		})
	}
}

func NewFakeCommand() *cobra.Command {
	return &cobra.Command{
		Use: "fake",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func FakeBoshClientFactory(ips []string) bosh.ClientFactoryFunc {
	return func(url string, allProxy string, trustedCertPEM []byte, certAppender bosh.CertAppender, directorFactory bosh.DirectorFactory, uaaFactory bosh.UAAFactory, boshAuth config.Authentication) (bosh.Client, error) {
		return &boshfakes.FakeClient{
			FindVMStub: func(s string, s2 string) (director.VMInfo, bool, error) {
				return director.VMInfo{
					IPs: ips,
				}, true, nil
			},
		}, nil
	}
}
