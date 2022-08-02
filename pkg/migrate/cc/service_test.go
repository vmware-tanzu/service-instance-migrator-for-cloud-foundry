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

package cc_test

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
	dbfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db/fakes"
	ccfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/fakes"
)

func TestNewCloudControllerService(t *testing.T) {
	type args struct {
		org              string
		space            string
		encryptionKey    string
		db               db.Repository
		cf               cf.Client
		manifestExporter cc.ManifestExporter
	}
	tests := []struct {
		name string
		args args
		want cc.DefaultCloudControllerService
	}{
		{
			name: "creates a new service",
			args: args{
				org:              "my-org",
				space:            "my-space",
				encryptionKey:    "enc-key",
				db:               new(dbfakes.FakeRepository),
				cf:               new(fakes.FakeClient),
				manifestExporter: new(ccfakes.FakeManifestExporter),
			},
			want: cc.DefaultCloudControllerService{
				Client:           new(fakes.FakeClient),
				Database:         new(dbfakes.FakeRepository),
				ManifestExporter: new(ccfakes.FakeManifestExporter),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cc.NewCloudControllerService(tt.args.db, tt.args.cf, tt.args.manifestExporter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCloudControllerService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultCloudControllerService_Create(t *testing.T) {
	type fields struct {
		Client   cf.Client
		Database db.Repository
	}
	type args struct {
		instance      *cf.ServiceInstance
		org           string
		space         string
		encryptionKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *cf.ServiceInstance
		wantErr bool
	}{
		{
			name: "creates service instance when service instance does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{}, nil
					},
					GetServiceInstanceByGuidStub: func(string) (cfclient.ServiceInstance, error) {
						return cfclient.ServiceInstance{}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{}, nil
					},
					ListServicePlansStub: func() ([]cfclient.ServicePlan, error) {
						return []cfclient.ServicePlan{
							{
								Name: "sharedVM",
							},
						}, nil
					},
					ListServicesStub: func() ([]cfclient.Service, error) {
						return []cfclient.Service{
							{
								Label: "SQLServer",
							},
						}, nil
					},
				},
				Database: &dbfakes.FakeRepository{
					CreateServiceInstanceStub: func(instance cfclient.ServiceInstance, space cfclient.Space, plan cfclient.ServicePlan, service cfclient.Service, s string) error {
						return nil
					},
					ServiceInstanceExistsStub: func(guid string) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				instance: &cf.ServiceInstance{
					Plan:    "sharedVM",
					Service: "SQLServer",
				},
				org:           "my-org",
				space:         "my-space",
				encryptionKey: "some-key",
			},
			want: &cf.ServiceInstance{
				Plan:    "sharedVM",
				Service: "SQLServer",
			},
			wantErr: false,
		},
		{
			name: "returns error when plan does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{}, nil
					},
					GetServiceInstanceByGuidStub: func(string) (cfclient.ServiceInstance, error) {
						return cfclient.ServiceInstance{}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{}, nil
					},
					ListServicePlansStub: func() ([]cfclient.ServicePlan, error) {
						return []cfclient.ServicePlan{
							{
								Name: "sharedVM",
							},
						}, nil
					},
					ListServicesStub: func() ([]cfclient.Service, error) {
						return []cfclient.Service{
							{
								Label: "SQLServer",
							},
						}, nil
					},
				},
				Database: &dbfakes.FakeRepository{
					CreateServiceInstanceStub: func(instance cfclient.ServiceInstance, space cfclient.Space, plan cfclient.ServicePlan, service cfclient.Service, s string) error {
						return nil
					},
					ServiceInstanceExistsStub: func(guid string) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				instance: &cf.ServiceInstance{
					Plan:    "some-plan",
					Service: "SQLServer",
				},
				org:           "my-org",
				space:         "my-space",
				encryptionKey: "some-key",
			},
			want: &cf.ServiceInstance{
				Plan:    "some-plan",
				Service: "SQLServer",
			},
			wantErr: true,
		},
		{
			name: "returns error when service does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{}, nil
					},
					GetServiceInstanceByGuidStub: func(string) (cfclient.ServiceInstance, error) {
						return cfclient.ServiceInstance{}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{}, nil
					},
					ListServicePlansStub: func() ([]cfclient.ServicePlan, error) {
						return []cfclient.ServicePlan{
							{
								Name: "sharedVM",
							},
						}, nil
					},
					ListServicesStub: func() ([]cfclient.Service, error) {
						return []cfclient.Service{
							{
								Label: "SQLServer",
							},
						}, nil
					},
				},
				Database: &dbfakes.FakeRepository{
					CreateServiceInstanceStub: func(instance cfclient.ServiceInstance, space cfclient.Space, plan cfclient.ServicePlan, service cfclient.Service, s string) error {
						return nil
					},
					ServiceInstanceExistsStub: func(guid string) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				instance: &cf.ServiceInstance{
					Plan:    "sharedVM",
					Service: "some-plan",
				},
				org:           "my-org",
				space:         "my-space",
				encryptionKey: "some-key",
			},
			want: &cf.ServiceInstance{
				Plan:    "sharedVM",
				Service: "some-plan",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := cc.DefaultCloudControllerService{
				Client:   tt.fields.Client,
				Database: tt.fields.Database,
			}
			err := r.Create(tt.args.org, tt.args.space, tt.args.instance, tt.args.encryptionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("MigrateInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.instance, tt.want) {
				t.Errorf("MigrateInstance() got = %v, want %v", tt.args.instance, tt.want)
			}
		})
	}
}

