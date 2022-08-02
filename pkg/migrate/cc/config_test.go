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

package cc_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name      string
		cfg       *config.Config
		getReader func(cfg *config.Config) *config.YAMLMigrationReader
		want      cc.Config
		wantErr   bool
	}{
		{
			name: "creates a config used for connecting to ccdb",
			cfg:  config.New(filepath.Join(pwd, "testdata"), ""),
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: cc.Config{
				SourceCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "192.168.1.1",
					Username:       "db-tas1-user",
					Password:       "db-tas1-password",
					EncryptionKey:  "enc-tas1-key",
					SSHHost:        "10.10.10.1",
					SSHUsername:    "ssh-tas1-user",
					SSHPassword:    "ssh-tas1-password",
					SSHPrivateKey:  "/path/to/tas1/ssh-key",
					TunnelRequired: false,
				},
				TargetCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "192.168.1.2",
					Username:       "db-tas2-user",
					Password:       "db-tas2-password",
					EncryptionKey:  "enc-tas2-key",
					SSHHost:        "10.10.10.2",
					SSHUsername:    "ssh-tas2-user",
					SSHPassword:    "ssh-tas2-password",
					SSHPrivateKey:  "/path/to/tas2/ssh-key",
					TunnelRequired: false,
				},
			},
			wantErr: false,
		},
		{
			name: "creates a config with an almost empty configured migrator",
			cfg:  config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-almost-empty-migrator.yml")),
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: cc.Config{
				SourceCloudControllerDatabase: cc.DatabaseConfig{
					TunnelRequired: true,
				},
				TargetCloudControllerDatabase: cc.DatabaseConfig{
					TunnelRequired: true,
				},
			},
			wantErr: false,
		},
		{
			name: "creates a cc config with a completely empty configured migrator",
			cfg:  config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-empty-migrator.yml")),
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: cc.Config{
				SourceCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "192.168.1.1",
					Username:       "db-tas1-user",
					Password:       "db-tas1-password",
					EncryptionKey:  "enc-tas1-key",
					SSHHost:        "10.10.10.1",
					SSHUsername:    "ssh-tas1-user",
					SSHPassword:    "ssh-tas1-password",
					SSHPrivateKey:  "/path/to/tas1/ssh-key",
					TunnelRequired: false,
				},
				TargetCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "192.168.1.2",
					Username:       "db-tas2-user",
					Password:       "db-tas2-password",
					EncryptionKey:  "enc-tas2-key",
					SSHHost:        "10.10.10.2",
					SSHUsername:    "ssh-tas2-user",
					SSHPassword:    "ssh-tas2-password",
					SSHPrivateKey:  "/path/to/tas2/ssh-key",
					TunnelRequired: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.getReader(tt.cfg).GetMigration()
			require.NoError(t, err)
			var conf cc.Config
			got := config.NewMapDecoder(conf).Decode(*m, "some_key")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMigration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("GetMigration() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
