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

package om_test

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/fakes"

	"github.com/google/go-cmp/cmp"
)

func TestEnvironment(t *testing.T) {
	type args struct {
		boshProperties *config.BoshProperties
		cfProperties   *config.CFProperties
		ccdbProperties *config.CCDBProperties
	}
	tests := []struct {
		name string
		args args
		want config.EnvProperties
	}{
		{
			name: "creates an env properties",
			args: args{
				boshProperties: &config.BoshProperties{
					URL:          "https://10.1.0.1:9999",
					AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					ClientID:     "some-bosh-client-id",
					ClientSecret: "some-bosh-client-secret",
					RootCA: []httpclient.CA{
						{
							Active:  true,
							CertPEM: "a very trustworthy cert",
						},
					},
					Deployment: "cf-some-guid",
				},
				cfProperties: &config.CFProperties{
					URL:      "https://api.cf.url.com",
					Username: "some-cf-user",
					Password: "some-cf-password",
				},
				ccdbProperties: &config.CCDBProperties{
					Host:          "10.1.22.3",
					Username:      "tas1_ccdb_username",
					Password:      "tas1_ccdb_password",
					EncryptionKey: "tas1_ccdb_enc_key",
					SSHHost:       "opsman.url.com",
					SSHUsername:   "some-om-ssh-user",
					SSHPrivateKey: "/path/to/ssh_key",
				},
			},
			want: config.EnvProperties{
				BoshProperties: &config.BoshProperties{
					URL:          "https://10.1.0.1:9999",
					AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					ClientID:     "some-bosh-client-id",
					ClientSecret: "some-bosh-client-secret",
					RootCA: []httpclient.CA{
						{
							Active:  true,
							CertPEM: "a very trustworthy cert",
						},
					},
					Deployment: "cf-some-guid",
				},
				CFProperties: &config.CFProperties{
					URL:      "https://api.cf.url.com",
					Username: "some-cf-user",
					Password: "some-cf-password",
				},
				CCDBProperties: &config.CCDBProperties{
					Host:          "10.1.22.3",
					Username:      "tas1_ccdb_username",
					Password:      "tas1_ccdb_password",
					EncryptionKey: "tas1_ccdb_enc_key",
					SSHHost:       "opsman.url.com",
					SSHUsername:   "some-om-ssh-user",
					SSHPrivateKey: "/path/to/ssh_key",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, om.Environment(tt.args.boshProperties, tt.args.cfProperties, tt.args.ccdbProperties)); diff != "" {
				t.Errorf("Environment() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_BOSHPropertiesBuilder_Build(t *testing.T) {
	type fields struct {
		client     *fakes.FakeClient
		opsManager config.OpsManager
	}
	tests := []struct {
		name       string
		fields     fields
		want       *config.BoshProperties
		beforeFunc func(*fakes.FakeClient)
	}{
		{
			name: "builds bosh config from opsman client calls",
			fields: fields{
				client: new(fakes.FakeClient),
				opsManager: config.OpsManager{
					URL:        "https://opsman.url.com",
					Hostname:   "opsman.url.com",
					PrivateKey: "/path/to/ssh_key",
					SshUser:    "some-om-ssh-user",
				},
			},
			beforeFunc: func(client *fakes.FakeClient) {
				client.DeployedProductReturns("cf-some-guid", nil)
				client.BoshEnvironmentReturns("https://10.1.0.1:9999", "some-bosh-client-id", "some-bosh-client-secret", nil)
				client.CertificateAuthoritiesReturns([]httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				}, nil)
			},
			want: &config.BoshProperties{
				URL:          "https://10.1.0.1:9999",
				AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
				ClientID:     "some-bosh-client-id",
				ClientSecret: "some-bosh-client-secret",
				RootCA: []httpclient.CA{
					{
						Active:  true,
						CertPEM: "a very trustworthy cert",
					},
				},
				Deployment: "cf-some-guid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc(tt.fields.client)
			b := om.NewBOSHPropertiesBuilder(
				tt.fields.client,
				tt.fields.opsManager,
			)
			if diff := cmp.Diff(tt.want, b.Build()); diff != "" {
				t.Errorf("BOSHPropertiesBuilder.Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_CCDBPropertiesBuilder_Build(t *testing.T) {
	type fields struct {
		client         *fakes.FakeClient
		boshProperties *config.BoshProperties
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		opsManager     config.OpsManager
	}
	tests := []struct {
		name       string
		fields     fields
		want       *config.CCDBProperties
		beforeFunc func(*fakes.FakeClient)
	}{
		{
			name: "builds ccdb config from opsman client calls",
			fields: fields{
				client: new(fakes.FakeClient),
				boshProperties: &config.BoshProperties{
					URL:          "https://bosh.url.com",
					AllProxy:     "ssh+socks5://some-om-ssh-user@opsman.url.com:22?private-key=/path/to/ssh_key",
					ClientID:     "some-bosh-client-id",
					ClientSecret: "some-bosh-client-secret",
					RootCA: []httpclient.CA{
						{
							Active:  true,
							CertPEM: "a very trustworthy cert",
						},
					},
					Deployment: "cf",
				},
				boshFactory: FakeBoshClientFactory([]string{"10.1.22.3"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
				opsManager: config.OpsManager{
					URL:        "https://opsman.url.com",
					Hostname:   "opsman.url.com",
					PrivateKey: "/path/to/ssh_key",
					SshUser:    "some-om-ssh-user",
				},
			},
			beforeFunc: func(client *fakes.FakeClient) {
				client.DeployedProductCredentialsReturns(httpclient.DeployedProductCredential{
					Credential: httpclient.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "tas1_ccdb_enc_key",
							"identity": "db_encryption",
						},
					},
				}, nil)
			},
			want: &config.CCDBProperties{
				Host:          "10.1.22.3",
				Username:      "tas1_ccdb_username",
				Password:      "tas1_ccdb_password",
				EncryptionKey: "tas1_ccdb_enc_key",
				SSHHost:       "opsman.url.com",
				SSHUsername:   "some-om-ssh-user",
				SSHPrivateKey: "/path/to/ssh_key",
			},
		},
	}
	for _, tt := range tests {
		tt.beforeFunc(tt.fields.client)
		t.Run(tt.name, func(t *testing.T) {
			b := om.NewCCDBPropertiesBuilder(
				tt.fields.client,
				tt.fields.opsManager,
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
				tt.fields.boshProperties,
			)
			if diff := cmp.Diff(tt.want, b.Build()); diff != "" {
				t.Errorf("CCDBPropertiesBuilder.Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_CFPropertiesBuilder_Build(t *testing.T) {
	type fields struct {
		client       *fakes.FakeClient
		cfDeployment string
	}
	tests := []struct {
		name       string
		fields     fields
		want       *config.CFProperties
		beforeFunc func(*fakes.FakeClient)
	}{
		{
			name: "builds cf config from opsman client calls",
			fields: fields{
				client:       new(fakes.FakeClient),
				cfDeployment: "",
			},
			beforeFunc: func(client *fakes.FakeClient) {
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
			},
			want: &config.CFProperties{
				URL:      "https://api.cf.url.com",
				Username: "some-cf-user",
				Password: "some-cf-password",
			},
		},
	}
	for _, tt := range tests {
		tt.beforeFunc(tt.fields.client)
		t.Run(tt.name, func(t *testing.T) {
			b := om.NewCFPropertiesBuilder(tt.fields.client, tt.fields.cfDeployment)
			if diff := cmp.Diff(tt.want, b.Build()); diff != "" {
				t.Errorf("CFPropertiesBuilder.Build() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