func TestDefaultCloudControllerService_Delete(t *testing.T) {
	type fields struct {
		Client   cf.Client
		Database db.Repository
	}
	type args struct {
		instance *cf.ServiceInstance
		org      string
		space    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "deletes a service instance",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(orgName string) (cfclient.Org, error) {
						return cfclient.Org{
							Name: orgName,
						}, nil
					},
					GetSpaceByNameStub: func(spaceName string, orgGUiD string) (cfclient.Space, error) {
						return cfclient.Space{
							Name: spaceName,
						}, nil
					},
				},
				Database: &dbfakes.FakeRepository{
					DeleteServiceInstanceStub: func(spaceGUID string, serviceInstanceGUID string) (bool, error) {
						return true, nil
					},
				},
			},
			args: args{
				instance: &cf.ServiceInstance{
					GUID: "some-guid",
				},
				org:   "my-org",
				space: "my-space",
			},
			wantErr: false,
		},
		{
			name: "returns error when service instance does not delete",
			fields: fields{
				Client:   &fakes.FakeClient{},
				Database: &dbfakes.FakeRepository{},
			},
			args: args{
				instance: &cf.ServiceInstance{
					GUID: "some-guid",
				},
				org:   "my-org",
				space: "my-space",
			},
			wantErr: true,
		},
		{
			name: "returns error when database returns error",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(orgName string) (cfclient.Org, error) {
						return cfclient.Org{
							Name: orgName,
						}, nil
					},
					GetSpaceByNameStub: func(spaceName string, orgGUiD string) (cfclient.Space, error) {
						return cfclient.Space{
							Name: spaceName,
						}, nil
					},
				},
				Database: &dbfakes.FakeRepository{
					DeleteServiceInstanceStub: func(spaceGUID string, serviceInstanceGUID string) (bool, error) {
						return false, errors.New("error when db tried to delete")
					},
				},
			},
			args: args{
				instance: &cf.ServiceInstance{
					GUID: "some-guid",
				},
				org:   "my-org",
				space: "my-space",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client:   tt.fields.Client,
				Database: tt.fields.Database,
			}
			if err := m.Delete(tt.args.org, tt.args.space, tt.args.instance); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultCloudControllerService_CreateServiceKey(t *testing.T) {
	type fields struct {
		Client   *fakes.FakeClient
		Database db.Repository
	}
	type args struct {
		si  cf.ServiceInstance
		key cf.ServiceKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    cfclient.CreateServiceKeyRequest
		wantErr bool
	}{
		{
			name: "creates a service key",
			fields: fields{
				Client: &fakes.FakeClient{
					CreateServiceKeyStub: func(request cfclient.CreateServiceKeyRequest) (cfclient.ServiceKey, error) {
						return cfclient.ServiceKey{
							Name:                "some-service-key",
							Guid:                "some-guid",
							ServiceInstanceGuid: "some-service-instance-guid",
							Credentials:         map[string]interface{}{},
							ServiceInstanceUrl:  "/v2/service-instance-guid",
						}, nil
					},
				},
				Database: &dbfakes.FakeRepository{},
			},
			args: args{
				si: cf.ServiceInstance{
					GUID:   "some-guid",
					Params: map[string]interface{}{"key1": "value1"},
				},
				key: cf.ServiceKey{
					Name:                "some-service-key",
					Guid:                "some-guid",
					ServiceInstanceGuid: "some-service-instance-guid",
					Credentials:         map[string]interface{}{},
					ServiceInstanceUrl:  "/v2/service-instance-guid",
				},
			},
			wantErr: false,
			want: cfclient.CreateServiceKeyRequest{
				Name:                "some-service-key",
				ServiceInstanceGuid: "some-service-instance-guid",
				Parameters:          map[string]interface{}{"key1": "value1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client:   tt.fields.Client,
				Database: tt.fields.Database,
			}
			if err := m.CreateServiceKey(tt.args.si, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("CreateServiceKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, tt.fields.Client.CreateServiceKeyArgsForCall(0))
		})
	}
}

func TestDefaultCloudControllerService_CreateServiceBinding(t *testing.T) {
	type fields struct {
		Client   cf.Client
		Database db.Repository
	}
	type args struct {
		binding       *cf.ServiceBinding
		appGUID       string
		encryptionKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "creates a service binding",
			fields: fields{
				Client: new(fakes.FakeClient),
				Database: &dbfakes.FakeRepository{
					CreateServiceBindingStub: func(binding cfclient.ServiceBinding, appGUID string, encryptionKey string) error {
						return nil
					},
				},
			},
			args: args{
				binding: &cf.ServiceBinding{
					Guid:                "some-guid",
					Name:                "some-name",
					AppGuid:             "",
					ServiceInstanceGuid: "some-instance-guid",
					Credentials:         map[string]interface{}{"key": "value"},
				},
				appGUID:       "some-app-guid",
				encryptionKey: "some-encryption-key",
			},
			wantErr: false,
		},
		{
			name: "returns an error when service binding cannot be saved",
			fields: fields{
				Client: new(fakes.FakeClient),
				Database: &dbfakes.FakeRepository{
					CreateServiceBindingStub: func(binding cfclient.ServiceBinding, appGUID string, encryptionKey string) error {
						return errors.New("failed to save service binding")
					},
				},
			},
			args: args{
				binding: &cf.ServiceBinding{
					Guid:                "some-guid",
					Name:                "some-name",
					AppGuid:             "",
					ServiceInstanceGuid: "some-instance-guid",
					Credentials:         map[string]interface{}{"key": "value"},
				},
				appGUID:       "some-app-guid",
				encryptionKey: "some-encryption-key",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client:   tt.fields.Client,
				Database: tt.fields.Database,
			}
			if err := m.CreateServiceBinding(tt.args.binding, tt.args.appGUID, tt.args.encryptionKey); (err != nil) != tt.wantErr {
				t.Errorf("CreateServiceBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultCloudControllerService_CreateApp(t *testing.T) {
	type fields struct {
		Client   cf.Client
		Database db.Repository
	}
	type args struct {
		name  string
		org   string
		space string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "creates an app",
			fields: fields{
				Client: &fakes.FakeClient{CreateAppStub: func(request cfclient.AppCreateRequest) (cfclient.App, error) {
					return cfclient.App{
						Name: "some-app",
						Guid: "some-app-guid",
					}, nil
				}},
			},
			args: args{
				name:  "some-app",
				org:   "some-org",
				space: "some-space",
			},
			want:    "some-app-guid",
			wantErr: false,
		},
		{
			name: "find app by returns error when org not found",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(s string) (cfclient.Org, error) {
						return cfclient.Org{}, errors.New("org not found")
					},
				},
			},
			args: args{
				name:  "some-app",
				org:   "some-org",
				space: "some-space",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "find app by returns error when space not found",
			fields: fields{
				Client: &fakes.FakeClient{
					GetSpaceByNameStub: func(s string, orgGUID string) (cfclient.Space, error) {
						return cfclient.Space{}, errors.New("space not found")
					},
				},
			},
			args: args{
				name:  "some-app",
				org:   "some-org",
				space: "some-space",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "find app by returns error when app not found",
			fields: fields{
				Client: &fakes.FakeClient{
					CreateAppStub: func(request cfclient.AppCreateRequest) (cfclient.App, error) {
						return cfclient.App{}, errors.New("app not found")
					},
				},
			},
			args: args{
				name:  "some-app",
				org:   "some-org",
				space: "some-space",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client:   tt.fields.Client,
				Database: tt.fields.Database,
			}
			got, err := m.CreateApp(tt.args.org, tt.args.space, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateApp() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultCloudControllerService_FindAppByGUID(t *testing.T) {
	type fields struct {
		Client   cf.Client
		Database db.Repository
	}
	type args struct {
		guid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "find app by guid",
			fields: fields{
				Client: &fakes.FakeClient{
					GetAppByGuidNoInlineCallStub: func(s string) (cfclient.App, error) {
						return cfclient.App{Name: "some-app", Guid: "some-app-guid"}, nil
					},
				},
			},
			args: args{
				guid: "some-app-guid",
			},
			want:    "some-app",
			wantErr: false,
		},
		{
			name: "find app by returns error",
			fields: fields{
				Client: &fakes.FakeClient{
					GetAppByGuidNoInlineCallStub: func(s string) (cfclient.App, error) {
						return cfclient.App{}, errors.New("app not found")
					},
				},
			},
			args: args{
				guid: "some-app-guid",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client:   tt.fields.Client,
				Database: tt.fields.Database,
			}
			got, err := m.FindAppByGUID(tt.args.guid)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAppByGUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindAppByGUID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultCloudControllerService_DownloadManifest(t *testing.T) {
	type fields struct {
		Client           *fakes.FakeClient
		Database         db.Repository
		ManifestExporter *ccfakes.FakeManifestExporter
	}
	type args struct {
		org     string
		space   string
		appName string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       cf.Application
		wantErr    bool
		beforeFunc func()
		afterFunc  func(*ccfakes.FakeManifestExporter, *fakes.FakeClient)
	}{
		{
			name: "exports app when no error occurs",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{Guid: "some-guid", Name: "some-org"}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{Guid: "some-guid", Name: "some-space"}, nil
					},
					AppByNameStub: func(string, string, string) (cfclient.App, error) {
						return cfclient.App{Guid: "some-guid", Name: "some-app"}, nil
					},
				},
				Database:         nil,
				ManifestExporter: new(ccfakes.FakeManifestExporter),
			},
			args: args{
				org:     "some-org",
				space:   "some-space",
				appName: "some-app",
			},
			want:    cf.Application{},
			wantErr: false,
			beforeFunc: func() {

			},
			afterFunc: func(manifestExporter *ccfakes.FakeManifestExporter, cf *fakes.FakeClient) {
				require.Equal(t, 1, manifestExporter.ExportAppManifestCallCount())
			},
		},
		{
			name: "returns error when org does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{}, errors.New("org not found")
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{Guid: "some-guid", Name: "some-space"}, nil
					},
					AppByNameStub: func(string, string, string) (cfclient.App, error) {
						return cfclient.App{Guid: "some-guid", Name: "some-app"}, nil
					},
				},
				Database:         nil,
				ManifestExporter: new(ccfakes.FakeManifestExporter),
			},
			args: args{
				org:     "some-org",
				space:   "some-space",
				appName: "some-app",
			},
			want:    cf.Application{},
			wantErr: true,
			beforeFunc: func() {

			},
			afterFunc: func(manifestExporter *ccfakes.FakeManifestExporter, cf *fakes.FakeClient) {
				require.Equal(t, 1, cf.GetOrgByNameCallCount())
				require.Equal(t, 0, cf.GetSpaceByNameCallCount())
				require.Equal(t, 0, cf.AppByNameCallCount())
				require.Equal(t, 0, manifestExporter.ExportAppManifestCallCount())
			},
		},
		{
			name: "returns error when space does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{Guid: "some-guid", Name: "some-org"}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{}, errors.New("space not found")
					},
					AppByNameStub: func(string, string, string) (cfclient.App, error) {
						return cfclient.App{Guid: "some-guid", Name: "some-app"}, nil
					},
				},
				Database:         nil,
				ManifestExporter: new(ccfakes.FakeManifestExporter),
			},
			args: args{
				org:     "some-org",
				space:   "some-space",
				appName: "some-app",
			},
			want:    cf.Application{},
			wantErr: true,
			beforeFunc: func() {

			},
			afterFunc: func(manifestExporter *ccfakes.FakeManifestExporter, cf *fakes.FakeClient) {
				require.Equal(t, 1, cf.GetOrgByNameCallCount())
				require.Equal(t, 1, cf.GetSpaceByNameCallCount())
				require.Equal(t, 0, cf.AppByNameCallCount())
				require.Equal(t, 0, manifestExporter.ExportAppManifestCallCount())
			},
		},
		{
			name: "returns error when app does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(string) (cfclient.Org, error) {
						return cfclient.Org{Guid: "some-guid", Name: "some-org"}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{Guid: "some-guid", Name: "some-space"}, nil
					},
					AppByNameStub: func(string, string, string) (cfclient.App, error) {
						return cfclient.App{}, errors.New("app not found")
					},
				},
				Database:         nil,
				ManifestExporter: new(ccfakes.FakeManifestExporter),
			},
			args: args{
				org:     "some-org",
				space:   "some-space",
				appName: "some-app",
			},
			want:    cf.Application{},
			wantErr: true,
			beforeFunc: func() {

			},
			afterFunc: func(manifestExporter *ccfakes.FakeManifestExporter, cf *fakes.FakeClient) {
				require.Equal(t, 1, cf.GetOrgByNameCallCount())
				require.Equal(t, 1, cf.GetSpaceByNameCallCount())
				require.Equal(t, 1, cf.AppByNameCallCount())
				require.Equal(t, 0, manifestExporter.ExportAppManifestCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client:           tt.fields.Client,
				Database:         tt.fields.Database,
				ManifestExporter: tt.fields.ManifestExporter,
			}
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			got, err := m.DownloadManifest(tt.args.org, tt.args.space, tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DownloadManifest() got = %v, want %v", got, tt.want)
			}
			if tt.afterFunc != nil {
				tt.afterFunc(tt.fields.ManifestExporter, tt.fields.Client)
			}
		})
	}
}

