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

package bosh_test

import (
	"errors"
	"testing"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	clifakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/cli/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/fakes"
	. "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_VerifyAuth(t *testing.T) {
	var (
		fakeDirector, fakeDirectorUnauthenticated *clifakes.FakeDirector
	)
	type fields struct {
		url             string
		allProxy        string
		trustedCertPEM  []byte
		boshAuth        Authentication
		certAppender    *fakes.FakeCertAppender
		directorFactory *fakes.FakeDirectorFactory
		uaaFactory      *fakes.FakeUAAFactory
	}
	var tests = []struct {
		name            string
		fields          fields
		wantErr         bool
		wantNewErr      bool
		wantErrMsg      string
		beforeNewFunc   func(fields fields)
		afterNewFunc    func(fields fields, client bosh.Client)
		afterVerifyFunc func(Authentication, *fakes.FakeUAAFactory)
	}{
		{
			name: "uaa is configured",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			beforeNewFunc: func(fields fields) {
				fakeDirectorFactory := fields.directorFactory
				fakeCertAppender := fields.certAppender
				fakeCertAppender.AppendCertsFromPEMReturns(true)
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, nil)
			},
			afterNewFunc: func(fields fields, client bosh.Client) {
				allProxy, directorConfig, taskReporter, fileReporter := fields.directorFactory.NewArgsForCall(0)
				assert.Equal(t, "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key", allProxy)
				assert.Equal(t, boshdir.FactoryConfig{
					Host:   "192.168.1.21",
					Port:   25555,
					CACert: "a totally trustworthy cert",
				}, directorConfig)
				assert.Nil(t, directorConfig.TokenFunc)
				assert.Equal(t, boshdir.NoopTaskReporter{}, taskReporter)
				assert.Equal(t, boshdir.NoopFileReporter{}, fileReporter)
				assert.Equal(t, 1, fakeDirectorUnauthenticated.InfoCallCount())
				assert.Equal(t, 1, fields.certAppender.AppendCertsFromPEMCallCount())
				assert.Equal(t, []byte("a totally trustworthy cert"), fields.certAppender.AppendCertsFromPEMArgsForCall(0))
				assert.Equal(t, time.Duration(5), client.(*bosh.ClientImpl).PollingInterval)
			},
			afterVerifyFunc: func(boshAuthConfig Authentication, fakeUAAFactory *fakes.FakeUAAFactory) {
				assert.Equal(t, 1, fakeUAAFactory.NewCallCount())
				uaaConfig := fakeUAAFactory.NewArgsForCall(0)
				assert.Equal(t, uaa.Config{
					AllProxy:     "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
					Host:         "uaa.url.example.com",
					Port:         12345,
					CACert:       "a totally trustworthy cert",
					ClientID:     boshAuthConfig.UAA.ClientCredentials.ID,
					ClientSecret: boshAuthConfig.UAA.ClientCredentials.Secret,
				}, uaaConfig)
			},
		},
		{
			name: "uaa is not configured (i.e. basic auth)",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					Basic: UserCredentials{Username: "example-username", Password: "example-password"},
				},
			},
			beforeNewFunc: func(fields fields) {
				fakeDirectorFactory := fields.directorFactory
				fakeCertAppender := fields.certAppender
				fakeCertAppender.AppendCertsFromPEMReturns(true)
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "basic",
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "basic",
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, nil)
			},
			afterNewFunc: func(fields fields, client bosh.Client) {
				allProxy, directorConfig, taskReporter, fileReporter := fields.directorFactory.NewArgsForCall(0)
				assert.Equal(t, "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key", allProxy)
				assert.Equal(t, boshdir.FactoryConfig{
					Host:   "192.168.1.21",
					Port:   25555,
					CACert: "a totally trustworthy cert",
				}, directorConfig)
				assert.Nil(t, directorConfig.TokenFunc)
				assert.Equal(t, boshdir.NoopTaskReporter{}, taskReporter)
				assert.Equal(t, boshdir.NoopFileReporter{}, fileReporter)
				assert.Equal(t, 1, fakeDirectorUnauthenticated.InfoCallCount())
				assert.Equal(t, 1, fields.certAppender.AppendCertsFromPEMCallCount())
				assert.Equal(t, []byte("a totally trustworthy cert"), fields.certAppender.AppendCertsFromPEMArgsForCall(0))
				assert.Equal(t, time.Duration(5), client.(*bosh.ClientImpl).PollingInterval)
			},
			afterVerifyFunc: func(boshAuthConfig Authentication, fakeUAAFactory *fakes.FakeUAAFactory) {
				assert.Equal(t, 0, fakeUAAFactory.NewCallCount())
			},
		},
		{
			name: "new returns error when url is bad",
			fields: fields{
				url:             "https://not a valid url",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					Basic: UserCredentials{Username: "example-username", Password: "example-password"},
				},
			},
			wantNewErr:    true,
			wantErrMsg:    "failed to build director config from url",
			beforeNewFunc: func(fields fields) {},
			afterVerifyFunc: func(boshAuthConfig Authentication, fakeUAAFactory *fakes.FakeUAAFactory) {
				assert.Equal(t, 0, fakeUAAFactory.NewCallCount())
			},
		},
		{
			name: "new returns error when director factory errors",
			fields: fields{
				url:             "https://example.org:25555",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					Basic: UserCredentials{Username: "example-username", Password: "example-password"},
				},
			},
			wantNewErr: true,
			wantErrMsg: "failed to build director: could not build director",
			beforeNewFunc: func(fields fields) {
				fields.directorFactory.NewReturnsOnCall(0, new(clifakes.FakeDirector), errors.New("could not build director"))
			},
		},
		{
			name: "new returns error when the director fails to GetInfo",
			fields: fields{
				url:             "https://example.org:25555",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					Basic: UserCredentials{Username: "example-username", Password: "example-password"},
				},
			},
			wantNewErr: true,
			wantErrMsg: "error fetching BOSH director information: could not get info",
			beforeNewFunc: func(fields fields) {
				fakeDirectorFactory := fields.directorFactory
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{}, errors.New("could not get info"))
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
			},
		},
		{
			name: "errors when uaa url is not valid",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "failed to build UAA config from url",
			beforeNewFunc: func(fields fields) {
				fakeDirectorFactory := fields.directorFactory
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "http://what is this",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
			},
			afterNewFunc: func(fields fields, client bosh.Client) {},
		},
		{
			name: "errors when uaa factory returns an error",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "failed to build UAA client: failed to build uaa",
			beforeNewFunc: func(fields fields) {
				fakeDirectorFactory := fields.directorFactory
				fakeUAAFactory := fields.uaaFactory
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, nil)
				fakeUAAFactory.NewReturns(new(fakes.FakeUAA), errors.New("failed to build uaa"))
			},
			afterNewFunc: func(fields fields, client bosh.Client) {},
		},
		{
			name: "errors when uaa is not deployed",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "failed to build UAA config from url: expected non-empty URL",
			beforeNewFunc: func(fields fields) {
				fakeDirectorFactory := fields.directorFactory
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "basic",
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
			},
			afterNewFunc: func(fields fields, client bosh.Client) {},
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeNewFunc(tt.fields)
			client, err := bosh.New(
				tt.fields.url,
				tt.fields.allProxy,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.directorFactory,
				tt.fields.uaaFactory,
				tt.fields.boshAuth,
			)
			if (err != nil) != tt.wantNewErr {
				t.Fatalf("New() error = %v, wantNewErr %v", err, tt.wantNewErr)
			}
			if tt.wantNewErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, client)
			tt.afterNewFunc(tt.fields, client)

			err = client.VerifyAuth()
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			require.NoError(t, err)
			tt.afterVerifyFunc(tt.fields.boshAuth, tt.fields.uaaFactory)
		})
	}
}

