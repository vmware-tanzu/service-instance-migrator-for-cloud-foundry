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
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	configfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"

	"github.com/stretchr/testify/require"
)

func TestDefaultMigratorRegistry(t *testing.T) {
	type fields struct {
		helper       *migrate.MigratorHelper
		cfg          *config.Config
		cfgLoader    *configfakes.FakeLoader
		factory      *fakes.FakeFactory
		clientHolder *fakes.FakeClientHolder
	}
	type args struct {
		serviceInstance *cf.ServiceInstance
		om              config.OpsManager
	}
	tests := []struct {
		name                   string
		fields                 fields
		args                   args
		want                   bool
		AssertFactoryCallCount func(*fakes.FakeFactory)
	}{
		{
			name: "should migrate when config contains service names",
			fields: fields{
				factory:   new(fakes.FakeFactory),
				helper:    &migrate.MigratorHelper{},
				cfg:       &config.Config{Services: []string{"SQLServer"}},
				cfgLoader: new(configfakes.FakeLoader),
			},
			args: args{
				serviceInstance: &cf.ServiceInstance{
					Service: "SQLServer",
				},
			},
			want: true,
			AssertFactoryCallCount: func(factory *fakes.FakeFactory) {
				require.Equal(t, 1, factory.NewCallCount())
			},
		},
		{
			name: "should migrate when config contains migrator names",
			fields: fields{
				factory:   new(fakes.FakeFactory),
				helper:    &migrate.MigratorHelper{},
				cfg:       &config.Config{Services: []string{"mysql", "sqlserver"}},
				cfgLoader: new(configfakes.FakeLoader),
			},
			args: args{
				serviceInstance: &cf.ServiceInstance{
					Service: "p.mysql",
				},
			},
			want: true,
			AssertFactoryCallCount: func(factory *fakes.FakeFactory) {
				require.Equal(t, 1, factory.NewCallCount())
			},
		},
		{
			name: "should not migrate when config does not contain migrator or service names",
			fields: fields{
				factory:   new(fakes.FakeFactory),
				helper:    &migrate.MigratorHelper{},
				cfg:       &config.Config{Services: []string{"PostgreSQL", "SQLServer"}},
				cfgLoader: new(configfakes.FakeLoader),
			},
			args: args{
				serviceInstance: &cf.ServiceInstance{
					Service: "p.mysql",
				},
			},
			want: false,
			AssertFactoryCallCount: func(factory *fakes.FakeFactory) {
				require.Equal(t, 0, factory.NewCallCount())
			},
		},
		{
			name: "should migrate when config service names match migrator name regardless of casing",
			fields: fields{
				factory:   new(fakes.FakeFactory),
				helper:    &migrate.MigratorHelper{},
				cfg:       &config.Config{Services: []string{"MySQL"}},
				cfgLoader: new(configfakes.FakeLoader),
			},
			args: args{
				serviceInstance: &cf.ServiceInstance{
					Service: "p.mysql",
				},
			},
			want: true,
			AssertFactoryCallCount: func(factory *fakes.FakeFactory) {
				require.Equal(t, 1, factory.NewCallCount())
			},
		},
		{
			name: "should migrate when config service names match service name regardless of casing",
			fields: fields{
				factory:   new(fakes.FakeFactory),
				helper:    &migrate.MigratorHelper{},
				cfg:       &config.Config{Services: []string{"p.MySQL"}},
				cfgLoader: new(configfakes.FakeLoader),
			},
			args: args{
				serviceInstance: &cf.ServiceInstance{
					Service: "p.mysql",
				},
			},
			want: true,
			AssertFactoryCallCount: func(factory *fakes.FakeFactory) {
				require.Equal(t, 1, factory.NewCallCount())
			},
		},
	}
	for _, tt := range tests {
		serviceInstanceMigrator := new(fakes.FakeServiceInstanceMigrator)
		tt.fields.factory.NewReturns(serviceInstanceMigrator, nil)
		t.Run(tt.name, func(t *testing.T) {
			r := migrate.NewMigratorRegistry(tt.fields.factory, tt.fields.helper, tt.fields.cfg, tt.fields.cfgLoader, tt.fields.clientHolder)
			var shouldMigrate bool
			var err error
			if _, shouldMigrate, err = r.Lookup("", "", tt.args.serviceInstance, tt.args.om, "", true); shouldMigrate != tt.want {
				t.Errorf("Lookup() = %v, want %v", shouldMigrate, tt.want)
			}
			require.NoError(t, err)
			tt.AssertFactoryCallCount(tt.fields.factory)
		})
	}
}
