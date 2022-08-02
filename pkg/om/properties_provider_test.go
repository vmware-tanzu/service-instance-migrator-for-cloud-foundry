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
	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	boshfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	configfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	credhubfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/fakes"
	"testing"
)

func TestPropertiesProvider_SourceBoshPropertiesBuilder(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
	}
	tests := []struct {
		name     string
		fields   fields
		wantFunc func(client om.Client, opsManager config.OpsManager) config.BoshPropertiesBuilder
	}{
		{
			name: "creates a bosh properties builder for a source foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory: FakeBoshClientFactory([]string{"192.168.12.24"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
			},
			wantFunc: func(client om.Client, opsManager config.OpsManager) config.BoshPropertiesBuilder {
				return om.NewBOSHPropertiesBuilder(client, opsManager)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.SourceBoshPropertiesBuilder()
			want := tt.wantFunc(client, tt.fields.opsmanConfig)
			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{})); diff != "" {
				t.Errorf("SourceBoshPropertiesBuilder() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPropertiesProvider_SourceCCDBPropertiesBuilder(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		boshProperties *config.BoshProperties
	}
	type args struct {
		boshPropertiesBuilderFunc func(*config.BoshProperties) *configfakes.FakeBoshPropertiesBuilder
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantFunc func(client om.Client, opsman config.OpsManager, boshFactory bosh.ClientFactory, credhubFactory credhub.ClientFactory, boshProperties *config.BoshProperties) config.CCDBPropertiesBuilder
	}{
		{
			name: "creates a ccdb properties builder for a source foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory:    &boshfakes.FakeClientFactory{},
				credhubFactory: &credhubfakes.FakeClientFactory{},
				boshProperties: &config.BoshProperties{
					URL: "some-url",
				},
			},
			args: args{
				boshPropertiesBuilderFunc: FakeBoshPropertiesBuilderFunc(),
			},
			wantFunc: func(client om.Client, opsman config.OpsManager, boshFactory bosh.ClientFactory, credhubFactory credhub.ClientFactory, boshProperties *config.BoshProperties) config.CCDBPropertiesBuilder {
				return om.NewCCDBPropertiesBuilder(client, opsman, boshFactory, credhubFactory, boshProperties)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.SourceCCDBPropertiesBuilder(tt.args.boshPropertiesBuilderFunc(tt.fields.boshProperties))
			want := tt.wantFunc(client, tt.fields.opsmanConfig, tt.fields.boshFactory, tt.fields.credhubFactory, tt.fields.boshProperties)
			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{}, boshfakes.FakeClientFactory{}, credhubfakes.FakeClientFactory{})); diff != "" {
				t.Errorf("SourceCCDBPropertiesBuilder() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPropertiesProvider_SourceCFPropertiesBuilder(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		deployment     string
	}
	tests := []struct {
		name     string
		fields   fields
		wantFunc func(client om.Client, cfDeployment string) config.CFPropertiesBuilder
	}{
		{
			name: "creates a bosh properties builder for a source foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory: FakeBoshClientFactory([]string{"192.168.12.24"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
			},
			wantFunc: func(client om.Client, cfDeployment string) config.CFPropertiesBuilder {
				return om.NewCFPropertiesBuilder(client, cfDeployment)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.SourceCFPropertiesBuilder()
			want := tt.wantFunc(client, tt.fields.deployment)
			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{})); diff != "" {
				t.Errorf("SourceCFPropertiesBuilder() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPropertiesProvider_Environment(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
	}
	type args struct {
		boshPropertiesBuilder config.BoshPropertiesBuilder
		cfPropertiesBuilder   config.CFPropertiesBuilder
		ccdbPropertiesBuilder config.CCDBPropertiesBuilder
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   config.EnvProperties
	}{
		{
			name: "creates a environment properties builder for a source foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Source: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory:    &boshfakes.FakeClientFactory{},
				credhubFactory: &credhubfakes.FakeClientFactory{},
			},
			args: args{
				boshPropertiesBuilder: &configfakes.FakeBoshPropertiesBuilder{},
				cfPropertiesBuilder:   &configfakes.FakeCFPropertiesBuilder{},
				ccdbPropertiesBuilder: &configfakes.FakeCCDBPropertiesBuilder{},
			},
			want: config.EnvProperties{},
		},
		{
			name: "creates a environment properties builder for a target foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory:    &boshfakes.FakeClientFactory{},
				credhubFactory: &credhubfakes.FakeClientFactory{},
			},
			args: args{
				boshPropertiesBuilder: &configfakes.FakeBoshPropertiesBuilder{},
				cfPropertiesBuilder:   &configfakes.FakeCFPropertiesBuilder{},
				ccdbPropertiesBuilder: &configfakes.FakeCCDBPropertiesBuilder{},
			},
			want: config.EnvProperties{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.Environment(tt.args.boshPropertiesBuilder, tt.args.cfPropertiesBuilder, tt.args.ccdbPropertiesBuilder)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{})); diff != "" {
				t.Errorf("Environment() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPropertiesProvider_TargetBoshPropertiesBuilder(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
	}
	tests := []struct {
		name     string
		fields   fields
		wantFunc func(client om.Client, opsManager config.OpsManager) config.BoshPropertiesBuilder
	}{
		{
			name: "creates a bosh properties builder for a target foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory: FakeBoshClientFactory([]string{"192.168.12.24"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
			},
			wantFunc: func(client om.Client, opsManager config.OpsManager) config.BoshPropertiesBuilder {
				return om.NewBOSHPropertiesBuilder(client, opsManager)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.TargetBoshPropertiesBuilder()
			want := tt.wantFunc(client, tt.fields.opsmanConfig)
			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{})); diff != "" {
				t.Errorf("SourceBoshPropertiesBuilder() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPropertiesProvider_TargetCCDBPropertiesBuilder(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		boshProperties *config.BoshProperties
	}
	type args struct {
		boshPropertiesBuilderFunc func(*config.BoshProperties) *configfakes.FakeBoshPropertiesBuilder
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantFunc func(client om.Client, opsman config.OpsManager, boshFactory bosh.ClientFactory, credhubFactory credhub.ClientFactory, boshProperties *config.BoshProperties) config.CCDBPropertiesBuilder
	}{
		{
			name: "creates a ccdb properties builder for a target foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory:    &boshfakes.FakeClientFactory{},
				credhubFactory: &credhubfakes.FakeClientFactory{},
				boshProperties: &config.BoshProperties{
					URL: "some-url",
				},
			},
			args: args{
				boshPropertiesBuilderFunc: FakeBoshPropertiesBuilderFunc(),
			},
			wantFunc: func(client om.Client, opsman config.OpsManager, boshFactory bosh.ClientFactory, credhubFactory credhub.ClientFactory, boshProperties *config.BoshProperties) config.CCDBPropertiesBuilder {
				return om.NewCCDBPropertiesBuilder(client, opsman, boshFactory, credhubFactory, boshProperties)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.TargetCCDBPropertiesBuilder(tt.args.boshPropertiesBuilderFunc(tt.fields.boshProperties))
			want := tt.wantFunc(client, tt.fields.opsmanConfig, tt.fields.boshFactory, tt.fields.credhubFactory, tt.fields.boshProperties)
			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{}, boshfakes.FakeClientFactory{}, credhubfakes.FakeClientFactory{})); diff != "" {
				t.Errorf("SourceCCDBPropertiesBuilder() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPropertiesProvider_TargetCFPropertiesBuilder(t *testing.T) {
	type fields struct {
		cfg            *config.Config
		opsmanConfig   config.OpsManager
		boshFactory    bosh.ClientFactory
		credhubFactory credhub.ClientFactory
		deployment     string
	}
	tests := []struct {
		name     string
		fields   fields
		wantFunc func(client om.Client, cfDeployment string) config.CFPropertiesBuilder
	}{
		{
			name: "creates a bosh properties builder for a target foundation",
			fields: fields{
				cfg: &config.Config{
					Foundations: struct {
						Source config.OpsManager `yaml:"source"`
						Target config.OpsManager `yaml:"target"`
					}{
						Target: config.OpsManager{
							URL:          "some-url",
							Username:     "some-username",
							Password:     "some-password",
							ClientID:     "some-client-id",
							ClientSecret: "some-client-secret",
							Hostname:     "some-hostname",
							IP:           "some-ip",
							PrivateKey:   "some-private-key",
							SshUser:      "some-ssh-user",
						},
					},
				},
				opsmanConfig: config.OpsManager{
					URL:          "some-url",
					Username:     "some-username",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-secret",
					Hostname:     "some-hostname",
					IP:           "some-ip",
					PrivateKey:   "some-private-key",
					SshUser:      "some-ssh-user",
				},
				boshFactory: FakeBoshClientFactory([]string{"192.168.12.24"}),
				credhubFactory: FakeCredhubClientFactory(map[string][]map[string]interface{}{
					"data": {map[string]interface{}{
						"value": map[string]interface{}{
							"username": "tas1_ccdb_username",
							"password": "tas1_ccdb_password",
						},
					}},
				}, nil),
			},
			wantFunc: func(client om.Client, cfDeployment string) config.CFPropertiesBuilder {
				return om.NewCFPropertiesBuilder(client, cfDeployment)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakes.FakeClient{}
			l := om.NewPropertiesProvider(
				tt.fields.cfg,
				tt.fields.opsmanConfig,
				FakeOpsManagerClientFactory(client),
				tt.fields.boshFactory,
				tt.fields.credhubFactory,
			)
			got := l.TargetCFPropertiesBuilder()
			want := tt.wantFunc(client, tt.fields.deployment)
			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(fakes.FakeClient{})); diff != "" {
				t.Errorf("TargetCFPropertiesBuilder() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func FakeBoshPropertiesBuilderFunc() func(p *config.BoshProperties) *configfakes.FakeBoshPropertiesBuilder {
	return func(p *config.BoshProperties) *configfakes.FakeBoshPropertiesBuilder {
		return &configfakes.FakeBoshPropertiesBuilder{
			BuildStub: func() *config.BoshProperties {
				return p
			},
		}
	}
}

func FakeOpsManagerClientFactory(client om.Client) om.ClientFactoryFunc {
	return func(url string, trustedCertPEM []byte, certAppender om.CertAppender, opsManagerFactory om.OpsManagerFactory, uaaFactory om.UAAFactory, auth config.Authentication) (om.Client, error) {
		return client, nil
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