func TestClient_FindDeployment(t *testing.T) {
	var (
		fakeDirector, fakeDirectorUnauthenticated *clifakes.FakeDirector
	)
	type args struct {
		pattern string
	}
	type fields struct {
		url             string
		allProxy        string
		trustedCertPEM  []byte
		boshAuth        Authentication
		certAppender    *fakes.FakeCertAppender
		directorFactory *fakes.FakeDirectorFactory
		uaaFactory      *fakes.FakeUAAFactory
	}
	var tests = []struct {
		name          string
		args          args
		fields        fields
		wantErr       bool
		wantNewErr    bool
		wantErrMsg    string
		wantResp      boshdir.DeploymentResp
		wantFound     bool
		beforeNewFunc func(fields fields, args args)
		afterFunc     func(fields fields, args args)
	}{
		{
			name: "deployment is found",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			args: args{pattern: "^cf-"},
			wantResp: boshdir.DeploymentResp{
				Name: "cf-some-guid",
			},
			wantFound: true,
			beforeNewFunc: func(fields fields, args args) {
				fakeDirectorFactory := fields.directorFactory
				fakeCertAppender := fields.certAppender
				fakeCertAppender.AppendCertsFromPEMReturns(true)
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, nil)
				fakeDirector.ListDeploymentsReturns([]boshdir.DeploymentResp{
					{
						Name: "cf-some-guid",
					},
				}, nil)
			},
			afterFunc: func(fields fields, args args) {
				assert.Equal(t, 1, fakeDirector.ListDeploymentsCallCount())
			},
		},
		{
			name: "deployment is not found",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			args:      args{pattern: "abc"},
			wantResp:  boshdir.DeploymentResp{},
			wantFound: false,
			beforeNewFunc: func(fields fields, args args) {
				fakeDirectorFactory := fields.directorFactory
				fakeCertAppender := fields.certAppender
				fakeCertAppender.AppendCertsFromPEMReturns(true)
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, nil)
				fakeDirector.ListDeploymentsReturns([]boshdir.DeploymentResp{
					{
						Name: "cf-some-guid",
					},
				}, nil)
			},
			afterFunc: func(fields fields, args args) {
				assert.Equal(t, 1, fakeDirector.ListDeploymentsCallCount())
			},
		},
		{
			name: "errors when director is mis-configured",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			args:       args{pattern: "^cf-"},
			wantResp:   boshdir.DeploymentResp{},
			wantFound:  false,
			wantErr:    true,
			wantErrMsg: "failed to build director: could not build director",
			beforeNewFunc: func(fields fields, args args) {
				fakeDirectorFactory := fields.directorFactory
				fakeCertAppender := fields.certAppender
				fakeCertAppender.AppendCertsFromPEMReturns(true)
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, errors.New("could not build director"))
			},
			afterFunc: func(fields fields, args args) {
				assert.Equal(t, 0, fakeDirector.ListDeploymentsCallCount())
			},
		},
		{
			name: "errors when ListDeployments returns an error",
			fields: fields{
				url:             "192.168.1.21",
				allProxy:        "ssh+socks5://ubuntu@jumpbox.example.org:22?private-key=/.ssh/private_key",
				trustedCertPEM:  []byte("a totally trustworthy cert"),
				certAppender:    new(fakes.FakeCertAppender),
				directorFactory: new(fakes.FakeDirectorFactory),
				uaaFactory:      new(fakes.FakeUAAFactory),
				boshAuth: Authentication{
					UAA: UAAAuthentication{
						ClientCredentials: ClientCredentials{
							ID:     "bosh-user",
							Secret: "bosh-secret",
						},
					},
				},
			},
			args:       args{pattern: "^cf-"},
			wantResp:   boshdir.DeploymentResp{},
			wantFound:  false,
			wantErr:    true,
			wantErrMsg: "cannot get the list of deployments: could not list deployments",
			beforeNewFunc: func(fields fields, args args) {
				fakeDirectorFactory := fields.directorFactory
				fakeCertAppender := fields.certAppender
				fakeCertAppender.AppendCertsFromPEMReturns(true)
				fakeDirectorUnauthenticated = new(clifakes.FakeDirector)
				fakeDirectorUnauthenticated.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(0, fakeDirectorUnauthenticated, nil)
				fakeDirector = new(clifakes.FakeDirector)
				fakeDirector.IsAuthenticatedReturns(true, nil)
				fakeDirector.InfoReturns(boshdir.Info{
					Version: "1.3262.0.0 (00000000)",
					User:    "bosh-username",
					Auth: boshdir.UserAuthentication{
						Type: "uaa",
						Options: map[string]interface{}{
							"url": "uaa.url.example.com:12345",
						},
					},
				}, nil)
				fakeDirectorFactory.NewReturnsOnCall(1, fakeDirector, nil)
				fakeDirector.ListDeploymentsReturns([]boshdir.DeploymentResp{}, errors.New("could not list deployments"))
			},
			afterFunc: func(fields fields, args args) {
				assert.Equal(t, 1, fakeDirector.ListDeploymentsCallCount())
			},
		},
	}
	g := NewGomegaWithT(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeNewFunc(tt.fields, tt.args)
			client, err := bosh.New(
				tt.fields.url,
				tt.fields.allProxy,
				tt.fields.trustedCertPEM,
				tt.fields.certAppender,
				tt.fields.directorFactory,
				tt.fields.uaaFactory,
				tt.fields.boshAuth,
			)
			if (err != nil) != tt.wantNewErr {
				t.Fatalf("New() error = %v, wantNewErr %v", err, tt.wantNewErr)
			}
			if tt.wantNewErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, client)

			resp, found, err := client.FindDeployment(tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErrMsg)))
				tt.afterFunc(tt.fields, tt.args)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantFound, found)
			require.Equal(t, tt.wantResp, resp)

			tt.afterFunc(tt.fields, tt.args)
		})
	}
}
