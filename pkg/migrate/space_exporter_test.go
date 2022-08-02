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
	"context"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	cffakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"reflect"
	"testing"
)

func TestNewSpaceExporter(t *testing.T) {
	type args struct {
		e      *migrate.DefaultServiceInstanceExporter
		holder migrate.ClientHolder
	}
	tests := []struct {
		name string
		args args
		want *migrate.SpaceExporter
	}{
		{
			name: "create a new space exporter",
			args: args{
				e:      &migrate.DefaultServiceInstanceExporter{},
				holder: new(fakes.FakeClientHolder),
			},
			want: &migrate.SpaceExporter{
				ServiceInstanceExporter: &migrate.DefaultServiceInstanceExporter{},
				ClientHolder:            new(fakes.FakeClientHolder),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrate.NewSpaceExporter(tt.args.e, tt.args.holder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpaceExporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceExporter_Export(t *testing.T) {
	type fields struct {
		holder                  *fakes.FakeClientHolder
		ServiceInstanceExporter *fakes.FakeServiceInstanceExporter
	}
	type args struct {
		dir   string
		org   string
		space string
		om    config.OpsManager
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		cfClient   *cffakes.FakeClient
		beforeFunc func(*fakes.FakeServiceInstanceExporter)
		afterFunc  func(*fakes.FakeServiceInstanceExporter)
	}{
		{
			name: "exports a space",
			cfClient: &cffakes.FakeClient{
				GetOrgByNameStub: func(string) (cfclient.Org, error) {
					return cfclient.Org{
						Name: "some-org",
					}, nil
				},
				GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
					return cfclient.Space{
						Name: "some-space",
					}, nil
				},
			},
			fields: fields{
				ServiceInstanceExporter: new(fakes.FakeServiceInstanceExporter),
				holder:                  new(fakes.FakeClientHolder),
			},
			args: args{
				dir:   "/path/to/export-dir",
				org:   "some-org",
				space: "some-space",
				om:    config.OpsManager{Hostname: "opsman.url.com"},
			},
			wantErr: false,
			beforeFunc: func(stubServiceInstanceExporter *fakes.FakeServiceInstanceExporter) {
				stubServiceInstanceExporter.ExportManagedServicesCalls(func(ctx context.Context, org cfclient.Org, space cfclient.Space, om config.OpsManager, dir string) error {
					require.Equal(t, "opsman.url.com", om.Hostname)
					require.Equal(t, "some-org", org.Name)
					require.Equal(t, "some-space", space.Name)
					require.Equal(t, "/path/to/export-dir", dir)

					return nil
				})
			},
			afterFunc: func(stubServiceInstanceExporter *fakes.FakeServiceInstanceExporter) {
				require.Equal(t, 1, stubServiceInstanceExporter.ExportManagedServicesCallCount())
				require.Equal(t, 1, stubServiceInstanceExporter.ExportUserProvidedServicesCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.holder.SourceCFClientStub = func() cf.Client {
				return tt.cfClient
			}
			e := migrate.SpaceExporter{
				ClientHolder:            tt.fields.holder,
				ServiceInstanceExporter: tt.fields.ServiceInstanceExporter,
			}
			tt.beforeFunc(tt.fields.ServiceInstanceExporter)
			if err := e.Export(context.TODO(), tt.args.om, tt.args.dir, tt.args.org, tt.args.space); (err != nil) != tt.wantErr {
				t.Errorf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.afterFunc(tt.fields.ServiceInstanceExporter)
		})
	}
}
