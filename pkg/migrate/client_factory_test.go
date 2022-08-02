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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"
	"reflect"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	boshfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	omfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/fakes"
)

func TestNewClientFactory(t *testing.T) {
	type args struct {
		configLoader  config.Loader
		boshFactory   bosh.ClientFactory
		opsmanFactory om.ClientFactory
		opsmanConfig  config.OpsManager
	}
	tests := []struct {
		name string
		args args
		want *migrate.ClientFactory
	}{
		{
			name: "creates a client factory",
			args: args{
				configLoader:  new(fakes.FakeLoader),
				boshFactory:   new(boshfakes.FakeClientFactory),
				opsmanFactory: new(omfakes.FakeClientFactory),
				opsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				},
			},
			want: &migrate.ClientFactory{
				ConfigLoader:  new(fakes.FakeLoader),
				BoshFactory:   new(boshfakes.FakeClientFactory),
				OpsmanFactory: new(omfakes.FakeClientFactory),
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrate.NewClientFactory(tt.args.configLoader, tt.args.boshFactory, tt.args.opsmanFactory, tt.args.opsmanConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClientFactory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientFactory_CFClient(t *testing.T) {
	type fields struct {
		ConfigLoader  *fakes.FakeLoader
		BoshFactory   *boshfakes.FakeClientFactory
		OpsmanFactory *omfakes.FakeClientFactory
		OpsmanConfig  config.OpsManager
	}
	type args struct {
		toSource bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   cf.Client
	}{
		{
			name: "creates a source cf client",
			fields: fields{
				ConfigLoader: &fakes.FakeLoader{
					SourceApiConfigStub: func() *config.CloudController {
						return &config.CloudController{
							URL:          "https://api.cf.url.com",
							Username:     "some-cf-user",
							Password:     "some-cf-password",
							ClientID:     "some-cf-client",
							ClientSecret: "some-cf-client-secret",
						}
					},
				},
				BoshFactory:   new(boshfakes.FakeClientFactory),
				OpsmanFactory: new(omfakes.FakeClientFactory),
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				},
			},
			args: args{
				toSource: true,
			},
			want: &cf.ClientImpl{
				Config: &cf.Config{
					URL:          "https://api.cf.url.com",
					Username:     "some-cf-user",
					Password:     "some-cf-password",
					ClientID:     "some-cf-client",
					ClientSecret: "some-cf-client-secret",
					SSLDisabled:  true,
				},
				RetryTimeout: cf.DefaultRetryTimeout,
				RetryPause:   cf.DefaultRetryPause,
			},
		},
		{
			name: "creates a target cf client",
			fields: fields{
				ConfigLoader: &fakes.FakeLoader{
					TargetApiConfigStub: func() *config.CloudController {
						return &config.CloudController{
							URL:          "https://api.cf.url.com",
							Username:     "some-cf-user",
							Password:     "some-cf-password",
							ClientID:     "some-cf-client",
							ClientSecret: "some-cf-client-secret",
						}
					},
				},
				BoshFactory:   new(boshfakes.FakeClientFactory),
				OpsmanFactory: new(omfakes.FakeClientFactory),
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				},
			},
			args: args{
				toSource: false,
			},
			want: &cf.ClientImpl{
				Config: &cf.Config{
					URL:          "https://api.cf.url.com",
					Username:     "some-cf-user",
					Password:     "some-cf-password",
					ClientID:     "some-cf-client",
					ClientSecret: "some-cf-client-secret",
					SSLDisabled:  true,
				},
				RetryTimeout: cf.DefaultRetryTimeout,
				RetryPause:   cf.DefaultRetryPause,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := migrate.ClientFactory{
				ConfigLoader:  tt.fields.ConfigLoader,
				BoshFactory:   tt.fields.BoshFactory,
				OpsmanFactory: tt.fields.OpsmanFactory,
				OpsmanConfig:  tt.fields.OpsmanConfig,
			}
			got := l.CFClient(tt.args.toSource)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(cf.ClientImpl{})); diff != "" {
				t.Errorf("CFClient() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientFactory_SourceOpsManClient(t *testing.T) {
	type fields struct {
		ConfigLoader  *fakes.FakeLoader
		BoshFactory   *boshfakes.FakeClientFactory
		OpsmanFactory *omfakes.FakeClientFactory
		OpsmanConfig  config.OpsManager
	}
	tests := []struct {
		name   string
		fields fields
		want   om.Client
	}{
		{
			name: "creates a source opsman client",
			fields: fields{
				ConfigLoader: new(fakes.FakeLoader),
				BoshFactory:  new(boshfakes.FakeClientFactory),
				OpsmanFactory: &omfakes.FakeClientFactory{
					NewStub: func(s string, bytes []byte, appender om.CertAppender, factory om.OpsManagerFactory, factory2 om.UAAFactory, authentication config.Authentication) (om.Client, error) {
						return om.New(
							"https://opsman.url.com",
							nil,
							nil,
							om.NewFactory(),
							uaa.NewFactory(),
							config.Authentication{
								UAA: config.UAAAuthentication{
									URL: "https://opsman.url.com/uaa",
									UserCredentials: config.UserCredentials{
										Username: "admin",
										Password: "admin-password",
									},
								},
							},
						)
					},
				},
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				},
			},
			want: &om.ClientImpl{
				URL: "https://opsman.url.com",
				Auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://opsman.url.com/uaa",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "admin-password",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := migrate.ClientFactory{
				ConfigLoader:  tt.fields.ConfigLoader,
				BoshFactory:   tt.fields.BoshFactory,
				OpsmanFactory: tt.fields.OpsmanFactory,
				OpsmanConfig:  tt.fields.OpsmanConfig,
			}
			got := l.SourceOpsManClient()
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(om.ClientImpl{})); diff != "" {
				t.Errorf("SourceOpsManClient() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientFactory_TargetOpsManClient(t *testing.T) {
	type fields struct {
		ConfigLoader  *fakes.FakeLoader
		BoshFactory   *boshfakes.FakeClientFactory
		OpsmanFactory *omfakes.FakeClientFactory
		OpsmanConfig  config.OpsManager
	}
	tests := []struct {
		name   string
		fields fields
		want   om.Client
	}{
		{
			name: "creates a target opsman client",
			fields: fields{
				ConfigLoader: new(fakes.FakeLoader),
				BoshFactory:  new(boshfakes.FakeClientFactory),
				OpsmanFactory: &omfakes.FakeClientFactory{
					NewStub: func(s string, bytes []byte, appender om.CertAppender, factory om.OpsManagerFactory, factory2 om.UAAFactory, authentication config.Authentication) (om.Client, error) {
						return om.New(
							"https://opsman.url.com",
							nil,
							nil,
							om.NewFactory(),
							uaa.NewFactory(),
							config.Authentication{
								UAA: config.UAAAuthentication{
									URL: "https://opsman.url.com/uaa",
									UserCredentials: config.UserCredentials{
										Username: "admin",
										Password: "admin-password",
									},
								},
							},
						)
					},
				},
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				},
			},
			want: &om.ClientImpl{
				URL: "https://opsman.url.com",
				Auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://opsman.url.com/uaa",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "admin-password",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := migrate.ClientFactory{
				ConfigLoader:  tt.fields.ConfigLoader,
				BoshFactory:   tt.fields.BoshFactory,
				OpsmanFactory: tt.fields.OpsmanFactory,
				OpsmanConfig:  tt.fields.OpsmanConfig,
			}
			got := l.TargetOpsManClient()
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(om.ClientImpl{})); diff != "" {
				t.Errorf("TargetOpsManClient() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientFactory_SourceBoshClient(t *testing.T) {
	type fields struct {
		ConfigLoader  *fakes.FakeLoader
		BoshFactory   *boshfakes.FakeClientFactory
		OpsmanFactory *omfakes.FakeClientFactory
		OpsmanConfig  config.OpsManager
	}
	tests := []struct {
		name   string
		fields fields
		want   bosh.Client
	}{
		{
			name: "creates a source bosh client",
			fields: fields{
				ConfigLoader: &fakes.FakeLoader{
					SourceBoshConfigStub: func() *config.Bosh {
						return &config.Bosh{
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
						}
					},
				},
				BoshFactory: &boshfakes.FakeClientFactory{
					NewStub: func(s string, s2 string, bytes []byte, appender bosh.CertAppender, factory bosh.DirectorFactory, factory2 bosh.UAAFactory, authentication config.Authentication) (bosh.Client, error) {
						return &bosh.ClientImpl{
							PollingInterval: 0,
							BoshInfo: bosh.Info{
								Version: "0.0.0",
								UserAuthentication: bosh.UserAuthentication{
									Options: bosh.AuthenticationOptions{
										URL: "https://10.1.0.1:8443",
									},
								},
							},
						}, nil
					},
				},
				OpsmanFactory: new(omfakes.FakeClientFactory),
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				}},
			want: &bosh.ClientImpl{
				PollingInterval: 0,
				BoshInfo: bosh.Info{
					Version: "0.0.0",
					UserAuthentication: bosh.UserAuthentication{
						Options: bosh.AuthenticationOptions{
							URL: "https://10.1.0.1:8443",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := migrate.ClientFactory{
				ConfigLoader:  tt.fields.ConfigLoader,
				BoshFactory:   tt.fields.BoshFactory,
				OpsmanFactory: tt.fields.OpsmanFactory,
				OpsmanConfig:  tt.fields.OpsmanConfig,
			}
			got := l.SourceBoshClient()
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(bosh.ClientImpl{})); diff != "" {
				t.Errorf("SourceBoshClient() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientFactory_TargetBoshClient(t *testing.T) {
	type fields struct {
		ConfigLoader  *fakes.FakeLoader
		BoshFactory   *boshfakes.FakeClientFactory
		OpsmanFactory *omfakes.FakeClientFactory
		OpsmanConfig  config.OpsManager
	}
	tests := []struct {
		name   string
		fields fields
		want   bosh.Client
	}{
		{
			name: "creates a target bosh client",
			fields: fields{
				ConfigLoader: &fakes.FakeLoader{
					TargetBoshConfigStub: func() *config.Bosh {
						return &config.Bosh{
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
						}
					},
				},
				BoshFactory: &boshfakes.FakeClientFactory{
					NewStub: func(s string, s2 string, bytes []byte, appender bosh.CertAppender, factory bosh.DirectorFactory, factory2 bosh.UAAFactory, authentication config.Authentication) (bosh.Client, error) {
						return &bosh.ClientImpl{
							PollingInterval: 0,
							BoshInfo: bosh.Info{
								Version: "0.0.0",
								UserAuthentication: bosh.UserAuthentication{
									Options: bosh.AuthenticationOptions{
										URL: "https://10.1.0.1:8443",
									},
								},
							},
						}, nil
					},
				},
				OpsmanFactory: new(omfakes.FakeClientFactory),
				OpsmanConfig: config.OpsManager{
					URL:          "https://opsman.url.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "client",
					ClientSecret: "client-secret",
					Hostname:     "opsman.url.com",
					IP:           "192.168.1.12",
					PrivateKey:   "/path/to/om/private-key",
					SshUser:      "om-ssh-user",
				}},
			want: &bosh.ClientImpl{
				PollingInterval: 0,
				BoshInfo: bosh.Info{
					Version: "0.0.0",
					UserAuthentication: bosh.UserAuthentication{
						Options: bosh.AuthenticationOptions{
							URL: "https://10.1.0.1:8443",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := migrate.ClientFactory{
				ConfigLoader:  tt.fields.ConfigLoader,
				BoshFactory:   tt.fields.BoshFactory,
				OpsmanFactory: tt.fields.OpsmanFactory,
				OpsmanConfig:  tt.fields.OpsmanConfig,
			}
			got := l.TargetBoshClient()
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(bosh.ClientImpl{})); diff != "" {
				t.Errorf("TargetBoshClient() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
