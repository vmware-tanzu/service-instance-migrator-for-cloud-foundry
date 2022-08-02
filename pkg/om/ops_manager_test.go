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
	. "github.com/onsi/gomega"
	"reflect"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
)

func Test_GetBOSHCredentials(t *testing.T) {
	type fields struct {
		client *fakes.FakeOpsManClient
	}
	tests := []struct {
		name       string
		fields     fields
		want       om.BoshCredentials
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns a valid bosh environment with user credentials",
			fields: fields{
				client: &fakes.FakeOpsManClient{
					GetBOSHCredentialsStub: func() (string, error) {
						return "BOSH_CLIENT=ops_manager BOSH_CLIENT_SECRET=some-secret BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate BOSH_ENVIRONMENT=192.168.1.21 bosh", nil
					},
				},
			},
			want: om.BoshCredentials{
				Client:       "ops_manager",
				ClientSecret: "some-secret",
				Environment:  "192.168.1.21",
			},
			wantErr: false,
		},
		{
			name: "errors when fails to get bosh command line credentials",
			fields: fields{
				client: &fakes.FakeOpsManClient{
					GetBOSHCredentialsStub: func() (string, error) {
						return "", errors.New("error getting creds from opsman")
					},
				},
			},
			want: om.BoshCredentials{
				Client:       "ops_manager",
				ClientSecret: "some-secret",
				Environment:  "192.168.1.21",
			},
			wantErr:    true,
			wantErrMsg: "error getting creds from opsman",
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := om.NewOpsManager(
				tt.fields.client,
			)
			got, err := o.GetBOSHCredentials()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBOSHCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBOSHCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetCertificateAuthorities(t *testing.T) {
	type fields struct {
		client *fakes.FakeOpsManClient
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
				client: &fakes.FakeOpsManClient{
					ListCertificateAuthoritiesStub: func() ([]httpclient.CA, error) {
						return []httpclient.CA{
							{
								GUID:    "some-guid",
								Issuer:  "some-issuer",
								Active:  true,
								CertPEM: "a totally trustworthy cert",
							},
						}, nil
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
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := om.NewOpsManager(
				tt.fields.client,
			)
			got, err := o.GetCertificateAuthorities()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCertificateAuthorities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCertificateAuthorities() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetDeployedProductCredentials(t *testing.T) {
	type fields struct {
		client *fakes.FakeOpsManClient
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
			name: "gets deployed products",
			fields: fields{
				client: &fakes.FakeOpsManClient{
					ListDeployedProductCredentialsStub: func(s string, s2 string) (httpclient.DeployedProductCredential, error) {
						return httpclient.DeployedProductCredential{
							Credential: httpclient.Credential{
								Type:  "",
								Value: nil,
							},
						}, nil
					},
				},
			},
			args: args{},
			want: httpclient.DeployedProductCredential{
				Credential: httpclient.Credential{
					Type:  "",
					Value: nil,
				},
			},
			wantErr: false,
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := om.NewOpsManager(
				tt.fields.client,
			)
			got, err := o.GetDeployedProductCredentials(tt.args.deployedProductGUID, tt.args.credentialRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeployedProductCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeployedProductCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetStagedProductProperties(t *testing.T) {
	type fields struct {
		client *fakes.FakeOpsManClient
	}
	type args struct {
		product string
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
			name: "gets staged product properties",
			fields: fields{
				client: &fakes.FakeOpsManClient{
					GetStagedProductPropertiesStub: func(s string) (map[string]httpclient.ResponseProperty, error) {
						return map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
							Value:        "sys.cf.example.com",
							Configurable: true,
							IsCredential: false,
						}}, nil
					},
				},
			},
			args: args{},
			want: map[string]httpclient.ResponseProperty{".cloud_controller.system_domain": {
				Value:        "sys.cf.example.com",
				Configurable: true,
				IsCredential: false,
			}},
			wantErr: false,
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := om.NewOpsManager(
				tt.fields.client,
			)
			got, err := o.GetStagedProductProperties(tt.args.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStagedProductProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStagedProductProperties() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ListDeployedProducts(t *testing.T) {
	type fields struct {
		client *fakes.FakeOpsManClient
	}
	tests := []struct {
		name       string
		fields     fields
		want       []httpclient.DeployedProduct
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns a list of deployed products",
			fields: fields{
				client: &fakes.FakeOpsManClient{
					ListDeployedProductsStub: func() ([]httpclient.DeployedProduct, error) {
						return []httpclient.DeployedProduct{
							{
								Type: "cf",
								GUID: "cf-some-guid",
							},
						}, nil
					},
				},
			},
			want: []httpclient.DeployedProduct{
				{
					Type: "cf",
					GUID: "cf-some-guid",
				}},
			wantErr: false,
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := om.NewOpsManager(
				tt.fields.client,
			)
			got, err := o.ListDeployedProducts()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDeployedProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListDeployedProducts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
