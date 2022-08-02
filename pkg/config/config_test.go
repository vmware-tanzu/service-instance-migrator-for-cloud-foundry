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

package config

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name       string
		want       *Config
		wantErr    bool
		beforeFunc func()
	}{
		{
			name: "creates a valid config when env var for config dir is set",
			want: &Config{
				ConfigDir:  filepath.Join(pwd, "testdata"),
				ConfigFile: filepath.Join(pwd, "testdata", "si-migrator.yml"),
				Name:       "si-migrator",
				DomainsToReplace: map[string]string{
					"apps.cf1.example.com": "apps.cf2.example.com",
				},
				ExportDir:    "service-export",
				ExcludedOrgs: []string{"org1", "org2"},
				SourceApi: CloudController{
					URL:          "https://api.cf1.example.com",
					Username:     "cf1-api-username",
					Password:     "cf1-api-password",
					ClientID:     "cf1-api-client",
					ClientSecret: "cf1-api-client-secret",
				},
				TargetApi: CloudController{
					URL:          "https://api.cf2.example.com",
					Username:     "cf2-api-username",
					Password:     "cf2-api-password",
					ClientID:     "cf2-api-client",
					ClientSecret: "cf2-api-client-secret",
				},
				Migration: Migration{},
				Foundations: struct {
					Source OpsManager `yaml:"source"`
					Target OpsManager `yaml:"target"`
				}{
					Source: OpsManager{
						URL:          "https://opsman.source.example.com",
						Username:     "fake-user",
						Password:     "fake-password",
						ClientID:     "fake-client-id",
						ClientSecret: "fake-client-secret",
						Hostname:     "opsman1.example.com",
						IP:           "1.1.1.1",
						PrivateKey:   "/path/to/om1_rsa_key",
						SshUser:      "ubuntu",
					},
					Target: OpsManager{
						URL:          "https://opsman.target.example.com",
						Username:     "fake-user",
						Password:     "fake-password",
						ClientID:     "fake-client-id",
						ClientSecret: "fake-client-secret",
						Hostname:     "opsman2.example.com",
						IP:           "1.1.1.2",
						PrivateKey:   "/path/to/om2_rsa_key",
						SshUser:      "ubuntu",
					},
				},
				DryRun:      false,
				Debug:       false,
				initialized: true,
			},
			wantErr: false,
			beforeFunc: func() {
				err = os.Setenv("SI_MIGRATOR_CONFIG_HOME", filepath.Join(pwd, "testdata"))
				require.NoError(t, err)
			},
		},
		{
			name: "creates a valid config when config dir is a file",
			want: &Config{
				ConfigDir:  filepath.Join(pwd, "testdata") + "/",
				ConfigFile: filepath.Join(pwd, "testdata", "config_no_orgs.yml"),
				Name:       "si-migrator",
				DomainsToReplace: map[string]string{
					"apps.cf1.example.com": "apps.cf2.example.com",
				},
				ExportDir:    filepath.Join(pwd, "export"),
				ExcludedOrgs: []string{},
				DryRun:       false,
				Debug:        false,
				initialized:  true,
			},
			wantErr: false,
			beforeFunc: func() {
				err = os.Setenv("SI_MIGRATOR_CONFIG_HOME", filepath.Join(pwd, "testdata", "config_no_orgs.yml"))
				require.NoError(t, err)
			},
		},
		{
			name: "creates a valid config when file location is specified",
			want: &Config{
				ConfigDir:  "",
				ConfigFile: filepath.Join(pwd, "testdata", "config_no_orgs.yml"),
				Name:       "si-migrator",
				DomainsToReplace: map[string]string{
					"apps.cf1.example.com": "apps.cf2.example.com",
				},
				ExportDir:    filepath.Join(pwd, "export"),
				ExcludedOrgs: []string{},
				DryRun:       false,
				Debug:        false,
				initialized:  true,
			},
			wantErr: false,
			beforeFunc: func() {
				err = os.Setenv("SI_MIGRATOR_CONFIG_FILE", filepath.Join(pwd, "testdata", "config_no_orgs.yml"))
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		_ = os.Unsetenv("SI_MIGRATOR_CONFIG_FILE")
		_ = os.Unsetenv("SI_MIGRATOR_CONFIG_HOME")
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			got, err := NewDefaultConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got,
				cmpopts.IgnoreUnexported(Config{}),
			); diff != "" {
				t.Errorf("NewMigrator() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	type args struct {
		configFile string
		configDir  string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "creates a full migration config with all the supported migrators",
			args: args{
				configFile: filepath.Join(pwd, "testdata", "si-migrator-full.yml"),
			},
			want: &Config{
				ConfigDir:  "",
				ConfigFile: filepath.Join(pwd, "testdata", "si-migrator-full.yml"),
				Foundations: struct {
					Source OpsManager `yaml:"source"`
					Target OpsManager `yaml:"target"`
				}{
					Source: OpsManager{
						URL:          "https://opsman.source.example.com",
						Username:     "fake-user",
						Password:     "fake-password",
						ClientID:     "fake-client-id",
						ClientSecret: "fake-client-secret",
						Hostname:     "opsman1.example.com",
						IP:           "1.1.1.1",
						PrivateKey:   "/path/to/om1_rsa_key",
						SshUser:      "ubuntu",
					},
					Target: OpsManager{
						URL:          "https://opsman.target.example.com",
						Username:     "fake-user",
						Password:     "fake-password",
						ClientID:     "fake-client-id",
						ClientSecret: "fake-client-secret",
						Hostname:     "opsman2.example.com",
						IP:           "1.1.1.2",
						PrivateKey:   "/path/to/om2_rsa_key",
						SshUser:      "ubuntu",
					},
				},
				Migration: Migration{
					UseDefaultMigrator: true,
					Migrators: []Migrator{
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
						{
							Name: "ecs",
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
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":      "minio",
								"backup_directory": "/tmp/mysql-backup",
								"minio": map[interface{}]interface{}{
									"alias":       "ecs-blobstore",
									"url":         "https://object.example.com:9021",
									"access_key":  "blobstore_access_key",
									"secret_key":  "blobstore_secret_key",
									"bucket_name": "mysql-tas1",
									"bucket_path": "p.mysql",
									"insecure":    false,
								},
								"scp": map[interface{}]interface{}{
									"username":              "mysql",
									"hostname":              "mysql-backup.example.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/tmp/backup_rsa_key",
								},
							},
						},
					},
				},
				SourceApi: CloudController{
					URL:          "https://api.cf1.example.com",
					Username:     "cf1-api-username",
					Password:     "cf1-api-password",
					ClientID:     "cf1-api-client",
					ClientSecret: "cf1-api-client-secret",
				},
				TargetApi: CloudController{
					URL:          "https://api.cf2.example.com",
					Username:     "cf2-api-username",
					Password:     "cf2-api-password",
					ClientID:     "cf2-api-client",
					ClientSecret: "cf2-api-client-secret",
				},
				SourceBosh: Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: Authentication{
						Basic: UserCredentials{
							Username: "some-username",
							Password: "some-password",
						},
						UAA: UAAAuthentication{
							URL: "some-url",
							ClientCredentials: ClientCredentials{
								ID:     "some-client-id",
								Secret: "some-client-secret",
							},
							UserCredentials: UserCredentials{
								Username: "some-username",
								Password: "some-password",
							},
						},
					},
				},
				TargetBosh: Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: Authentication{
						Basic: UserCredentials{
							Username: "some-username",
							Password: "some-password",
						},
						UAA: UAAAuthentication{
							URL: "some-url",
							ClientCredentials: ClientCredentials{
								ID:     "some-client-id",
								Secret: "some-client-secret",
							},
							UserCredentials: UserCredentials{
								Username: "some-username",
								Password: "some-password",
							},
						},
					},
				},
				Name: "si-migrator",
				DomainsToReplace: map[string]string{
					"apps.cf1.example.com": "apps.cf2.example.com",
				},
				ExportDir:         "service-export",
				ExcludedOrgs:      []string{"org1", "org2"},
				IgnoreServiceKeys: true,
				DryRun:            false,
				Debug:             false,
				initialized:       true,
			},
			wantErr: false,
		},
		{
			name: "creates a config with viper overrides when config dir is set",
			args: args{
				configDir: filepath.Join(pwd, "testdata"),
			},
			want: &Config{
				ConfigDir:  filepath.Join(pwd, "testdata"),
				ConfigFile: filepath.Join(pwd, "testdata", "si-migrator.yml"),
				Foundations: struct {
					Source OpsManager `yaml:"source"`
					Target OpsManager `yaml:"target"`
				}{
					Source: OpsManager{
						URL:          "https://opsman.source.example.com",
						Username:     "fake-user",
						Password:     "fake-password",
						ClientID:     "fake-client-id",
						ClientSecret: "fake-client-secret",
						Hostname:     "opsman1.example.com",
						IP:           "1.1.1.1",
						PrivateKey:   "/path/to/om1_rsa_key",
						SshUser:      "ubuntu",
					},
					Target: OpsManager{
						URL:          "https://opsman.target.example.com",
						Username:     "fake-user",
						Password:     "fake-password",
						ClientID:     "fake-client-id",
						ClientSecret: "fake-client-secret",
						Hostname:     "opsman2.example.com",
						IP:           "1.1.1.2",
						PrivateKey:   "/path/to/om2_rsa_key",
						SshUser:      "ubuntu",
					},
				},
				SourceApi: CloudController{
					URL:          "https://api.cf1.example.com",
					Username:     "cf1-api-username",
					Password:     "cf1-api-password",
					ClientID:     "cf1-api-client",
					ClientSecret: "cf1-api-client-secret",
				},
				TargetApi: CloudController{
					URL:          "https://api.cf2.example.com",
					Username:     "cf2-api-username",
					Password:     "cf2-api-password",
					ClientID:     "cf2-api-client",
					ClientSecret: "cf2-api-client-secret",
				},
				Migration: Migration{},
				Name:      "si-migrator",
				DomainsToReplace: map[string]string{
					"apps.cf1.example.com": "apps.cf2.example.com",
				},
				ExportDir:    "service-export",
				ExcludedOrgs: []string{"org1", "org2"},
				DryRun:       false,
				Debug:        false,
				initialized:  true,
			},
			wantErr: false,
		},
		{
			name: "creates a config from provided file",
			args: args{
				configFile: filepath.Join(pwd, "testdata", "config_no_orgs.yml"),
				configDir:  filepath.Join(pwd, "testdata"),
			},
			want: &Config{
				ConfigDir:  filepath.Join(pwd, "testdata"),
				ConfigFile: filepath.Join(pwd, "testdata", "config_no_orgs.yml"),
				Name:       "si-migrator",
				DomainsToReplace: map[string]string{
					"apps.cf1.example.com": "apps.cf2.example.com",
				},
				ExportDir:    filepath.Join(pwd, "export"),
				ExcludedOrgs: []string{},
				DryRun:       false,
				Debug:        false,
				initialized:  true,
			},
			wantErr: false,
		},
		{
			name: "creates a valid config with services set",
			args: args{
				configFile: filepath.Join(pwd, "testdata", "config_services.yml"),
			},
			want: &Config{
				ConfigFile:  filepath.Join(pwd, "testdata", "config_services.yml"),
				Name:        "si-migrator",
				ExportDir:   filepath.Join(pwd, "export"),
				Services:    []string{"credhub", "sqlserver"},
				initialized: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.configDir, tt.args.configFile)
			if diff := cmp.Diff(tt.want, got,
				cmpopts.IgnoreUnexported(Config{}),
			); diff != "" {
				t.Errorf("NewMigrator() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
