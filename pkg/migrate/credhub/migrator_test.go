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

package credhub_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	configfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	flowfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/credhub"
)

func TestMigrator_Migrate(t *testing.T) {
	type fields struct {
		flow            *flowfakes.FakeFlow
		migrationReader *configfakes.FakeMigrationReader
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "runs migration flow",
			fields: fields{
				flow: &flowfakes.FakeFlow{
					RunStub: func(ctx context.Context, i interface{}, b bool) (flow.Result, error) {
						return &cf.ServiceInstance{Name: "some-instance"}, nil
					},
				},
				migrationReader: new(configfakes.FakeMigrationReader),
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := credhub.NewMigrator(tt.fields.flow, tt.fields.migrationReader)
			if _, err := m.Migrate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, 1, tt.fields.flow.RunCallCount())
		})
	}
}

func TestMigrator_Validate(t *testing.T) {
	type fields struct {
		flow            *flowfakes.FakeFlow
		migrationReader *configfakes.FakeMigrationReader
	}
	type args struct {
		si     *cf.ServiceInstance
		export bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "credentials not set on credhub service instance",
			fields: fields{
				flow:            new(flowfakes.FakeFlow),
				migrationReader: new(configfakes.FakeMigrationReader),
			},
			args: args{
				si: &cf.ServiceInstance{
					Name: "some-instance",
				},
				export: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := credhub.NewMigrator(tt.fields.flow, tt.fields.migrationReader)
			if err := m.Validate(tt.args.si, tt.args.export); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMigrator(t *testing.T) {
	type args struct {
		flow *flowfakes.FakeFlow
		mr   *configfakes.FakeMigrationReader
	}
	tests := []struct {
		name string
		args args
		want *credhub.Migrator
	}{
		{
			name: "",
			args: args{
				flow: new(flowfakes.FakeFlow),
				mr:   new(configfakes.FakeMigrationReader),
			},
			want: &credhub.Migrator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := credhub.NewMigrator(tt.args.flow, tt.args.mr)
			if diff := cmp.Diff(tt.want, got,
				cmpopts.IgnoreUnexported(credhub.Migrator{}),
			); diff != "" {
				t.Errorf("NewMigrator() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
