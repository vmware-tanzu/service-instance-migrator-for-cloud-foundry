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

package cmd_test

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"reflect"
	"testing"
)

func TestFactory_NewOrgExporter(t *testing.T) {
	type fields struct {
		cfg    *config.Config
		holder *fakes.FakeClientHolder
	}
	type args struct {
		exporter *fakes.FakeServiceInstanceExporter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *migrate.OrgExporter
	}{
		{
			name: "create a new org exporter",
			fields: fields{
				cfg: &config.Config{
					ExcludedOrgs: []string{},
				},
				holder: new(fakes.FakeClientHolder),
			},
			args: args{
				exporter: &fakes.FakeServiceInstanceExporter{},
			},
			want: &migrate.OrgExporter{
				ClientHolder:            new(fakes.FakeClientHolder),
				ExcludedOrgs:            []string{},
				ServiceInstanceExporter: &fakes.FakeServiceInstanceExporter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa := cmd.ExportMigratorFactory{
				Config:       tt.fields.cfg,
				ClientHolder: tt.fields.holder,
			}
			if got := fa.NewOrgExporter(tt.args.exporter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOrgExporter() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFactory_NewSpaceExporter(t *testing.T) {
	type fields struct {
		cfg    *config.Config
		holder *fakes.FakeClientHolder
	}
	type args struct {
		exporter *fakes.FakeServiceInstanceExporter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *migrate.SpaceExporter
	}{
		{
			name: "create a new space exporter",
			fields: fields{
				cfg:    &config.Config{},
				holder: new(fakes.FakeClientHolder),
			},
			args: args{
				exporter: &fakes.FakeServiceInstanceExporter{},
			},
			want: &migrate.SpaceExporter{
				ClientHolder:            new(fakes.FakeClientHolder),
				ServiceInstanceExporter: &fakes.FakeServiceInstanceExporter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa := cmd.ExportMigratorFactory{
				Config:       tt.fields.cfg,
				ClientHolder: tt.fields.holder,
			}
			if got := fa.NewSpaceExporter(tt.args.exporter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpaceExporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFactory_NewOrgImporter(t *testing.T) {
	type fields struct {
		cfg    *config.Config
		holder *fakes.FakeClientHolder
	}
	type args struct {
		si *fakes.FakeServiceInstanceImporter
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   *migrate.OrgImporter
	}{
		{
			name: "create a new org importer",
			args: args{
				si: &fakes.FakeServiceInstanceImporter{},
			},
			fields: fields{
				cfg: &config.Config{
					ExcludedOrgs: []string{},
				},
				holder: new(fakes.FakeClientHolder),
			},
			want: &migrate.OrgImporter{
				SpaceImporter: &migrate.SpaceImporter{
					ServiceInstanceImporter: &fakes.FakeServiceInstanceImporter{},
				},
				ExcludedOrgs: []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa := cmd.ImportMigratorFactory{
				Config:       tt.fields.cfg,
				ClientHolder: tt.fields.holder,
			}
			if got := fa.NewOrgImporter(tt.args.si); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOrgImporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFactory_NewSpaceImporter(t *testing.T) {
	type fields struct {
		cfg    *config.Config
		holder *fakes.FakeClientHolder
	}
	type args struct {
		importer *fakes.FakeServiceInstanceImporter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *migrate.SpaceImporter
	}{
		{
			name: "create a new space importer",
			fields: fields{
				cfg:    &config.Config{},
				holder: new(fakes.FakeClientHolder),
			},
			args: args{
				importer: &fakes.FakeServiceInstanceImporter{},
			},
			want: &migrate.SpaceImporter{
				ServiceInstanceImporter: &fakes.FakeServiceInstanceImporter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa := cmd.ImportMigratorFactory{
				Config:       tt.fields.cfg,
				ClientHolder: tt.fields.holder,
			}
			if got := fa.NewSpaceImporter(tt.args.importer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpaceImporter() = %v, want %v", got, tt.want)
			}
		})
	}
}
