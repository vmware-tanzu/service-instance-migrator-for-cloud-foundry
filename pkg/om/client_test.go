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
	"errors"
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"
	uaafakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa/fakes"
)

func TestClient_BoshEnvironment(t *testing.T) {
	type fields struct {
		url               string
		trustedCertPEM    []byte
		certAppender      *fakes.FakeCertAppender
		auth              config.Authentication
		uaaFactory        *fakes.FakeUAAFactory
		opsManagerFactory *fakes.FakeOpsManagerFactory
	}
	tests := []struct {
		name       string
		fields     fields
		want       string
		want1      string
		want2      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns a valid bosh environment with user credentials",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetBOSHCredentialsStub: func() (om.BoshCredentials, error) {
								return om.BoshCredentials{
									Client:       "ops_manager",
									ClientSecret: "some-secret",
									Environment:  "192.168.1.21",
								}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			want:    "192.168.1.21",
			want1:   "ops_manager",
			want2:   "some-secret",
			wantErr: false,
		},
		{
			name: "returns a valid bosh environment with client credentials",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						ClientCredentials: config.ClientCredentials{
							ID:     "ops_manager_client",
							Secret: "ops_manager_secret",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetBOSHCredentialsStub: func() (om.BoshCredentials, error) {
								return om.BoshCredentials{
									Client:       "ops_manager",
									ClientSecret: "some-secret",
									Environment:  "192.168.1.21",
								}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			want:    "192.168.1.21",
			want1:   "ops_manager",
			want2:   "some-secret",
			wantErr: false,
		},
		{
			name: "errors when url is bad",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://not a valid url",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build UAA client: failed to build UAA config from url",
		},
		{
			name: "errors when uaa auth is not set",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{},
			},
			wantErr:    true,
			wantErrMsg: "uaa auth must be set for opsman api authentication",
		},
		{
			name: "errors when GetBOSHCredentials fails",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetBOSHCredentialsStub: func() (om.BoshCredentials, error) {
								return om.BoshCredentials{}, errors.New("error getting bosh creds")
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "cannot get bosh environment variables from ops manager: error getting bosh creds",
		},
		{
			name: "errors when opsman url is blank",
			fields: fields{
				url:            "",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetBOSHCredentialsStub: func() (om.BoshCredentials, error) {
								return om.BoshCredentials{}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build ops manager config from url",
		},
		{
			name: "errors when opsman url is bad",
			fields: fields{
				url:            "https://not a valid url",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetBOSHCredentialsStub: func() (om.BoshCredentials, error) {
								return om.BoshCredentials{}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build ops manager config from url",
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := om.New(
				tt.fields.url,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.opsManagerFactory,
				tt.fields.uaaFactory,
				tt.fields.auth,
			)
			require.NoError(t, err)
			got, got1, got2, err := c.BoshEnvironment()
			if (err != nil) != tt.wantErr {
				t.Errorf("BoshEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if got != tt.want {
				t.Errorf("BoshEnvironment() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("BoshEnvironment() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("BoshEnvironment() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestClient_CertificateAuthorities(t *testing.T) {
	type fields struct {
		url               string
		trustedCertPEM    []byte
		certAppender      om.CertAppender
		auth              config.Authentication
		uaaFactory        om.UAAFactory
		opsManagerFactory om.OpsManagerFactory
	}
	tests := []struct {
		name       string
		fields     fields
		want       []httpclient.CA
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns valid certificate authorities",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetCertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
								return []httpclient.CA{
									{
										GUID:    "some-guid",
										Issuer:  "some-issuer",
										Active:  true,
										CertPEM: "a totally trustworthy cert",
									},
								}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			want: []httpclient.CA{
				{
					GUID:    "some-guid",
					Issuer:  "some-issuer",
					Active:  true,
					CertPEM: "a totally trustworthy cert",
				},
			},
			wantErr: false,
		},
		{
			name: "errors when url is bad",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://not a valid url",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build UAA client: failed to build UAA config from url",
		},
		{
			name: "errors when GetCertificateAuthorities fails",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetCertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
								return []httpclient.CA{}, errors.New("error getting CAs")
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "cannot get certificate authorities from ops manager: error getting CAs",
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := om.New(
				tt.fields.url,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.opsManagerFactory,
				tt.fields.uaaFactory,
				tt.fields.auth,
			)
			require.NoError(t, err)
			got, err := c.CertificateAuthorities()
			if (err != nil) != tt.wantErr {
				t.Errorf("CertificateAuthorities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CertificateAuthorities() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeployedProduct(t *testing.T) {
	type fields struct {
		url               string
		trustedCertPEM    []byte
		certAppender      om.CertAppender
		auth              config.Authentication
		uaaFactory        om.UAAFactory
		opsManagerFactory om.OpsManagerFactory
	}
	type args struct {
		productType string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns a valid cf product",
			args: args{
				productType: "cf",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							ListDeployedProductsStub: func() ([]httpclient.DeployedProduct, error) {
								return []httpclient.DeployedProduct{
									{
										Type: "cf",
										GUID: "cf-some-guid",
									},
								}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			want:    "cf-some-guid",
			wantErr: false,
		},
		{
			name: "errors when url is bad",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://not a valid url",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build UAA client: failed to build UAA config from url",
		},
		{
			name: "errors when ListDeployedProducts fails",
			args: args{
				productType: "cf",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							ListDeployedProductsStub: func() ([]httpclient.DeployedProduct, error) {
								return []httpclient.DeployedProduct{}, errors.New("error getting deployed products")
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "cannot get a list of deployed products from ops manager, product type \"cf\": error getting deployed products",
		},
		{
			name: "errors when no products are found",
			args: args{
				productType: "cf",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							ListDeployedProductsStub: func() ([]httpclient.DeployedProduct, error) {
								return []httpclient.DeployedProduct{}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: fmt.Sprintf("failed to find product for type %q", "cf"),
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := om.New(
				tt.fields.url,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.opsManagerFactory,
				tt.fields.uaaFactory,
				tt.fields.auth,
			)
			require.NoError(t, err)
			got, err := c.DeployedProduct(tt.args.productType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployedProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if got != tt.want {
				t.Errorf("DeployedProduct() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeployedProductCredentials(t *testing.T) {
	type fields struct {
		url               string
		trustedCertPEM    []byte
		certAppender      om.CertAppender
		auth              config.Authentication
		uaaFactory        om.UAAFactory
		opsManagerFactory om.OpsManagerFactory
	}
	type args struct {
		deployedProductGUID string
		credentialRef       string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       httpclient.DeployedProductCredential
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns valid product credentials",
			args: args{
				deployedProductGUID: "cf-some-guid",
				credentialRef:       ".uaa.admin_credentials",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetDeployedProductCredentialsStub: func(s string, s2 string) (httpclient.DeployedProductCredential, error) {
								return httpclient.DeployedProductCredential{
									Credential: httpclient.Credential{
										Type: "simple_credentials",
										Value: map[string]string{
											"password": "some-password",
											"identity": "some-identity",
										},
									},
								}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			want: httpclient.DeployedProductCredential{
				Credential: httpclient.Credential{
					Type: "simple_credentials",
					Value: map[string]string{
						"password": "some-password",
						"identity": "some-identity",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors when url is bad",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://not a valid url",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build UAA client: failed to build UAA config from url",
		},
		{
			name: "errors when GetDeployedProductCredentials fails",
			args: args{
				deployedProductGUID: "cf-some-guid",
				credentialRef:       ".uaa.admin_credentials",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetDeployedProductCredentialsStub: func(s string, s2 string) (httpclient.DeployedProductCredential, error) {
								return httpclient.DeployedProductCredential{}, errors.New("error getting deployed product creds")
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "cannot get deployed product credentials from ops manager, product guid \"cf-some-guid\": error getting deployed product creds",
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := om.New(
				tt.fields.url,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.opsManagerFactory,
				tt.fields.uaaFactory,
				tt.fields.auth,
			)
			require.NoError(t, err)
			got, err := c.DeployedProductCredentials(tt.args.deployedProductGUID, tt.args.credentialRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployedProductCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeployedProductCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_StagedProductProperties(t *testing.T) {
	type fields struct {
		url               string
		trustedCertPEM    []byte
		certAppender      om.CertAppender
		auth              config.Authentication
		uaaFactory        om.UAAFactory
		opsManagerFactory om.OpsManagerFactory
	}
	type args struct {
		deployedProductGUID string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       map[string]httpclient.ResponseProperty
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns valid staged product config",
			args: args{
				deployedProductGUID: "cf-some-guid",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetStagedProductPropertiesStub: func(s string) (map[string]httpclient.ResponseProperty, error) {
								return map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
									Value:        "sys.cf.example.com",
									Configurable: true,
									IsCredential: false,
								}}, nil
							},
						}
						return opsman, nil
					},
				},
			},
			want: map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
				Value:        "sys.cf.example.com",
				Configurable: true,
				IsCredential: false,
			}},
			wantErr: false,
		},
		{
			name: "errors when url is bad",
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "https://not a valid url",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{},
			},
			wantErr:    true,
			wantErrMsg: "failed to create ops manager client: failed to build UAA client: failed to build UAA config from url",
		},
		{
			name: "errors when GetStagedProductProperties fails",
			args: args{
				deployedProductGUID: "cf-some-guid",
			},
			fields: fields{
				url:            "https://opsman.url.example.com",
				trustedCertPEM: []byte("a totally trustworthy cert"),
				certAppender: &fakes.FakeCertAppender{
					AppendCertsFromPEMStub: func(bytes []byte) bool {
						return true
					},
				},
				auth: config.Authentication{
					UAA: config.UAAAuthentication{
						URL: "uaa.url.example.com:12345",
						UserCredentials: config.UserCredentials{
							Username: "admin",
							Password: "some-password",
						},
					},
				},
				uaaFactory: &fakes.FakeUAAFactory{
					NewStub: func(c uaa.Config) (uaa.UAA, error) {
						return new(uaafakes.FakeUAA), nil
					},
				},
				opsManagerFactory: &fakes.FakeOpsManagerFactory{
					NewStub: func(o om.Config) (om.OpsManager, error) {
						opsman := &fakes.FakeOpsManager{
							GetStagedProductPropertiesStub: func(s string) (map[string]httpclient.ResponseProperty, error) {
								return nil, errors.New("error getting staged product config")
							},
						}
						return opsman, nil
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "cannot get staged product properties from ops manager, product guid \"cf-some-guid\": error getting staged product config",
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := om.New(
				tt.fields.url,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.opsManagerFactory,
				tt.fields.uaaFactory,
				tt.fields.auth,
			)
			require.NoError(t, err)
			got, err := c.StagedProductProperties(tt.args.deployedProductGUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("StagedProductProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StagedProductProperties() got = %v, want %v", got, tt.want)
			}
		})
	}
}
