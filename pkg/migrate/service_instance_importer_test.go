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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
)

func TestManagedServiceInstanceImporter_ImportManagedService(t *testing.T) {
	type fields struct {
		Registry *fakes.FakeMigratorRegistry
	}
	type args struct {
		ctx       context.Context
		org       string
		space     string
		instance  *cf.ServiceInstance
		importDir string
		om        config.OpsManager
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		afterFunc func(t *testing.T, fields fields)
	}{
		{
			name: "imports a managed service",
			fields: fields{
				Registry: &fakes.FakeMigratorRegistry{
					LookupStub: func(org string, space string, instance *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (migrate.ServiceInstanceMigrator, bool, error) {
						return &fakes.FakeServiceInstanceMigrator{
							MigrateStub: func(ctx context.Context) (*cf.ServiceInstance, error) {
								return &cf.ServiceInstance{}, nil
							},
						}, true, nil
					},
				},
			},
			args: args{
				ctx:   context.TODO(),
				org:   "some-org",
				space: "some-space",
				instance: &cf.ServiceInstance{
					Name:    "mysqldb",
					GUID:    "some-guid",
					Type:    "managed_service_instance",
					Service: "p.mysql",
				},
				importDir: "/path/to/import-dir",
				om: config.OpsManager{
					Hostname: "opsman.url.com",
				},
			},
			wantErr: false,
			afterFunc: func(t *testing.T, fields fields) {
				require.Equal(t, 1, fields.Registry.LookupCallCount())
				_, _, si, om, _, _ := fields.Registry.LookupArgsForCall(0)
				require.Equal(t, "opsman.url.com", om.Hostname)
				require.Equal(t, &cf.ServiceInstance{
					Name:    "mysqldb",
					GUID:    "some-guid",
					Type:    "managed_service_instance",
					Service: "p.mysql",
				}, si)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := migrate.ManagedServiceInstanceImporter{
				Registry: tt.fields.Registry,
			}
			if err := i.ImportManagedService(tt.args.ctx, tt.args.org, tt.args.space, tt.args.instance, tt.args.om, tt.args.importDir); (err != nil) != tt.wantErr {
				t.Errorf("ImportManagedService() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.afterFunc(t, tt.fields)
		})
	}
}
