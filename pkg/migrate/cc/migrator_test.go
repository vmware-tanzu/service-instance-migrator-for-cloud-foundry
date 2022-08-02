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
	"context"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/fakes"
)

func TestMigrate_Export(t *testing.T) {
	type fields struct {
		ServiceInstance        *cf.ServiceInstance
		CloudControllerService *fakes.FakeService
		Flow                   flow.Flow
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *cf.ServiceInstance
		serviceKey cf.ServiceKey
		wantErr    bool
		beforeFunc func(*fakes.FakeService)
		afterFunc  func(*cf.ServiceInstance, *fakes.FakeService)
	}{
		{
			name: "exports service instance with no service bindings",
			fields: fields{
				ServiceInstance: &cf.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-guid",
				},
				CloudControllerService: new(fakes.FakeService),
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name: "some-service-instance",
				GUID: "some-guid",
				Apps: make(map[string]string),
			},
			wantErr: false,
			afterFunc: func(want *cf.ServiceInstance, fakeCloudControllerService *fakes.FakeService) {
				_, _, si := fakeCloudControllerService.DeleteArgsForCall(0)
				require.Equal(t, 1, fakeCloudControllerService.DeleteCallCount())
				require.Equal(t, want, si)
				require.Equal(t, 0, fakeCloudControllerService.CreateCallCount())
				require.Equal(t, 0, fakeCloudControllerService.CreateServiceKeyCallCount())
			},
		},
		{
			name: "exports service instance with one service bindings",
			fields: fields{
				ServiceInstance: &cf.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-guid",
					ServiceBindings: []cf.ServiceBinding{
						{
							Guid:                "some-binding-guid",
							AppGuid:             "some-app-guid",
							ServiceInstanceGuid: "some-service-instance-guid",
						},
					},
				},
				CloudControllerService: &fakes.FakeService{
					FindAppByGUIDStub: func(s string) (string, error) {
						return "some-app-name", nil
					},
					DeleteStub: func(org, space string, instance *cf.ServiceInstance) error {
						return nil
					},
					DownloadManifestStub: func(string, string, string) (cf.Application, error) {
						return cf.Application{Name: "some-app-name"}, nil
					},
					ServiceInstanceExistsStub: func(string, string, string) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name: "some-service-instance",
				GUID: "some-guid",
				ServiceBindings: []cf.ServiceBinding{
					{
						Guid:                "some-binding-guid",
						AppGuid:             "some-app-guid",
						ServiceInstanceGuid: "some-service-instance-guid",
					},
				},
				Apps: map[string]string{"some-binding-guid": "some-app-name"},
				AppManifest: struct {
					Applications []cf.Application `yaml:"applications"`
				}{Applications: []cf.Application{
					{
						Name: "some-app-name",
					},
				}},
			},
			wantErr: false,
			afterFunc: func(want *cf.ServiceInstance, fakeCloudControllerService *fakes.FakeService) {
				_, _, si := fakeCloudControllerService.DeleteArgsForCall(0)
				require.Equal(t, 1, fakeCloudControllerService.FindAppByGUIDCallCount())
				require.Equal(t, "some-app-guid", fakeCloudControllerService.FindAppByGUIDArgsForCall(0))
				require.Equal(t, 1, fakeCloudControllerService.DeleteCallCount())
				require.Equal(t, want, si)
				require.Equal(t, 0, fakeCloudControllerService.CreateCallCount())
				require.Equal(t, 0, fakeCloudControllerService.CreateServiceKeyCallCount())
			},
		},
		{
			name: "exports service instance with multiple service bindings",
			fields: fields{
				ServiceInstance: &cf.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-guid",
					ServiceBindings: []cf.ServiceBinding{
						{
							Guid:                "one-binding-guid",
							AppGuid:             "one-app-guid",
							ServiceInstanceGuid: "one-service-instance-guid",
						},
						{
							Guid:                "two-binding-guid",
							AppGuid:             "two-app-guid",
							ServiceInstanceGuid: "two-service-instance-guid",
						},
					},
				},
				CloudControllerService: &fakes.FakeService{
					FindAppByGUIDStub: func(s string) (string, error) {
						return "some-app-name", nil
					},
					DeleteStub: func(string, string, *cf.ServiceInstance) error {
						return nil
					},
					DownloadManifestStub: func(string, string, string) (cf.Application, error) {
						return cf.Application{Name: "some-app-name"}, nil
					},
					ServiceInstanceExistsStub: func(string, string, string) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name: "some-service-instance",
				GUID: "some-guid",
				ServiceBindings: []cf.ServiceBinding{
					{
						Guid:                "one-binding-guid",
						AppGuid:             "one-app-guid",
						ServiceInstanceGuid: "one-service-instance-guid",
					},
					{
						Guid:                "two-binding-guid",
						AppGuid:             "two-app-guid",
						ServiceInstanceGuid: "two-service-instance-guid",
					},
				},
				Apps: map[string]string{
					"one-binding-guid": "some-app-name",
					"two-binding-guid": "some-app-name",
				},
				AppManifest: struct {
					Applications []cf.Application `yaml:"applications"`
				}{Applications: []cf.Application{
					{
						Name: "some-app-name-1",
					},
					{
						Name: "some-app-name-2",
					},
				}},
			},
			wantErr: false,
			beforeFunc: func(service *fakes.FakeService) {
				service.DownloadManifestReturnsOnCall(0, cf.Application{Name: "some-app-name-1"}, nil)
				service.DownloadManifestReturnsOnCall(1, cf.Application{Name: "some-app-name-2"}, nil)
			},
			afterFunc: func(want *cf.ServiceInstance, fakeCloudControllerService *fakes.FakeService) {
				_, _, si := fakeCloudControllerService.DeleteArgsForCall(0)
				require.Equal(t, 2, fakeCloudControllerService.FindAppByGUIDCallCount())
				require.Equal(t, "one-app-guid", fakeCloudControllerService.FindAppByGUIDArgsForCall(0))
				require.Equal(t, "two-app-guid", fakeCloudControllerService.FindAppByGUIDArgsForCall(1))
				require.Equal(t, 1, fakeCloudControllerService.DeleteCallCount())
				require.Equal(t, want, si)
				require.Equal(t, 0, fakeCloudControllerService.CreateCallCount())
				require.Equal(t, 0, fakeCloudControllerService.CreateServiceKeyCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc(tt.fields.CloudControllerService)
			}
			m := cc.NewMigrator(
				cc.Export("some-org", "some-space", tt.fields.CloudControllerService, tt.fields.ServiceInstance),
			)
			got, err := m.Migrate(tt.args.ctx)
			tt.afterFunc(tt.want, tt.fields.CloudControllerService)
			if (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Migrate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigrate_Import(t *testing.T) {
	type fields struct {
		ServiceInstance        *cf.ServiceInstance
		CloudControllerService *fakes.FakeService
		Flow                   flow.Flow
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *cf.ServiceInstance
		serviceKey cf.ServiceKey
		wantErr    bool
		beforeFunc func(*fakes.FakeService)
		afterFunc  func(*cf.ServiceInstance, *fakes.FakeService)
	}{
		{
			name: "imports service instance with service bindings",
			fields: fields{
				ServiceInstance: &cf.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-guid",
					ServiceBindings: []cf.ServiceBinding{
						{
							Guid:                "one-binding-guid",
							AppGuid:             "one-app-guid",
							ServiceInstanceGuid: "one-service-instance-guid",
							Credentials: map[string]interface{}{
								"uid": "one-uid",
								"pw":  "one-pw",
							},
						},
						{
							Guid:                "two-binding-guid",
							AppGuid:             "two-app-guid",
							ServiceInstanceGuid: "two-service-instance-guid",
							Credentials: map[string]interface{}{
								"uid": "two-uid",
								"pw":  "two-pw",
							},
						},
					},
					Apps: map[string]string{
						"one-binding-guid": "one-app-name",
						"two-binding-guid": "two-app-name",
					},
				},
				CloudControllerService: &fakes.FakeService{
					CreateAppStub: func(org, space, s string) (string, error) {
						return "some-app-guid", nil
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name: "some-service-instance",
				GUID: "some-guid",
				ServiceBindings: []cf.ServiceBinding{
					{
						Guid:                "one-binding-guid",
						AppGuid:             "one-app-guid",
						ServiceInstanceGuid: "one-service-instance-guid",
						Credentials: map[string]interface{}{
							"uid": "one-uid",
							"pw":  "one-pw",
						},
					},
					{
						Guid:                "two-binding-guid",
						AppGuid:             "two-app-guid",
						ServiceInstanceGuid: "two-service-instance-guid",
						Credentials: map[string]interface{}{
							"uid": "two-uid",
							"pw":  "two-pw",
						},
					},
				},
				Apps: map[string]string{
					"one-binding-guid": "one-app-name",
					"two-binding-guid": "two-app-name",
				},
			},
			wantErr: false,
			afterFunc: func(want *cf.ServiceInstance, fakeCloudControllerService *fakes.FakeService) {
				_, _, si, _ := fakeCloudControllerService.CreateArgsForCall(0)
				_, _, name0 := fakeCloudControllerService.CreateAppArgsForCall(0)
				_, _, name1 := fakeCloudControllerService.CreateAppArgsForCall(1)
				require.Equal(t, 0, fakeCloudControllerService.DeleteCallCount())
				require.Equal(t, 1, fakeCloudControllerService.CreateCallCount())
				require.Equal(t, want, si)
				require.Equal(t, 2, fakeCloudControllerService.CreateAppCallCount())
				require.Equal(t, 2, fakeCloudControllerService.CreateServiceBindingCallCount())
				require.Equal(t, "one-app-name", name0)
				require.Equal(t, "two-app-name", name1)
			},
		},
		{
			name: "imports service instance when app name is taken",
			fields: fields{
				ServiceInstance: &cf.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-guid",
					ServiceBindings: []cf.ServiceBinding{
						{
							Guid:                "one-binding-guid",
							AppGuid:             "one-app-guid",
							ServiceInstanceGuid: "one-service-instance-guid",
							Credentials: map[string]interface{}{
								"uid": "one-uid",
								"pw":  "one-pw",
							},
						},
						{
							Guid:                "two-binding-guid",
							AppGuid:             "two-app-guid",
							ServiceInstanceGuid: "two-service-instance-guid",
							Credentials: map[string]interface{}{
								"uid": "two-uid",
								"pw":  "two-pw",
							},
						},
					},
					Apps: map[string]string{
						"one-binding-guid": "one-app-name",
						"two-binding-guid": "two-app-name",
					},
				},
				CloudControllerService: &fakes.FakeService{
					CreateAppStub: func(org, space, s string) (string, error) {
						return "two-app-name", nil
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name: "some-service-instance",
				GUID: "some-guid",
				ServiceBindings: []cf.ServiceBinding{
					{
						Guid:                "one-binding-guid",
						AppGuid:             "one-app-guid",
						ServiceInstanceGuid: "one-service-instance-guid",
						Credentials: map[string]interface{}{
							"uid": "one-uid",
							"pw":  "one-pw",
						},
					},
					{
						Guid:                "two-binding-guid",
						AppGuid:             "two-app-guid",
						ServiceInstanceGuid: "two-service-instance-guid",
						Credentials: map[string]interface{}{
							"uid": "two-uid",
							"pw":  "two-pw",
						},
					},
				},
				Apps: map[string]string{
					"one-binding-guid": "one-app-name",
					"two-binding-guid": "two-app-name",
				},
			},
			beforeFunc: func(fakeCloudControllerService *fakes.FakeService) {
				fakeCloudControllerService.CreateAppReturnsOnCall(0, "", errors.New("app name already exists"))
			},
			wantErr: false,
			afterFunc: func(want *cf.ServiceInstance, fakeCloudControllerService *fakes.FakeService) {
				_, _, si, _ := fakeCloudControllerService.CreateArgsForCall(0)
				_, _, name0 := fakeCloudControllerService.CreateAppArgsForCall(0)
				_, _, name1 := fakeCloudControllerService.CreateAppArgsForCall(1)
				require.Equal(t, 0, fakeCloudControllerService.DeleteCallCount())
				require.Equal(t, 1, fakeCloudControllerService.CreateCallCount())
				require.Equal(t, want, si)
				require.Equal(t, 2, fakeCloudControllerService.CreateAppCallCount())
				require.Equal(t, 1, fakeCloudControllerService.CreateServiceBindingCallCount())
				require.Equal(t, "one-app-name", name0)
				require.Equal(t, "two-app-name", name1)
			},
		},
		{
			name: "imports service instance with service keys",
			fields: fields{
				ServiceInstance: &cf.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-guid",
					ServiceKeys: []cf.ServiceKey{
						{
							Name: "some-service-key",
							Guid: "some-guid",
							Credentials: map[string]interface{}{
								"uid": "some-uid",
								"pw":  "some-pw",
							},
							ServiceInstanceGuid: "some-service-instance-guid",
						},
					},
				},
				CloudControllerService: &fakes.FakeService{
					CreateStub: func(org, space string, instance *cf.ServiceInstance, encryptionKey string) error {
						return nil
					},
					CreateServiceKeyStub: func(cf.ServiceInstance, cf.ServiceKey) error {
						return nil
					},
				},
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &cf.ServiceInstance{
				Name: "some-service-instance",
				GUID: "some-guid",
				ServiceKeys: []cf.ServiceKey{
					{
						Name: "some-service-key",
						Guid: "some-guid",
						Credentials: map[string]interface{}{
							"uid": "some-uid",
							"pw":  "some-pw",
						},
						ServiceInstanceGuid: "some-service-instance-guid",
					},
				},
			},
			wantErr: false,
			afterFunc: func(want *cf.ServiceInstance, fakeCloudControllerService *fakes.FakeService) {
				_, _, si, _ := fakeCloudControllerService.CreateArgsForCall(0)
				require.Equal(t, 0, fakeCloudControllerService.DeleteCallCount())
				require.Equal(t, 1, fakeCloudControllerService.CreateCallCount())
				require.Equal(t, want, si)
				require.Equal(t, 1, fakeCloudControllerService.CreateServiceKeyCallCount())
				actualServiceInstance, actualServiceKey := fakeCloudControllerService.CreateServiceKeyArgsForCall(0)
				require.Equal(t, *want, actualServiceInstance)
				require.Equal(t, want.ServiceKeys[0], actualServiceKey)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc(tt.fields.CloudControllerService)
			}
			m := cc.NewMigrator(
				cc.Import("some-org", "some-space", tt.fields.CloudControllerService, tt.fields.ServiceInstance, "some-encryption-key"),
			)
			got, err := m.Migrate(tt.args.ctx)
			tt.afterFunc(tt.want, tt.fields.CloudControllerService)
			if (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Migrate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
