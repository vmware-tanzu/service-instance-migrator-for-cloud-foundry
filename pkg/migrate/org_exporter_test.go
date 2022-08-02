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
	"bytes"
	"context"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	cffakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"net/url"
	"reflect"
	"testing"
)

func TestNewOrgExporter(t *testing.T) {
	type args struct {
		holder       *fakes.FakeClientHolder
		excludedOrgs []string
		includedOrgs []string
		e            *migrate.DefaultServiceInstanceExporter
	}
	tests := []struct {
		name string
		args args
		want *migrate.OrgExporter
	}{
		{
			name: "creates a new OrgExporter",
			args: args{
				holder:       &fakes.FakeClientHolder{},
				includedOrgs: []string{},
				excludedOrgs: []string{},
				e: &migrate.DefaultServiceInstanceExporter{
					ClientHolder: &fakes.FakeClientHolder{},
				},
			},
			want: &migrate.OrgExporter{
				ClientHolder: &fakes.FakeClientHolder{},
				IncludedOrgs: []string{},
				ExcludedOrgs: []string{},
				ServiceInstanceExporter: &migrate.DefaultServiceInstanceExporter{
					ClientHolder: &fakes.FakeClientHolder{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrate.NewOrgExporter(tt.args.e, tt.args.holder, tt.args.includedOrgs, tt.args.excludedOrgs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOrgExporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrgExporter_Export(t *testing.T) {
	stubServiceInstanceExporter := new(fakes.FakeServiceInstanceExporter)
	type args struct {
		dir  string
		orgs string
		om   config.OpsManager
	}
	type fields struct {
		holder   *fakes.FakeClientHolder
		exporter migrate.ServiceInstanceExporter
	}
	tests := []struct {
		name       string
		cfClient   *cffakes.FakeClient
		fields     fields
		args       args
		wantErr    bool
		want       string
		beforeFunc func()
		afterFunc  func()
	}{
		{
			name: "exports an org",
			cfClient: &cffakes.FakeClient{
				GetOrgByNameStub: func(string) (cfclient.Org, error) {
					return cfclient.Org{
						Name: "some-org",
					}, nil
				},
				ListSpacesByQueryStub: func(url.Values) ([]cfclient.Space, error) {
					return []cfclient.Space{
						{
							Name: "some-space",
						},
					}, nil
				},
			},
			args: args{
				dir:  "/path/to/export-dir",
				orgs: "some-org",
				om: config.OpsManager{
					Hostname: "opsman.source.url.com",
				},
			},
			fields: fields{
				holder:   new(fakes.FakeClientHolder),
				exporter: stubServiceInstanceExporter,
			},
			want: "",
			beforeFunc: func() {
				stubServiceInstanceExporter.ExportManagedServicesCalls(func(ctx context.Context, org cfclient.Org, space cfclient.Space, om config.OpsManager, dir string) error {
					require.Equal(t, "opsman.source.url.com", om.Hostname)
					require.Equal(t, "some-org", org.Name)
					require.Equal(t, "some-space", space.Name)
					require.Equal(t, "/path/to/export-dir", dir)
					return nil
				})
				stubServiceInstanceExporter.ExportUserProvidedServicesCalls(func(ctx context.Context, org cfclient.Org, space cfclient.Space, dir string) error {
					require.Equal(t, "some-org", org.Name)
					require.Equal(t, "some-space", space.Name)
					require.Equal(t, "/path/to/export-dir", dir)
					return nil
				})
			},
			afterFunc: func() {
				require.Equal(t, 1, stubServiceInstanceExporter.ExportManagedServicesCallCount())
				require.Equal(t, 1, stubServiceInstanceExporter.ExportUserProvidedServicesCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			tt.fields.holder.SourceCFClientStub = func() cf.Client {
				return tt.cfClient
			}
			e := migrate.OrgExporter{
				ClientHolder:            tt.fields.holder,
				ServiceInstanceExporter: tt.fields.exporter,
			}
			tt.beforeFunc()
			if err := e.Export(context.TODO(), tt.args.om, tt.args.dir, tt.args.orgs); (err != nil) != tt.wantErr {
				t.Errorf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, writer.String())
			tt.afterFunc()
		})
	}
}

func TestOrgExporter_ExportAll(t *testing.T) {
	stubServiceInstanceExporter := new(fakes.FakeServiceInstanceExporter)
	type args struct {
		dir string
		om  config.OpsManager
	}
	type fields struct {
		holder   *fakes.FakeClientHolder
		exporter migrate.ServiceInstanceExporter
	}
	tests := []struct {
		name       string
		cfClient   cf.Client
		fields     fields
		args       args
		wantErr    bool
		want       string
		beforeFunc func()
		afterFunc  func()
	}{
		{
			name: "exports an org",
			cfClient: &cffakes.FakeClient{
				ListSpacesStub: func() ([]cfclient.Space, error) {
					return []cfclient.Space{
						{
							Name: "space-1",
						},
						{
							Name: "space-2",
						},
					}, nil
				},
				GetOrgByGuidStub: func(string) (cfclient.Org, error) {
					return cfclient.Org{
						Name: "some-org",
					}, nil
				},
			},
			args: args{
				dir: "/path/to/export-dir",
				om: config.OpsManager{
					Hostname: "opsman.source.url.com",
				},
			},
			fields: fields{
				holder:   new(fakes.FakeClientHolder),
				exporter: stubServiceInstanceExporter,
			},
			want: "",
			beforeFunc: func() {
				stubServiceInstanceExporter.ExportManagedServicesCalls(func(ctx context.Context, org cfclient.Org, space cfclient.Space, om config.OpsManager, dir string) error {
					require.Equal(t, "opsman.source.url.com", om.Hostname)
					require.Equal(t, "some-org", org.Name)
					require.Equal(t, "/path/to/export-dir", dir)
					return nil
				})
				stubServiceInstanceExporter.ExportUserProvidedServicesCalls(func(ctx context.Context, org cfclient.Org, space cfclient.Space, dir string) error {
					require.Equal(t, "some-org", org.Name)
					require.Equal(t, "/path/to/export-dir", dir)
					return nil
				})
			},
			afterFunc: func() {
				require.Equal(t, 2, stubServiceInstanceExporter.ExportManagedServicesCallCount())
				require.Equal(t, 2, stubServiceInstanceExporter.ExportUserProvidedServicesCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			tt.fields.holder.SourceCFClientStub = func() cf.Client {
				return tt.cfClient
			}
			e := migrate.OrgExporter{
				ClientHolder:            tt.fields.holder,
				ServiceInstanceExporter: tt.fields.exporter,
			}
			tt.beforeFunc()
			if err := e.ExportAll(context.TODO(), tt.args.om, tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, writer.String())
			tt.afterFunc()
		})
	}
}