func TestDefaultCloudControllerService_ServiceInstanceExists(t *testing.T) {
	type fields struct {
		Client *fakes.FakeClient
	}
	type args struct {
		org          string
		space        string
		instanceName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      bool
		wantErr   bool
		afterFunc func(*fakes.FakeClient)
	}{
		{
			name: "returns true if service instance exists",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(s string) (cfclient.Org, error) {
						return cfclient.Org{Guid: "some-guid", Name: "some-org"}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{Guid: "some-guid", Name: "some-space"}, nil
					},
					ListServiceInstancesByQueryStub: func(url.Values) ([]cfclient.ServiceInstance, error) {
						return []cfclient.ServiceInstance{
							{
								Name: "some-service",
							},
						}, nil
					},
				},
			},
			args: args{
				org:          "some-org",
				space:        "some-space",
				instanceName: "",
			},
			want:    true,
			wantErr: false,
			afterFunc: func(cf *fakes.FakeClient) {
				require.Equal(t, 1, cf.GetOrgByNameCallCount())
				require.Equal(t, 1, cf.GetSpaceByNameCallCount())
				require.Equal(t, 1, cf.ListServiceInstancesByQueryCallCount())
			},
		},
		{
			name: "returns false if service instance does not exist",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(s string) (cfclient.Org, error) {
						return cfclient.Org{Guid: "some-guid", Name: "some-org"}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{Guid: "some-guid", Name: "some-space"}, nil
					},
				},
			},
			args: args{
				org:          "some-org",
				space:        "some-space",
				instanceName: "",
			},
			want:    false,
			wantErr: false,
			afterFunc: func(cf *fakes.FakeClient) {
				require.Equal(t, 1, cf.GetOrgByNameCallCount())
				require.Equal(t, 1, cf.GetSpaceByNameCallCount())
			},
		},
		{
			name: "returns false if error occurs",
			fields: fields{
				Client: &fakes.FakeClient{
					GetOrgByNameStub: func(s string) (cfclient.Org, error) {
						return cfclient.Org{Guid: "some-guid", Name: "some-org"}, nil
					},
					GetSpaceByNameStub: func(string, string) (cfclient.Space, error) {
						return cfclient.Space{Guid: "some-guid", Name: "some-space"}, nil
					},
					ListServiceInstancesByQueryStub: func(url.Values) ([]cfclient.ServiceInstance, error) {
						return []cfclient.ServiceInstance{}, errors.New("some-error")
					},
				},
			},
			args: args{
				org:          "some-org",
				space:        "some-space",
				instanceName: "",
			},
			want:    false,
			wantErr: true,
			afterFunc: func(cf *fakes.FakeClient) {
				require.Equal(t, 1, cf.GetOrgByNameCallCount())
				require.Equal(t, 1, cf.GetSpaceByNameCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.DefaultCloudControllerService{
				Client: tt.fields.Client,
			}
			got, err := m.ServiceInstanceExists(tt.args.org, tt.args.space, tt.args.instanceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceInstanceExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ServiceInstanceExists() got = %v, want %v", got, tt.want)
			}
			if tt.afterFunc != nil {
				tt.afterFunc(tt.fields.Client)
			}
		})
	}
}
