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
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewOrgImporter(t *testing.T) {
	type args struct {
		spaceImporter *migrate.SpaceImporter
		includedOrgs  []string
		excludedOrgs  []string
	}
	tests := []struct {
		name string
		args args
		want *migrate.OrgImporter
	}{
		{
			name: "create a new org importer",
			args: args{
				spaceImporter: &migrate.SpaceImporter{
					ServiceInstanceImporter: new(fakes.FakeServiceInstanceImporter),
				},
				includedOrgs: []string{"org3", "org4"},
				excludedOrgs: []string{"org1", "org2"},
			},
			want: &migrate.OrgImporter{
				SpaceImporter: &migrate.SpaceImporter{
					ServiceInstanceImporter: new(fakes.FakeServiceInstanceImporter),
				},
				IncludedOrgs: []string{"org3", "org4"},
				ExcludedOrgs: []string{"org1", "org2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrate.NewOrgImporter(tt.args.spaceImporter, tt.args.includedOrgs, tt.args.excludedOrgs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOrgImporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrgImporter_Import(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	type fields struct {
		excludedOrgs  []string
		SpaceImporter *migrate.SpaceImporter
	}
	type args struct {
		ctx  context.Context
		dir  string
		orgs []string
		om   config.OpsManager
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "imports orgs",
			fields: fields{
				SpaceImporter: &migrate.SpaceImporter{
					ServiceInstanceImporter: new(fakes.FakeServiceInstanceImporter),
				},
				excludedOrgs: []string{"org1", "org2"},
			},
			args: args{
				ctx:  context.TODO(),
				dir:  filepath.Join(pwd, "testdata"),
				orgs: nil,
				om:   config.OpsManager{Hostname: "opsman.example.com"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := migrate.OrgImporter{
				ExcludedOrgs:  tt.fields.excludedOrgs,
				SpaceImporter: tt.fields.SpaceImporter,
			}
			if err := e.Import(tt.args.ctx, tt.args.om, tt.args.dir, tt.args.orgs...); (err != nil) != tt.wantErr {
				t.Errorf("Import() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrgImporter_excludeOrg(t *testing.T) {
	type fields struct {
		excludedOrgs  []string
		SpaceImporter *migrate.SpaceImporter
	}
	type args struct {
		orgName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "exclude orgs",
			fields: fields{
				SpaceImporter: &migrate.SpaceImporter{},
				excludedOrgs:  []string{"org1", "org2"},
			},
			args: args{
				orgName: "org2",
			},
			want: true,
		},
		{
			name: "does not exclude orgs",
			fields: fields{
				SpaceImporter: &migrate.SpaceImporter{},
				excludedOrgs:  []string{"org1", "org2"},
			},
			args: args{
				orgName: "org3",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := migrate.OrgImporter{
				ExcludedOrgs:  tt.fields.excludedOrgs,
				SpaceImporter: tt.fields.SpaceImporter,
			}
			if got := e.ExcludeOrg(tt.args.orgName); got != tt.want {
				t.Errorf("ExcludeOrg() = %v, want %v", got, tt.want)
			}
		})
	}
}
