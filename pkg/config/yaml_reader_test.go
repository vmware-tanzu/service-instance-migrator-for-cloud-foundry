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

package config_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewReader_GetMigration(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name      string
		cfg       *config.Config
		getReader func(cfg *config.Config) *config.YAMLMigrationReader
		want      *config.Migration
		wantErr   bool
	}{
		{
			name: "creates a migration config with a fully configured migrator from default migration reader",
			cfg:  config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-sqlserver-migrator.yml")),
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &config.Migration{
				Migrators: []config.Migrator{
					{
						Name: "sqlserver",
						Value: map[string]interface{}{
							"source_ccdb": map[interface{}]interface{}{
								"db_host":           "192.168.11.24",
								"db_username":       "tas1_ccdb_username",
								"db_password":       "tas1_ccdb_password",
								"db_encryption_key": "tas1_ccdb_enc_key",
								"ssh_host":          "opsman1.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om1_rsa_key",
								"ssh_tunnel":        true,
							},
							"target_ccdb": map[interface{}]interface{}{
								"db_host":           "192.168.12.24",
								"db_username":       "tas2_ccdb_username",
								"db_password":       "tas2_ccdb_password",
								"db_encryption_key": "tas2_ccdb_enc_key",
								"ssh_host":          "opsman2.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om2_rsa_key",
								"ssh_tunnel":        true,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "creates a migration config with an almost empty configured migrator from default migration reader",
			cfg:  config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-almost-empty-migrator.yml")),
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &config.Migration{
				Migrators: []config.Migrator{
					{
						Name: "sqlserver",
						Value: map[string]interface{}{
							"source_ccdb": map[interface{}]interface{}{
								"ssh_tunnel": false,
							},
							"target_ccdb": map[interface{}]interface{}{
								"ssh_tunnel": false,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "creates a migration config with a completely empty configured migrator from default migration reader",
			cfg:  config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-empty-migrator.yml")),
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &config.Migration{
				Migrators: []config.Migrator{{Name: "sqlserver"}},
			},
			wantErr: false,
		},
		{
			name: "creates a migration config with a fully configured migrator from default migration reader -- no dir",
			cfg: &config.Config{
				Migration: config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"source_ccdb": map[string]interface{}{
									"db_host":           "192.168.12.24",
									"db_username":       "tas1_ccdb_username",
									"db_password":       "tas1_ccdb_password",
									"db_encryption_key": "tas1_ccdb_enc_key",
									"ssh_host":          "opsman1.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om2_rsa_key",
									"ssh_tunnel":        true,
								},
							},
						},
					},
				},
			},
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &config.Migration{
				Migrators: []config.Migrator{
					{
						Name: "sqlserver",
						Value: map[string]interface{}{
							"source_ccdb": map[string]interface{}{
								"db_host":           "192.168.12.24",
								"db_username":       "tas1_ccdb_username",
								"db_password":       "tas1_ccdb_password",
								"db_encryption_key": "tas1_ccdb_enc_key",
								"ssh_host":          "opsman1.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om2_rsa_key",
								"ssh_tunnel":        true,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "creates a migration config with an empty configured migrator from default migration reader -- no dir",
			cfg: &config.Config{
				Migration: config.Migration{
					Migrators: []config.Migrator{
						{
							Name:  "sqlserver",
							Value: map[string]interface{}{},
						},
					},
				},
			},
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &config.Migration{
				Migrators: []config.Migrator{
					{
						Name:  "sqlserver",
						Value: map[string]interface{}{},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.getReader(tt.cfg).GetMigration()
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

func TestNewReader_ReadFrom(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "reads a migration config from reader",
			wantErr: false,
			want: `---
foundations:
  source:
    password: fake-password
    url: https://opsman.source.example.com
    client_id: "fake-client-id"
    client_secret: "fake-client-secret"
    username: fake-user
  target:
    password: fake-password
    url: https://opsman.target.example.com
    client_id: "fake-client-id"
    client_secret: "fake-client-secret"
    username: fake-user
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := config.ReadFrom(strings.NewReader(filepath.Join(pwd, "testdata")))
			require.NoError(t, err)
			got, err := r.GetString()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMigration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, got, tt.want)
		})
	}
}

func TestNewReader_ReadFromDir(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "reads a migration config from path",
			wantErr: false,
			want: `---
foundations:
  source:
    password: fake-password
    url: https://opsman.source.example.com
    client_id: "fake-client-id"
    client_secret: "fake-client-secret"
    username: fake-user
  target:
    password: fake-password
    url: https://opsman.target.example.com
    client_id: "fake-client-id"
    client_secret: "fake-client-secret"
    username: fake-user
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := config.ReadFromDir(filepath.Join(pwd, "testdata"))
			require.NoError(t, err)
			got, err := r.GetString()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMigration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, got, tt.want)
		})
	}
}
