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
	"testing"
)

func TestConfig_ParseValidate(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	type args struct {
		configFileName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    config.Config
	}{
		{
			name: "config object with uaa auth using client credentials",
			args: args{
				configFileName: "bosh_config_with_uaa_client_creds.yml",
			},
			want: config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Source: config.OpsManager{
						URL:          "https://opsman-1.src.tas.example.com",
						Username:     "admin-om1",
						Password:     "some-om1-password",
						ClientID:     "",
						ClientSecret: "",
						Hostname:     "opsman-1.src.tas.example.com",
						IP:           "",
						PrivateKey:   "/Users/user/.ssh/opsman1",
						SshUser:      "ubuntu-1",
					},
					Target: config.OpsManager{
						URL:          "https://opsman-2.dst.tas.example.com",
						Username:     "admin-om2",
						Password:     "some-om2-password",
						ClientID:     "",
						ClientSecret: "",
						Hostname:     "opsman-2.dst.tas.example.com",
						IP:           "",
						PrivateKey:   "/Users/user/.ssh/opsman2",
						SshUser:      "ubuntu-2",
					},
				},
				SourceApi: config.CloudController{
					URL:          "https://api.src.tas.example.com",
					Username:     "",
					Password:     "",
					ClientID:     "client-with-cloudcontroller-admin-permissions",
					ClientSecret: "client-secret",
				},
				SourceBosh: config.Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "some-url",
							ClientCredentials: config.ClientCredentials{
								ID:     "some-client-id",
								Secret: "some-client-secret",
							},
						},
					},
				},
				TargetApi: config.CloudController{
					URL:          "https://api.dst.tas.example.com",
					Username:     "",
					Password:     "",
					ClientID:     "client-with-cloudcontroller-admin-permissions",
					ClientSecret: "client-secret",
				},
				TargetBosh: config.Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "some-url",
							ClientCredentials: config.ClientCredentials{
								ID:     "some-client-id",
								Secret: "some-client-secret",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "config object with uaa auth using user credentials",
			args: args{
				configFileName: "bosh_config_with_uaa_user_creds.yml",
			},
			want: config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Source: config.OpsManager{
						URL:          "https://opsman-1.src.tas.example.com",
						Username:     "admin-om1",
						Password:     "some-om1-password",
						ClientID:     "",
						ClientSecret: "",
						Hostname:     "opsman-1.src.tas.example.com",
						IP:           "",
						PrivateKey:   "/Users/user/.ssh/opsman1",
						SshUser:      "ubuntu-1",
					},
					Target: config.OpsManager{
						URL:          "https://opsman-2.dst.tas.example.com",
						Username:     "admin-om2",
						Password:     "some-om2-password",
						ClientID:     "",
						ClientSecret: "",
						Hostname:     "opsman-2.dst.tas.example.com",
						IP:           "",
						PrivateKey:   "/Users/user/.ssh/opsman2",
						SshUser:      "ubuntu-2",
					},
				},
				SourceApi: config.CloudController{
					URL:          "https://api.src.tas.example.com",
					Username:     "",
					Password:     "",
					ClientID:     "client-with-cloudcontroller-admin-permissions",
					ClientSecret: "client-secret",
				},
				SourceBosh: config.Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "some-url",
							UserCredentials: config.UserCredentials{
								Username: "some-username",
								Password: "some-password",
							},
						},
					},
				},
				TargetApi: config.CloudController{
					URL:          "https://api.dst.tas.example.com",
					Username:     "",
					Password:     "",
					ClientID:     "client-with-cloudcontroller-admin-permissions",
					ClientSecret: "client-secret",
				},
				TargetBosh: config.Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: config.Authentication{
						UAA: config.UAAAuthentication{
							URL: "some-url",
							UserCredentials: config.UserCredentials{
								Username: "some-username",
								Password: "some-password",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "config object with basic auth",
			args: args{
				configFileName: "bosh_config.yml",
			},
			wantErr: false,
			want: config.Config{
				Foundations: struct {
					Source config.OpsManager `yaml:"source"`
					Target config.OpsManager `yaml:"target"`
				}{
					Source: config.OpsManager{
						URL:          "https://opsman-1.src.tas.example.com",
						Username:     "admin-om1",
						Password:     "some-om1-password",
						ClientID:     "",
						ClientSecret: "",
						Hostname:     "opsman-1.src.tas.example.com",
						IP:           "",
						PrivateKey:   "/Users/user/.ssh/opsman1",
						SshUser:      "ubuntu-1",
					},
					Target: config.OpsManager{
						URL:          "https://opsman-2.dst.tas.example.com",
						Username:     "admin-om2",
						Password:     "some-om2-password",
						ClientID:     "",
						ClientSecret: "",
						Hostname:     "opsman-2.dst.tas.example.com",
						IP:           "",
						PrivateKey:   "/Users/user/.ssh/opsman2",
						SshUser:      "ubuntu-2",
					},
				},
				SourceApi: config.CloudController{
					URL:          "https://api.src.tas.example.com",
					Username:     "",
					Password:     "",
					ClientID:     "client-with-cloudcontroller-admin-permissions",
					ClientSecret: "client-secret",
				},
				SourceBosh: config.Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: config.Authentication{
						Basic: config.UserCredentials{
							Username: "some-username",
							Password: "some-password",
						},
					},
				},
				TargetApi: config.CloudController{
					URL:          "https://api.dst.tas.example.com",
					Username:     "",
					Password:     "",
					ClientID:     "client-with-cloudcontroller-admin-permissions",
					ClientSecret: "client-secret",
				},
				TargetBosh: config.Bosh{
					URL:         "some-url",
					AllProxy:    "some-proxy-url",
					TrustedCert: "some-cert",
					Authentication: config.Authentication{
						Basic: config.UserCredentials{
							Username: "some-username",
							Password: "some-password",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFilePath := filepath.Join(cwd, "testdata", tt.args.configFileName)
			got, parseErr := config.Parse(configFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.NoError(t, parseErr)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
