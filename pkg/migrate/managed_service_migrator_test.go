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
	"net/url"
	"testing"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	cffakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
)

func TestManagedServiceMigrator_Migrate(t *testing.T) {
	type fields struct {
		config       *config.Config
		instance     *cf.ServiceInstance
		Org          string
		Space        string
		ClientHolder *fakes.FakeClientHolder
		isExport     bool
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		cfClient  *cffakes.FakeClient
		fields    fields
		args      args
		want      *cf.ServiceInstance
		wantErr   bool
		afterFunc func(t *testing.T, fakeClient *cffakes.FakeClient)
	}{
		{
			name: "creates a service when service does not exist",
			cfClient: &cffakes.FakeClient{
				ListServicePlansStub: func() ([]cfclient.ServicePlan, error) {
					return []cfclient.ServicePlan{
						{
							Name: "some-plan",
						},
					}, nil
				},
			},
			fields: fields{
				config: &config.Config{
					DomainsToReplace: map[string]string{"cf1.example.com": "cf2.example.com"},
					DryRun:           false,
				},
				instance: &cf.ServiceInstance{
					Name:   "some-service",
					Tags:   "tag1,tag2",
					Params: map[string]interface{}{"username": "some-user", "password": "some-password"},
					Plan:   "some-plan",
				},
				Org:          "some-org",
				Space:        "some-space",
				ClientHolder: new(fakes.FakeClientHolder),
				isExport:     false,
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name:   "some-service",
				Tags:   "tag1,tag2",
				Params: map[string]interface{}{"username": "some-user", "password": "some-password"},
				Plan:   "some-plan",
			},
			wantErr: false,
			afterFunc: func(t *testing.T, fakeClient *cffakes.FakeClient) {
				require.Equal(t, 1, fakeClient.CreateServiceInstanceCallCount())
				require.Equal(t, 0, fakeClient.UpdateSICallCount())
			},
		},
		{
			name: "updates a service when service exists",
			cfClient: &cffakes.FakeClient{
				ListServiceInstancesByQueryStub: func(values url.Values) ([]cfclient.ServiceInstance, error) {
					return []cfclient.ServiceInstance{
						{
							Name: "some-service",
						},
					}, nil
				},
			},
			fields: fields{
				config: &config.Config{
					DomainsToReplace: map[string]string{"cf1.example.com": "cf2.example.com"},
					DryRun:           false,
				},
				instance: &cf.ServiceInstance{
					Name:   "some-service",
					Tags:   "tag1,tag2",
					Params: map[string]interface{}{"username": "some-user", "password": "some-password"},
				},
				Org:          "some-org",
				Space:        "some-space",
				ClientHolder: new(fakes.FakeClientHolder),
				isExport:     false,
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name:   "some-service",
				Tags:   "tag1,tag2",
				Params: map[string]interface{}{"username": "some-user", "password": "some-password"},
			},
			wantErr: false,
			afterFunc: func(t *testing.T, fakeClient *cffakes.FakeClient) {
				require.Equal(t, 0, fakeClient.CreateServiceInstanceCallCount())
				require.Equal(t, 1, fakeClient.UpdateSICallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.ClientHolder.CFClientStub = func(bool) cf.Client {
				return tt.cfClient
			}
			got, err := flow.RunWith(migrate.CreateService(tt.fields.Org,
				tt.fields.Space,
				tt.fields.instance,
				tt.fields.ClientHolder,
				tt.fields.isExport), config.ContextWithConfig(context.TODO(), tt.fields.config), nil, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			cmp.Diff(got, tt.want)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CreateService() mismatch (-want +got):\n%s", diff)
			}
			tt.afterFunc(t, tt.cfClient)
		})
	}
}
