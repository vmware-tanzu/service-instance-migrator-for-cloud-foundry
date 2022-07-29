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

package migrate_test

import (
	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	boshfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	configfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	credhubfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigLoader_BuildSourceConfig(t *testing.T) {
	type fields struct {
		cfg       *config.Config
		opsmanCfg config.OpsManager
		opsman    *fakes.FakeClient
		mr        *configfakes.FakeMigrationReader
		ocf       *fakes.FakeClientFactory
		pp        *configfakes.FakePropertiesProvider
	}
	tests := []struct {
		name   string
		fields fields
		want   *config.Config
	}{
		{
			name: "builds cf source config from opsman client calls",
			fields: fields{
				mr:  new(configfakes.FakeMigrationReader),
				ocf: new(fakes.FakeClientFactory),
				pp: &configfakes.FakePropertiesProvider{
					EnvironmentStub: func(builder config.BoshPropertiesBuilder, builder2 config.CFPropertiesBuilder, builder3 config.CCDBPropertiesBuilder) config.EnvProperties {
						return config.EnvProperties{
							BoshProperties: &config.BoshProperties{
								URL:          "https://bosh.url.com",
								AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
								ClientID:     "some-bosh-client-id",
								ClientSecret: "some-bosh-client-secret",
								RootCA: []httpclient.CA{
									{
										CertPEM: "a trustworthy cert",
									},
								},
								Deployment: "",
							},
							CFProperties: &config.CFProperties{
								URL:      "https://api.cf.url.com",
								Username: "some-cf-username",
								Password: "some-cf-password",
							},
							CCDBProperties: &config.CCDBProperties{
								Host:          "192.168.12.24",
								Username:      "tas1_ccdb_username",
								Password:      "tas1_ccdb_password",
								EncryptionKey: "tas1_ccdb_enc_key",
								SSHHost:       "opsman.url.com",
								SSHUsername:   "some-om-ssh-user",
								SSHPrivateKey: "/path/to/ssh_key",
							},
						}
					},
				},
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						UseDefaultMigrator: false,
						Migrators: []config.Migrator{
							{
								Name: "sqlserver",
								Value: map[string]interface{}{
									"source_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "ecs",
								Value: map[string]interface{}{
									"source_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				opsmanCfg: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "some-om-username",
					Password:     "some-om-password",
					ClientID:     "some-om-client-id",
					ClientSecret: "some-om-client-secret",
					Hostname:     "opsman.url.com",
					IP:           "10.0.0.1",
					PrivateKey:   "/path/to/ssh_key",
					SshUser:      "some-om-ssh-user",
				},
				opsman: &fakes.FakeClient{
					BoshEnvironmentStub: func() (string, string, string, error) {
						return "https://bosh.url.com", "some-bosh-client-id", "some-bosh-client-secret", nil
					},
					CertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
						return []httpclient.CA{
							{
								Active:  true,
								CertPEM: "a trustworthy cert",
							},
						}, nil
					},
					DeployedProductStub: func(string) (string, error) {
						return "cf-some-guid", nil
					},
					DeployedProductCredentialsStub: func(string, string) (httpclient.DeployedProductCredential, error) {
						return httpclient.DeployedProductCredential{
							Credential: httpclient.Credential{
								Type: "simple_credentials",
								Value: map[string]string{
									"password": "some-cf-password",
									"identity": "some-cf-username",
								},
							},
						}, nil
					},
					StagedProductPropertiesStub: func(string) (map[string]httpclient.ResponseProperty, error) {
						return map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
							Value:        "cf.url.com",
							Configurable: true,
							IsCredential: false,
						}}, nil
					},
				},
			},
			want: &config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Source: config.OpsManager{
						URL:          "https://opsman.url.com",
						Username:     "some-om-username",
						Password:     "some-om-password",
						ClientID:     "some-om-client-id",
						ClientSecret: "some-om-client-secret",
						Hostname:     "opsman.url.com",
						IP:           "10.0.0.1",
						PrivateKey:   "/path/to/ssh_key",
						SshUser:      "some-om-ssh-user",
					},
				},
				Migration: config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"sqlserver": cc.Config{
									SourceCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas1_ccdb_username",
										Password:       "tas1_ccdb_password",
										EncryptionKey:  "tas1_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "ecs",
							Value: map[string]interface{}{
								"ecs": cc.Config{
									SourceCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas1_ccdb_username",
										Password:       "tas1_ccdb_password",
										EncryptionKey:  "tas1_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":           "scp",
								"username":              "mysql",
								"hostname":              "mysql.backups.com",
								"port":                  22,
								"destination_directory": "/var/vcap/data/mysql/backups",
								"private_key":           "/.ssh/key",
							},
						},
					},
				},
				SourceApi: config.CloudController{
					URL:      "https://api.cf.url.com",
					Username: "some-cf-username",
					Password: "some-cf-password",
				},
				SourceBosh: config.Bosh{
					URL:         "https://bosh.url.com",
					AllProxy:    "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					TrustedCert: "a trustworthy cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "https://bosh.url.com:8443",
							ClientCredentials: config.ClientCredentials{
								ID:     "some-bosh-client-id",
								Secret: "some-bosh-client-secret",
							},
						},
					},
				},
			},
		},
		{
			name: "appends port to bosh url when other than 443/0",
			fields: fields{
				mr:  new(configfakes.FakeMigrationReader),
				ocf: new(fakes.FakeClientFactory),
				pp: &configfakes.FakePropertiesProvider{
					EnvironmentStub: func(builder config.BoshPropertiesBuilder, builder2 config.CFPropertiesBuilder, builder3 config.CCDBPropertiesBuilder) config.EnvProperties {
						return config.EnvProperties{
							BoshProperties: &config.BoshProperties{
								URL:          "https://bosh.url.com:9999",
								AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
								ClientID:     "some-bosh-client-id",
								ClientSecret: "some-bosh-client-secret",
								RootCA: []httpclient.CA{
									{
										CertPEM: "a trustworthy cert",
									},
								},
								Deployment: "",
							},
							CFProperties: &config.CFProperties{
								URL:      "https://api.cf.url.com",
								Username: "some-cf-username",
								Password: "some-cf-password",
							},
							CCDBProperties: &config.CCDBProperties{
								Host:          "192.168.12.24",
								Username:      "tas1_ccdb_username",
								Password:      "tas1_ccdb_password",
								EncryptionKey: "tas1_ccdb_enc_key",
								SSHHost:       "opsman.url.com",
								SSHUsername:   "some-om-ssh-user",
								SSHPrivateKey: "/path/to/ssh_key",
							},
						}
					},
				},
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						UseDefaultMigrator: false,
						Migrators: []config.Migrator{
							{
								Name: "sqlserver",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "ecs",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				opsmanCfg: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "some-om-username",
					Password:     "some-om-password",
					ClientID:     "some-om-client-id",
					ClientSecret: "some-om-client-secret",
					Hostname:     "opsman.url.com",
					IP:           "10.0.0.1",
					PrivateKey:   "/path/to/ssh_key",
					SshUser:      "some-om-ssh-user",
				},
				opsman: &fakes.FakeClient{
					BoshEnvironmentStub: func() (string, string, string, error) {
						return "https://bosh.url.com:9999", "some-bosh-client-id", "some-bosh-client-secret", nil
					},
					CertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
						return []httpclient.CA{
							{
								Active:  true,
								CertPEM: "a trustworthy cert",
							},
						}, nil
					},
					DeployedProductStub: func(string) (string, error) {
						return "cf-some-guid", nil
					},
					DeployedProductCredentialsStub: func(string, string) (httpclient.DeployedProductCredential, error) {
						return httpclient.DeployedProductCredential{
							Credential: httpclient.Credential{
								Type: "simple_credentials",
								Value: map[string]string{
									"password": "some-cf-password",
									"identity": "some-cf-username",
								},
							},
						}, nil
					},
					StagedProductPropertiesStub: func(string) (map[string]httpclient.ResponseProperty, error) {
						return map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
							Value:        "cf.url.com",
							Configurable: true,
							IsCredential: false,
						}}, nil
					},
				},
			},
			want: &config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Source: config.OpsManager{
						URL:          "https://opsman.url.com",
						Username:     "some-om-username",
						Password:     "some-om-password",
						ClientID:     "some-om-client-id",
						ClientSecret: "some-om-client-secret",
						Hostname:     "opsman.url.com",
						IP:           "10.0.0.1",
						PrivateKey:   "/path/to/ssh_key",
						SshUser:      "some-om-ssh-user",
					},
				},
				Migration: config.Migration{
					UseDefaultMigrator: false,
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"sqlserver": cc.Config{
									SourceCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas1_ccdb_username",
										Password:       "tas1_ccdb_password",
										EncryptionKey:  "tas1_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "ecs",
							Value: map[string]interface{}{
								"ecs": cc.Config{
									SourceCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas1_ccdb_username",
										Password:       "tas1_ccdb_password",
										EncryptionKey:  "tas1_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":           "scp",
								"username":              "mysql",
								"hostname":              "mysql.backups.com",
								"port":                  22,
								"destination_directory": "/var/vcap/data/mysql/backups",
								"private_key":           "/.ssh/key",
							},
						},
					},
				},
				SourceApi: config.CloudController{
					URL:      "https://api.cf.url.com",
					Username: "some-cf-username",
					Password: "some-cf-password",
				},
				SourceBosh: config.Bosh{
					URL:         "https://bosh.url.com:9999",
					AllProxy:    "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					TrustedCert: "a trustworthy cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "https://bosh.url.com:8443",
							ClientCredentials: config.ClientCredentials{
								ID:     "some-bosh-client-id",
								Secret: "some-bosh-client-secret",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.fields.cfg
			opsman := tt.fields.opsman
			mr := tt.fields.mr
			mr.GetMigrationReturns(&config.Migration{
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
					{
						Name: "ecs",
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
					{
						Name: "mysql",
						Value: map[string]interface{}{
							"backup_type":           "scp",
							"username":              "mysql",
							"hostname":              "mysql.backups.com",
							"port":                  22,
							"destination_directory": "/var/vcap/data/mysql/backups",
							"private_key":           "/.ssh/key",
						},
					},
				},
			}, nil)
			opsman.DeployedProductCredentialsReturnsOnCall(0, httpclient.DeployedProductCredential{
				Credential: httpclient.Credential{
					Type: "simple_credentials",
					Value: map[string]string{
						"password": "some-cf-password",
						"identity": "some-cf-username",
					},
				},
			}, nil)
			opsman.DeployedProductCredentialsReturnsOnCall(1, httpclient.DeployedProductCredential{
				Credential: httpclient.Credential{
					Type: "simple_credentials",
					Value: map[string]string{
						"password": "tas1_ccdb_enc_key",
						"identity": "db_encryption",
					},
				},
			}, nil)
			ocf := tt.fields.ocf
			ocf.NewReturns(opsman, nil)
			l := migrate.NewConfigLoader(cfg, mr, tt.fields.pp)
			l.BuildSourceConfig()
			if diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("BuildSourceConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigLoader_BuildTargetConfig(t *testing.T) {
	type fields struct {
		cfg       *config.Config
		opsmanCfg config.OpsManager
		opsman    *fakes.FakeClient
		mr        *configfakes.FakeMigrationReader
		ocf       *fakes.FakeClientFactory
		pp        *configfakes.FakePropertiesProvider
	}
	tests := []struct {
		name   string
		fields fields
		want   *config.Config
	}{
		{
			name: "builds cf target config from opsman client calls",
			fields: fields{
				mr:  new(configfakes.FakeMigrationReader),
				ocf: new(fakes.FakeClientFactory),
				pp: &configfakes.FakePropertiesProvider{
					EnvironmentStub: func(builder config.BoshPropertiesBuilder, builder2 config.CFPropertiesBuilder, builder3 config.CCDBPropertiesBuilder) config.EnvProperties {
						return config.EnvProperties{
							BoshProperties: &config.BoshProperties{
								URL:          "https://bosh.url.com",
								AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
								ClientID:     "some-bosh-client-id",
								ClientSecret: "some-bosh-client-secret",
								RootCA: []httpclient.CA{
									{
										CertPEM: "a trustworthy cert",
									},
								},
								Deployment: "",
							},
							CFProperties: &config.CFProperties{
								URL:      "https://api.cf.url.com",
								Username: "some-cf-username",
								Password: "some-cf-password",
							},
							CCDBProperties: &config.CCDBProperties{
								Host:          "192.168.12.24",
								Username:      "tas2_ccdb_username",
								Password:      "tas2_ccdb_password",
								EncryptionKey: "tas2_ccdb_enc_key",
								SSHHost:       "opsman.url.com",
								SSHUsername:   "some-om-ssh-user",
								SSHPrivateKey: "/path/to/ssh_key",
							},
						}
					},
				},
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						UseDefaultMigrator: false,
						Migrators: []config.Migrator{
							{
								Name: "sqlserver",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "ecs",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				opsmanCfg: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "some-om-username",
					Password:     "some-om-password",
					ClientID:     "some-om-client-id",
					ClientSecret: "some-om-client-secret",
					Hostname:     "opsman.url.com",
					IP:           "10.0.0.1",
					PrivateKey:   "/path/to/ssh_key",
					SshUser:      "some-om-ssh-user",
				},
				opsman: &fakes.FakeClient{
					BoshEnvironmentStub: func() (string, string, string, error) {
						return "https://bosh.url.com", "some-bosh-client-id", "some-bosh-client-secret", nil
					},
					CertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
						return []httpclient.CA{
							{
								Active:  true,
								CertPEM: "a trustworthy cert",
							},
						}, nil
					},
					DeployedProductStub: func(string) (string, error) {
						return "cf-some-guid", nil
					},
					DeployedProductCredentialsStub: func(string, string) (httpclient.DeployedProductCredential, error) {
						return httpclient.DeployedProductCredential{
							Credential: httpclient.Credential{
								Type: "simple_credentials",
								Value: map[string]string{
									"password": "some-cf-password",
									"identity": "some-cf-username",
								},
							},
						}, nil
					},
					StagedProductPropertiesStub: func(string) (map[string]httpclient.ResponseProperty, error) {
						return map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
							Value:        "cf.url.com",
							Configurable: true,
							IsCredential: false,
						}}, nil
					},
				},
			},
			want: &config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Target: config.OpsManager{
						URL:          "https://opsman.url.com",
						Username:     "some-om-username",
						Password:     "some-om-password",
						ClientID:     "some-om-client-id",
						ClientSecret: "some-om-client-secret",
						Hostname:     "opsman.url.com",
						IP:           "10.0.0.1",
						PrivateKey:   "/path/to/ssh_key",
						SshUser:      "some-om-ssh-user",
					},
				},
				Migration: config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"sqlserver": cc.Config{
									TargetCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas2_ccdb_username",
										Password:       "tas2_ccdb_password",
										EncryptionKey:  "tas2_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "ecs",
							Value: map[string]interface{}{
								"ecs": cc.Config{
									TargetCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas2_ccdb_username",
										Password:       "tas2_ccdb_password",
										EncryptionKey:  "tas2_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":           "scp",
								"username":              "mysql",
								"hostname":              "mysql.backups.com",
								"port":                  22,
								"destination_directory": "/var/vcap/data/mysql/backups",
								"private_key":           "/.ssh/key",
							},
						},
					},
				},
				TargetApi: config.CloudController{
					URL:      "https://api.cf.url.com",
					Username: "some-cf-username",
					Password: "some-cf-password",
				},
				TargetBosh: config.Bosh{
					URL:         "https://bosh.url.com",
					AllProxy:    "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					TrustedCert: "a trustworthy cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "https://bosh.url.com:8443",
							ClientCredentials: config.ClientCredentials{
								ID:     "some-bosh-client-id",
								Secret: "some-bosh-client-secret",
							},
						},
					},
				},
			},
		},
		{
			name: "appends port to bosh url when other than 443/0",
			fields: fields{
				mr:  new(configfakes.FakeMigrationReader),
				ocf: new(fakes.FakeClientFactory),
				pp: &configfakes.FakePropertiesProvider{
					EnvironmentStub: func(builder config.BoshPropertiesBuilder, builder2 config.CFPropertiesBuilder, builder3 config.CCDBPropertiesBuilder) config.EnvProperties {
						return config.EnvProperties{
							BoshProperties: &config.BoshProperties{
								URL:          "https://bosh.url.com:9999",
								AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
								ClientID:     "some-bosh-client-id",
								ClientSecret: "some-bosh-client-secret",
								RootCA: []httpclient.CA{
									{
										CertPEM: "a trustworthy cert",
									},
								},
								Deployment: "",
							},
							CFProperties: &config.CFProperties{
								URL:      "https://api.cf.url.com",
								Username: "some-cf-username",
								Password: "some-cf-password",
							},
							CCDBProperties: &config.CCDBProperties{
								Host:          "192.168.12.24",
								Username:      "tas2_ccdb_username",
								Password:      "tas2_ccdb_password",
								EncryptionKey: "tas2_ccdb_enc_key",
								SSHHost:       "opsman.url.com",
								SSHUsername:   "some-om-ssh-user",
								SSHPrivateKey: "/path/to/ssh_key",
							},
						}
					},
				},
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						UseDefaultMigrator: false,
						Migrators: []config.Migrator{
							{
								Name: "sqlserver",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "ecs",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				opsmanCfg: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "some-om-username",
					Password:     "some-om-password",
					ClientID:     "some-om-client-id",
					ClientSecret: "some-om-client-secret",
					Hostname:     "opsman.url.com",
					IP:           "10.0.0.1",
					PrivateKey:   "/path/to/ssh_key",
					SshUser:      "some-om-ssh-user",
				},
				opsman: &fakes.FakeClient{
					BoshEnvironmentStub: func() (string, string, string, error) {
						return "https://bosh.url.com:9999", "some-bosh-client-id", "some-bosh-client-secret", nil
					},
					CertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
						return []httpclient.CA{
							{
								Active:  true,
								CertPEM: "a trustworthy cert",
							},
						}, nil
					},
					DeployedProductStub: func(string) (string, error) {
						return "cf-some-guid", nil
					},
					StagedProductPropertiesStub: func(string) (map[string]httpclient.ResponseProperty, error) {
						return map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
							Value:        "cf.url.com",
							Configurable: true,
							IsCredential: false,
						}}, nil
					},
				},
			},
			want: &config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Target: config.OpsManager{
						URL:          "https://opsman.url.com",
						Username:     "some-om-username",
						Password:     "some-om-password",
						ClientID:     "some-om-client-id",
						ClientSecret: "some-om-client-secret",
						Hostname:     "opsman.url.com",
						IP:           "10.0.0.1",
						PrivateKey:   "/path/to/ssh_key",
						SshUser:      "some-om-ssh-user",
					},
				},
				Migration: config.Migration{
					UseDefaultMigrator: false,
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"sqlserver": cc.Config{
									TargetCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas2_ccdb_username",
										Password:       "tas2_ccdb_password",
										EncryptionKey:  "tas2_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "ecs",
							Value: map[string]interface{}{
								"ecs": cc.Config{
									TargetCloudControllerDatabase: cc.DatabaseConfig{
										Host:           "192.168.12.24",
										Username:       "tas2_ccdb_username",
										Password:       "tas2_ccdb_password",
										EncryptionKey:  "tas2_ccdb_enc_key",
										SSHHost:        "opsman.url.com",
										SSHUsername:    "some-om-ssh-user",
										SSHPassword:    "",
										SSHPrivateKey:  "/path/to/ssh_key",
										TunnelRequired: true,
									},
								},
							},
						},
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":           "scp",
								"username":              "mysql",
								"hostname":              "mysql.backups.com",
								"port":                  22,
								"destination_directory": "/var/vcap/data/mysql/backups",
								"private_key":           "/.ssh/key",
							},
						},
					},
				},
				TargetApi: config.CloudController{
					URL:      "https://api.cf.url.com",
					Username: "some-cf-username",
					Password: "some-cf-password",
				},
				TargetBosh: config.Bosh{
					URL:         "https://bosh.url.com:9999",
					AllProxy:    "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					TrustedCert: "a trustworthy cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "https://bosh.url.com:8443",
							ClientCredentials: config.ClientCredentials{
								ID:     "some-bosh-client-id",
								Secret: "some-bosh-client-secret",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.fields.cfg
			opsman := tt.fields.opsman
			mr := tt.fields.mr
			mr.GetMigrationReturns(&config.Migration{
				Migrators: []config.Migrator{
					{
						Name: "sqlserver",
						Value: map[string]interface{}{
							"target_ccdb": map[string]interface{}{
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
							"target_ccdb": map[string]interface{}{
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
							"backup_type":           "scp",
							"username":              "mysql",
							"hostname":              "mysql.backups.com",
							"port":                  22,
							"destination_directory": "/var/vcap/data/mysql/backups",
							"private_key":           "/.ssh/key",
						},
					},
				},
			}, nil)
			opsman.DeployedProductCredentialsReturnsOnCall(0, httpclient.DeployedProductCredential{
				Credential: httpclient.Credential{
					Type: "simple_credentials",
					Value: map[string]string{
						"password": "some-cf-password",
						"identity": "some-cf-username",
					},
				},
			}, nil)
			opsman.DeployedProductCredentialsReturnsOnCall(1, httpclient.DeployedProductCredential{
				Credential: httpclient.Credential{
					Type: "simple_credentials",
					Value: map[string]string{
						"password": "tas2_ccdb_enc_key",
						"identity": "db_encryption",
					},
				},
			}, nil)
			ocf := tt.fields.ocf
			ocf.NewReturns(opsman, nil)
			l := migrate.NewConfigLoader(cfg, mr, tt.fields.pp)
			l.BuildTargetConfig()
			if diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("BuildTargetConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigLoader_BoshConfig(t *testing.T) {
	type fields struct {
		cfg           *config.Config
		opsmanCfg     config.OpsManager
		mr            *configfakes.FakeMigrationReader
		opsmanFactory *fakes.FakeClientFactory
		boshFactory   bosh.ClientFactory
		pp            *configfakes.FakePropertiesProvider
	}
	type args struct {
		toSource bool
	}
	tests := []struct {
		name       string
		opsman     *fakes.FakeClient
		fields     fields
		args       args
		want       *config.Bosh
		beforeFunc func(fields, *fakes.FakeClient)
	}{
		{
			name:       "source bosh config is loaded correctly when config struct is set",
			opsman:     new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {},
			fields: fields{
				cfg: &config.Config{
					SourceBosh: config.Bosh{
						URL:         "https://10.1.0.1",
						AllProxy:    "ssh+socks5://some-user@opsman.example.com:22?private-key=/path/to/ssh-key",
						TrustedCert: "a very trustworthy cert",
						Authentication: config.Authentication{
							UAA: config.UAAAuthentication{
								URL: "https://10.1.0.1:8443",
								ClientCredentials: config.ClientCredentials{
									ID:     "some-client-id",
									Secret: "some-client-secret",
								},
							},
						},
					},
				},
				opsmanCfg: config.OpsManager{
					Hostname:   "opsman.example.com",
					PrivateKey: "/path/to/ssh-key",
					SshUser:    "some-user",
				},
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
			},
			args: args{
				toSource: true,
			},
			want: &config.Bosh{
				URL:         "https://10.1.0.1",
				AllProxy:    "ssh+socks5://some-user@opsman.example.com:22?private-key=/path/to/ssh-key",
				TrustedCert: "a very trustworthy cert",
				Authentication: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://10.1.0.1:8443",
						ClientCredentials: config.ClientCredentials{
							ID:     "some-client-id",
							Secret: "some-client-secret",
						},
					},
				},
			},
		},
		{
			name:   "source bosh config is loaded correctly from opsman when config struct is not set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.BoshEnvironmentReturns("10.1.0.1", "some-client-id", "some-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							Hostname:   "opsman.url.com",
							SshUser:    "some-om-ssh-user",
							PrivateKey: "/path/to/ssh-key",
						},
					},
					Migration:  config.Migration{},
					SourceBosh: config.Bosh{},
				},
				pp: &configfakes.FakePropertiesProvider{
					SourceBoshPropertiesBuilderStub: func() config.BoshPropertiesBuilder {
						return &configfakes.FakeBoshPropertiesBuilder{
							BuildStub: func() *config.BoshProperties {
								return &config.BoshProperties{
									URL:          "https://10.1.0.1",
									AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
									ClientID:     "some-bosh-client-id",
									ClientSecret: "some-bosh-client-secret",
									RootCA: []httpclient.CA{
										{
											CertPEM: "a very trustworthy cert",
										},
									},
									Deployment: "",
								}
							},
						}
					},
				},
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
			},
			args: args{
				toSource: true,
			},
			want: &config.Bosh{
				URL:         "https://10.1.0.1",
				AllProxy:    "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh-key",
				TrustedCert: "a very trustworthy cert",
				Authentication: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://10.1.0.1:8443",
						ClientCredentials: config.ClientCredentials{
							ID:     "some-bosh-client-id",
							Secret: "some-bosh-client-secret",
						},
					},
				},
			},
		},
		{
			name:       "target bosh config is loaded correctly when config struct is set",
			opsman:     new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {},
			fields: fields{
				cfg: &config.Config{
					TargetBosh: config.Bosh{
						URL:         "https://10.1.0.1",
						AllProxy:    "ssh+socks5://some-user@opsman.example.com:22?private-key=/path/to/ssh-key",
						TrustedCert: "a very trustworthy cert",
						Authentication: config.Authentication{
							UAA: config.UAAAuthentication{
								URL: "https://10.1.0.1:8443",
								ClientCredentials: config.ClientCredentials{
									ID:     "some-client-id",
									Secret: "some-client-secret",
								},
							},
						},
					},
				},
				opsmanCfg: config.OpsManager{
					Hostname:   "opsman.example.com",
					PrivateKey: "/path/to/ssh-key",
					SshUser:    "some-user",
				},
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
			},
			args: args{
				toSource: false,
			},
			want: &config.Bosh{
				URL:         "https://10.1.0.1",
				AllProxy:    "ssh+socks5://some-user@opsman.example.com:22?private-key=/path/to/ssh-key",
				TrustedCert: "a very trustworthy cert",
				Authentication: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://10.1.0.1:8443",
						ClientCredentials: config.ClientCredentials{
							ID:     "some-client-id",
							Secret: "some-client-secret",
						},
					},
				},
			},
		},
		{
			name:   "target bosh config is loaded correctly from opsman when config struct is not set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.BoshEnvironmentReturns("10.1.0.1", "some-client-id", "some-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							Hostname:   "opsman.url.com",
							SshUser:    "some-om-ssh-user",
							PrivateKey: "/path/to/ssh-key",
						},
					},
					Migration:  config.Migration{},
					TargetBosh: config.Bosh{},
				},
				pp: &configfakes.FakePropertiesProvider{
					TargetBoshPropertiesBuilderStub: func() config.BoshPropertiesBuilder {
						return &configfakes.FakeBoshPropertiesBuilder{
							BuildStub: func() *config.BoshProperties {
								return &config.BoshProperties{
									URL:          "https://10.1.0.1",
									AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
									ClientID:     "some-bosh-client-id",
									ClientSecret: "some-bosh-client-secret",
									RootCA: []httpclient.CA{
										{
											CertPEM: "a very trustworthy cert",
										},
									},
									Deployment: "",
								}
							},
						}
					},
				},
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
			},
			args: args{
				toSource: false,
			},
			want: &config.Bosh{
				URL:         "https://10.1.0.1",
				AllProxy:    "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh-key",
				TrustedCert: "a very trustworthy cert",
				Authentication: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://10.1.0.1:8443",
						ClientCredentials: config.ClientCredentials{
							ID:     "some-bosh-client-id",
							Secret: "some-bosh-client-secret",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc(tt.fields, tt.opsman)
			l := migrate.NewConfigLoader(
				tt.fields.cfg,
				tt.fields.mr,
				tt.fields.pp,
			)
			if diff := cmp.Diff(tt.want, l.BoshConfig(tt.args.toSource), cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("BoshConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigLoader_CFConfig(t *testing.T) {
	type fields struct {
		cfg           *config.Config
		mr            *configfakes.FakeMigrationReader
		opsmanFactory *fakes.FakeClientFactory
		pp            *configfakes.FakePropertiesProvider
	}
	type args struct {
		toSource bool
	}
	tests := []struct {
		name       string
		opsman     *fakes.FakeClient
		fields     fields
		args       args
		want       *config.CloudController
		beforeFunc func(fields, *fakes.FakeClient)
	}{
		{
			name:       "source cf config is loaded correctly when config struct is set",
			opsman:     new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {},
			fields: fields{
				cfg: &config.Config{
					SourceApi: config.CloudController{
						URL:      "https://api.cf.url.com",
						Username: "some-cf-user",
						Password: "some-cf-password",
					},
				},
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource: true,
			},
			want: &config.CloudController{
				URL:      "https://api.cf.url.com",
				Username: "some-cf-user",
				Password: "some-cf-password",
			},
		},
		{
			name:   "source cf config is loaded correctly from opsman when config struct is not set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.DeployedProductReturns("cf-some-guid", nil)
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "some-cf-password",
							"identity": "some-cf-user",
						},
					},
				}, nil)
				client.StagedProductPropertiesReturns(map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
					Value:        "cf.url.com",
					Configurable: true,
					IsCredential: false,
				}}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
			},
			fields: fields{
				cfg: &config.Config{
					SourceApi: config.CloudController{},
				},
				opsmanFactory: new(fakes.FakeClientFactory),
				pp: &configfakes.FakePropertiesProvider{
					SourceCFPropertiesBuilderStub: func() config.CFPropertiesBuilder {
						return &configfakes.FakeCFPropertiesBuilder{
							BuildStub: func() *config.CFProperties {
								return &config.CFProperties{
									URL:      "https://api.cf.url.com",
									Username: "some-cf-user",
									Password: "some-cf-password",
								}
							},
						}
					},
				},
			},
			args: args{
				toSource: true,
			},
			want: &config.CloudController{
				URL:      "https://api.cf.url.com",
				Username: "some-cf-user",
				Password: "some-cf-password",
			},
		},
		{
			name:       "target cf config is loaded correctly when config struct is set",
			opsman:     new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {},
			fields: fields{
				cfg: &config.Config{
					TargetApi: config.CloudController{
						URL:      "https://api.cf.url.com",
						Username: "some-cf-user",
						Password: "some-cf-password",
					},
				},
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource: false,
			},
			want: &config.CloudController{
				URL:      "https://api.cf.url.com",
				Username: "some-cf-user",
				Password: "some-cf-password",
			},
		},
		{
			name:   "target cf config is loaded correctly from opsman when config struct is not set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.DeployedProductReturns("cf-some-guid", nil)
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "some-cf-password",
							"identity": "some-cf-user",
						},
					},
				}, nil)
				client.StagedProductPropertiesReturns(map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
					Value:        "cf.url.com",
					Configurable: true,
					IsCredential: false,
				}}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
			},
			fields: fields{
				cfg: &config.Config{
					TargetApi: config.CloudController{},
				},
				opsmanFactory: new(fakes.FakeClientFactory),
				pp: &configfakes.FakePropertiesProvider{
					TargetCFPropertiesBuilderStub: func() config.CFPropertiesBuilder {
						return &configfakes.FakeCFPropertiesBuilder{
							BuildStub: func() *config.CFProperties {
								return &config.CFProperties{
									URL:      "https://api.cf.url.com",
									Username: "some-cf-user",
									Password: "some-cf-password",
								}
							},
						}
					},
				},
			},
			args: args{
				toSource: false,
			},
			want: &config.CloudController{
				URL:      "https://api.cf.url.com",
				Username: "some-cf-user",
				Password: "some-cf-password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc(tt.fields, tt.opsman)
			l := migrate.NewConfigLoader(
				tt.fields.cfg,
				tt.fields.mr,
				tt.fields.pp,
			)
			if got := l.CFConfig(tt.args.toSource); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CFConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigLoader_CCDBConfig(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanCfg      config.OpsManager
		mr             *configfakes.FakeMigrationReader
		opsmanFactory  *fakes.FakeClientFactory
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		pp             *configfakes.FakePropertiesProvider
	}
	type args struct {
		toSource     bool
		migratorType string
	}
	tests := []struct {
		name       string
		opsman     *fakes.FakeClient
		fields     fields
		args       args
		want       interface{}
		beforeFunc func(fields, *fakes.FakeClient)
	}{
		{
			name:   "source ccdb config is loaded correctly when config struct is set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
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
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
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
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource:     true,
				migratorType: "sqlserver",
			},
			want: &cc.Config{
				SourceCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "192.168.12.24",
					Username:       "tas1_ccdb_username",
					Password:       "tas1_ccdb_password",
					EncryptionKey:  "tas1_ccdb_enc_key",
					SSHHost:        "opsman1.example.com",
					SSHUsername:    "ubuntu",
					SSHPassword:    "",
					SSHPrivateKey:  "/tmp/om2_rsa_key",
					TunnelRequired: true,
				},
			},
		},
		{
			name:   "source cf config is loaded correctly from opsman when config struct is not set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.BoshEnvironmentReturns("10.1.0.1", "some-client-id", "some-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
				client.DeployedProductReturns("cf-some-guid", nil)
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "tas1_ccdb_enc_key",
							"identity": "db_encryption",
						},
					},
				}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
					Migrators: []config.Migrator{
						{
							Name:  "sqlserver",
							Value: map[string]interface{}{},
						},
					},
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						UseDefaultMigrator: false,
						Migrators: []config.Migrator{
							{
								Name:  "sqlserver",
								Value: map[string]interface{}{},
							},
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
				opsmanCfg: config.OpsManager{
					Hostname:   "opsman.example.com",
					PrivateKey: "/path/to/ssh-key",
					SshUser:    "some-user",
				},
				pp: &configfakes.FakePropertiesProvider{
					SourceCCDBPropertiesBuilderStub: func(builder config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
						return &configfakes.FakeCCDBPropertiesBuilder{
							BuildStub: func() *config.CCDBProperties {
								return &config.CCDBProperties{
									Host:          "10.1.0.1",
									Username:      "tas1_ccdb_username",
									Password:      "tas1_ccdb_password",
									EncryptionKey: "tas1_ccdb_enc_key",
									SSHHost:       "opsman.url.com",
									SSHUsername:   "some-om-ssh-user",
									SSHPrivateKey: "/path/to/ssh_key",
								}
							},
						}
					},
				},
			},
			args: args{
				toSource:     true,
				migratorType: "sqlserver",
			},
			want: &cc.Config{
				SourceCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "10.1.0.1",
					Username:       "tas1_ccdb_username",
					Password:       "tas1_ccdb_password",
					EncryptionKey:  "tas1_ccdb_enc_key",
					SSHHost:        "opsman.url.com",
					SSHUsername:    "some-om-ssh-user",
					SSHPassword:    "",
					SSHPrivateKey:  "/path/to/ssh_key",
					TunnelRequired: true,
				},
			},
		},
		{
			name:   "target ccdb config is loaded correctly when config struct is set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				fields.mr.GetMigrationReturns(&config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"target_ccdb": map[string]interface{}{
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
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						Migrators: []config.Migrator{
							{
								Name: "sqlserver",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{
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
				},
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource:     false,
				migratorType: "sqlserver",
			},
			want: &cc.Config{
				TargetCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "192.168.12.24",
					Username:       "tas2_ccdb_username",
					Password:       "tas2_ccdb_password",
					EncryptionKey:  "tas2_ccdb_enc_key",
					SSHHost:        "opsman2.example.com",
					SSHUsername:    "ubuntu",
					SSHPassword:    "",
					SSHPrivateKey:  "/tmp/om2_rsa_key",
					TunnelRequired: true,
				},
			},
		},
		{
			name:   "target cf config is loaded correctly from opsman when config struct is not set",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.BoshEnvironmentReturns("10.1.0.1", "some-client-id", "some-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
				client.DeployedProductReturns("cf-some-guid", nil)
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "tas2_ccdb_enc_key",
							"identity": "db_encryption",
						},
					},
				}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
					Migrators: []config.Migrator{
						{
							Name:  "sqlserver",
							Value: map[string]interface{}{},
						},
					},
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
					Migration: config.Migration{
						UseDefaultMigrator: false,
						Migrators: []config.Migrator{
							{
								Name: "sqlserver",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "ecs",
								Value: map[string]interface{}{
									"target_ccdb": map[string]interface{}{},
								},
							},
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas2_ccdb_username",
							"password": "tas2_ccdb_password",
						},
					}},
				}, nil),
				opsmanCfg: config.OpsManager{
					Hostname:   "opsman.example.com",
					PrivateKey: "/path/to/ssh-key",
					SshUser:    "some-user",
				},
				pp: &configfakes.FakePropertiesProvider{
					TargetCCDBPropertiesBuilderStub: func(builder config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
						return &configfakes.FakeCCDBPropertiesBuilder{
							BuildStub: func() *config.CCDBProperties {
								return &config.CCDBProperties{
									Host:          "10.1.0.1",
									Username:      "tas2_ccdb_username",
									Password:      "tas2_ccdb_password",
									EncryptionKey: "tas2_ccdb_enc_key",
									SSHHost:       "opsman.url.com",
									SSHUsername:   "some-om-ssh-user",
									SSHPrivateKey: "/path/to/ssh_key",
								}
							},
						}
					},
				},
			},
			args: args{
				toSource:     false,
				migratorType: "sqlserver",
			},
			want: &cc.Config{
				TargetCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "10.1.0.1",
					Username:       "tas2_ccdb_username",
					Password:       "tas2_ccdb_password",
					EncryptionKey:  "tas2_ccdb_enc_key",
					SSHHost:        "opsman.url.com",
					SSHUsername:    "some-om-ssh-user",
					SSHPassword:    "",
					SSHPrivateKey:  "/path/to/ssh_key",
					TunnelRequired: true,
				},
			},
		},
		{
			name:   "migrator does not contain any source ccdb config",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
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
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
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
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource:     true,
				migratorType: "mysql",
			},
			want: nil,
		},
		{
			name:   "migrator does not contain any target ccdb config",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"target_ccdb": map[string]interface{}{
									"db_host":           "192.168.12.24",
									"db_username":       "tas2_ccdb_username",
									"db_password":       "tas2_ccdb_password",
									"db_encryption_key": "tas2_ccdb_enc_key",
									"ssh_host":          "opsman1.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om2_rsa_key",
									"ssh_tunnel":        true,
								},
							},
						},
					},
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
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
							{
								Name: "mysql",
								Value: map[string]interface{}{
									"backup_type":           "scp",
									"username":              "mysql",
									"hostname":              "mysql.backups.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/.ssh/key",
								},
							},
						},
					},
				},
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource:     false,
				migratorType: "mysql",
			},
			want: nil,
		},
		{
			name:   "source migrator does not contain any config",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
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
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
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
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource:     true,
				migratorType: "mysql",
			},
			want: nil,
		},
		{
			name:   "target migrator does not contain any config",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				fields.opsmanFactory.NewReturns(client, nil)
				fields.mr.GetMigrationReturns(&config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"target_ccdb": map[string]interface{}{
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
				}, nil)
			},
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "https://opsman.url.com",
							Username:     "some-om-username",
							Password:     "some-om-password",
							ClientID:     "some-om-client-id",
							ClientSecret: "some-om-client-secret",
							Hostname:     "opsman.url.com",
							IP:           "10.0.0.1",
							PrivateKey:   "/path/to/ssh_key",
							SshUser:      "some-om-ssh-user",
						},
					},
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
				mr:            new(configfakes.FakeMigrationReader),
				opsmanFactory: new(fakes.FakeClientFactory),
			},
			args: args{
				toSource:     false,
				migratorType: "mysql",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc(tt.fields, tt.opsman)
			l := migrate.NewConfigLoader(
				tt.fields.cfg,
				tt.fields.mr,
				tt.fields.pp,
			)
			got := l.CCDBConfig(tt.args.migratorType, tt.args.toSource)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("TestConfigLoader_CCDBConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigLoader_MigrationConfig(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	type fields struct {
		cfg            *config.Config
		opsmanFactory  *fakes.FakeClientFactory
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		pp             *configfakes.FakePropertiesProvider
	}
	type args struct {
		toSource     bool
		migratorType string
	}
	tests := []struct {
		name       string
		opsman     *fakes.FakeClient
		fields     fields
		args       args
		getReader  func(cfg *config.Config) *config.YAMLMigrationReader
		want       *cc.Config
		beforeFunc func(fields, *fakes.FakeClient)
	}{
		{
			name:   "creates a source cc config from an empty configured migrator",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.BoshEnvironmentReturns("10.1.0.1", "some-client-id", "some-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
				client.DeployedProductReturns("cf-some-guid", nil)
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "tas1_ccdb_enc_key",
							"identity": "db_encryption",
						},
					},
				}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
			},
			fields: fields{
				cfg:           config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-empty-migrator.yml")),
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.1.0.1"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
				pp: &configfakes.FakePropertiesProvider{
					SourceCCDBPropertiesBuilderStub: func(builder config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
						return &configfakes.FakeCCDBPropertiesBuilder{
							BuildStub: func() *config.CCDBProperties {
								return &config.CCDBProperties{
									Host:          "10.1.0.1",
									Username:      "tas1_ccdb_username",
									Password:      "tas1_ccdb_password",
									EncryptionKey: "tas1_ccdb_enc_key",
									SSHHost:       "opsman.source.example.com",
									SSHUsername:   "some-om1-ssh-user",
									SSHPrivateKey: "/path/to/om1_ssh_key",
								}
							},
						}
					},
				},
			},
			args: args{
				toSource:     true,
				migratorType: "sqlserver",
			},
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &cc.Config{
				SourceCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "10.1.0.1",
					Username:       "tas1_ccdb_username",
					Password:       "tas1_ccdb_password",
					EncryptionKey:  "tas1_ccdb_enc_key",
					SSHHost:        "opsman.source.example.com",
					SSHUsername:    "some-om1-ssh-user",
					SSHPrivateKey:  "/path/to/om1_ssh_key",
					TunnelRequired: true,
				},
			},
		},
		{
			name:   "creates a target cc config from an empty configured migrator",
			opsman: new(fakes.FakeClient),
			beforeFunc: func(fields fields, client *fakes.FakeClient) {
				client.BoshEnvironmentReturns("10.2.0.1", "some-client-id", "some-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
				client.DeployedProductReturns("cf-some-guid", nil)
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "tas2_ccdb_enc_key",
							"identity": "db_encryption",
						},
					},
				}, nil)
				fields.opsmanFactory.NewReturns(client, nil)
			},
			fields: fields{
				cfg:           config.New(filepath.Join(pwd, "testdata"), filepath.Join(pwd, "testdata", "si-migrator-empty-migrator.yml")),
				opsmanFactory: new(fakes.FakeClientFactory),
				boshFactory:   FakeBoshClientFactory([]string{"10.2.0.1"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas2_ccdb_username",
							"password": "tas2_ccdb_password",
						},
					}},
				}, nil),
				pp: &configfakes.FakePropertiesProvider{
					TargetCCDBPropertiesBuilderStub: func(builder config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
						return &configfakes.FakeCCDBPropertiesBuilder{
							BuildStub: func() *config.CCDBProperties {
								return &config.CCDBProperties{
									Host:          "10.2.0.1",
									Username:      "tas2_ccdb_username",
									Password:      "tas2_ccdb_password",
									EncryptionKey: "tas2_ccdb_enc_key",
									SSHHost:       "opsman.target.example.com",
									SSHUsername:   "some-om2-ssh-user",
									SSHPrivateKey: "/path/to/om2_ssh_key",
								}
							},
						}
					},
				},
			},
			args: args{
				toSource:     false,
				migratorType: "sqlserver",
			},
			getReader: func(cfg *config.Config) *config.YAMLMigrationReader {
				r, err := config.NewMigrationReader(cfg)
				require.NoError(t, err)
				return r
			},
			want: &cc.Config{
				TargetCloudControllerDatabase: cc.DatabaseConfig{
					Host:           "10.2.0.1",
					Username:       "tas2_ccdb_username",
					Password:       "tas2_ccdb_password",
					EncryptionKey:  "tas2_ccdb_enc_key",
					SSHHost:        "opsman.target.example.com",
					SSHUsername:    "some-om2-ssh-user",
					SSHPrivateKey:  "/path/to/om2_ssh_key",
					TunnelRequired: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc(tt.fields, tt.opsman)
			mr := tt.getReader(tt.fields.cfg)
			l := migrate.NewConfigLoader(
				tt.fields.cfg,
				mr,
				tt.fields.pp,
			)
			got := l.CCDBConfig(tt.args.migratorType, tt.args.toSource)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("TestConfigLoader_MigrationConfig() mismatch (-want +got):\n%s", diff)
			}
		})
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

func FakeCredhubClientFactory(creds map[string][]map[string]interface{}, err error) credhub.ClientFactoryFunc {
	return func(url string, credhubPort string, uaaPort string, allProxy string, caCert []byte, clientID string, clientSecret string) credhub.Client {
		return &credhubfakes.FakeClient{
			GetCredsStub: func(s string) (map[string][]map[string]interface{}, error) {
				return creds, err
			},
		}
	}
}
