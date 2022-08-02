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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestNewImporter(t *testing.T) {
	type args struct {
		cf         cf.Client
		skipUpdate bool
		writer     io.Writer
		sii        *fakes.FakeServiceInstanceImporter
	}
	tests := []struct {
		name string
		args args
		want *migrate.SpaceImporter
	}{
		{
			name: "creates a new SpaceImporter",
			args: args{
				cf:         nil,
				skipUpdate: false,
				writer:     nil,
				sii:        new(fakes.FakeServiceInstanceImporter),
			},
			want: &migrate.SpaceImporter{
				ServiceInstanceImporter: new(fakes.FakeServiceInstanceImporter),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrate.NewSpaceImporter(tt.args.sii); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpaceImporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceImporter_Import(t *testing.T) {
	pwd, _ := os.Getwd()
	type fields struct {
		skipUpdate              bool
		serviceInstanceImporter *fakes.FakeServiceInstanceImporter
	}
	type args struct {
		dir   string
		org   string
		space string
		om    config.OpsManager
		cfg   *config.Config
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		afterFunc func(*fakes.FakeServiceInstanceImporter)
	}{
		{
			name: "import service instances",
			fields: fields{
				skipUpdate:              false,
				serviceInstanceImporter: new(fakes.FakeServiceInstanceImporter),
			},
			args: args{
				dir:   pwd + "/testdata",
				org:   "cloudfoundry",
				space: "test-app",
				om: config.OpsManager{
					Hostname: "opsman.url.com",
				},
				cfg: &config.Config{
					DomainsToReplace: map[string]string{"apps.cf1.example.com": "apps.cf2.example.com"},
					DryRun:           false,
				},
			},
			afterFunc: func(i *fakes.FakeServiceInstanceImporter) {
				require.Equal(t, 2, i.ImportManagedServiceCallCount())
			},
			wantErr: false,
		},
		{
			name: "import only migrates specific instances",
			fields: fields{
				skipUpdate:              false,
				serviceInstanceImporter: new(fakes.FakeServiceInstanceImporter),
			},
			args: args{
				dir:   pwd + "/testdata",
				org:   "cloudfoundry",
				space: "test-app",
				om: config.OpsManager{
					Hostname: "opsman.url.com",
				},
				cfg: &config.Config{
					DomainsToReplace: map[string]string{"apps.cf1.example.com": "apps.cf2.example.com"},
					DryRun:           false,
					Instances:        []string{"sql-test"},
				},
			},
			afterFunc: func(i *fakes.FakeServiceInstanceImporter) {
				require.Equal(t, 1, i.ImportManagedServiceCallCount())
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := migrate.SpaceImporter{
				ServiceInstanceImporter: tt.fields.serviceInstanceImporter,
			}
			ctx := config.ContextWithConfig(context.TODO(), tt.args.cfg)
			if err := e.Import(ctx, tt.args.om, tt.args.dir, tt.args.org, tt.args.space); (err != nil) != tt.wantErr {
				t.Errorf("Import() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.afterFunc(tt.fields.serviceInstanceImporter)
		})
	}
}
