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
	"bytes"
	"context"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	cffakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

func TestNewServiceInstanceExporter(t *testing.T) {
	type args struct {
		holder   *fakes.FakeClientHolder
		registry *fakes.FakeMigratorRegistry
		parser   *fakes.FakeServiceInstanceParser
		cfg      *config.Config
	}
	tests := []struct {
		name string
		args args
		want *migrate.DefaultServiceInstanceExporter
	}{
		{
			name: "create a new service instance exporter",
			args: args{
				holder:   new(fakes.FakeClientHolder),
				registry: new(fakes.FakeMigratorRegistry),
				parser:   new(fakes.FakeServiceInstanceParser),
			},
			want: &migrate.DefaultServiceInstanceExporter{
				ClientHolder: new(fakes.FakeClientHolder),
				Registry:     new(fakes.FakeMigratorRegistry),
				Parser:       new(fakes.FakeServiceInstanceParser),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrate.NewServiceInstanceExporter(tt.args.cfg, tt.args.holder, tt.args.registry, tt.args.parser); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServiceInstanceExporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceInstanceExporter_ExportManagedServices(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	type fields struct {
		holder   *fakes.FakeClientHolder
		registry *fakes.FakeMigratorRegistry
		parser   *fakes.FakeServiceInstanceParser
		cfg      *config.Config
	}
	type args struct {
		ctx   context.Context
		org   cfclient.Org
		space cfclient.Space
		dir   string
		om    config.OpsManager
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		cfClient   *cffakes.FakeClient
		beforeFunc func(context.Context) context.Context
		afterFunc  func(context.Context, *fakes.FakeClientHolder, *fakes.FakeMigratorRegistry, *fakes.FakeServiceInstanceParser)
	}{
		{
			name: "export a managed service with service bindings",
			cfClient: &cffakes.FakeClient{
				GetClientConfigStub: func() *cfclient.Config {
					return cfclient.DefaultConfig()
				},
				ListServiceBindingsByQueryStub: func(values url.Values) ([]cfclient.ServiceBinding, error) {
					return []cfclient.ServiceBinding{
						{
							Guid:                "some-guid",
							Name:                "some-instance",
							AppGuid:             "some-app-guid",
							ServiceInstanceGuid: "some-service-instance-guid",
							Credentials:         map[string]interface{}{"hostname": "some-host", "username": "some-username", "password": "some-password"},
						},
					}, nil
				},
				ListSpaceServiceInstancesStub: func(spaceGUID string) ([]cfclient.ServiceInstance, error) {
					return []cfclient.ServiceInstance{{
						Name:            "some-instance",
						ServicePlanGuid: "some-plan-guid",
						Type:            "managed_service_instance",
						ServiceGuid:     "some-service-guid",
						Guid:            spaceGUID,
					}}, nil
				},
				GetServiceInstanceParamsStub: func(string) (map[string]interface{}, error) {
					return map[string]interface{}{"param1": "value1"}, nil
				},
				GetServicePlanByGUIDStub: func(string) (*cfclient.ServicePlan, error) {
					return &cfclient.ServicePlan{
						Name: "some-plan",
					}, nil
				},
				GetServiceByGuidStub: func(string) (cfclient.Service, error) {
					return cfclient.Service{
						Label: "some-service",
					}, nil
				},
			},
			fields: fields{
				cfg:    &config.Config{},
				holder: new(fakes.FakeClientHolder),
				registry: &fakes.FakeMigratorRegistry{
					LookupStub: func(org string, space string, instance *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (migrate.ServiceInstanceMigrator, bool, error) {
						return &fakes.FakeServiceInstanceMigrator{
							MigrateStub: func(ctx context.Context) (*cf.ServiceInstance, error) {
								return &cf.ServiceInstance{
									AppManifest: cf.Manifest{Applications: []cf.Application{
										{
											Name: "some-app",
										},
									}},
								}, nil
							},
						}, true, nil
					},
				},
				parser: new(fakes.FakeServiceInstanceParser),
			},
			args: args{
				ctx: context.TODO(),
				org: cfclient.Org{
					Name: "",
				},
				space: cfclient.Space{
					Name: "",
				},
				dir: pwd + "/testdata",
			},
			wantErr: false,
			beforeFunc: func(ctx context.Context) context.Context {
				return config.ContextWithSummary(ctx, report.NewSummary(&bytes.Buffer{}))
			},
			afterFunc: func(ctx context.Context, fakeHolder *fakes.FakeClientHolder, r *fakes.FakeMigratorRegistry, p *fakes.FakeServiceInstanceParser) {
				cfClient := fakeHolder.SourceCFClient().(*cffakes.FakeClient)
				require.Equal(t, 1, cfClient.ListServiceBindingsByQueryCallCount())
				require.Equal(t, 1, cfClient.ListServiceKeysByQueryCallCount())
				require.Equal(t, 1, cfClient.ListSpaceServiceInstancesCallCount())
				require.Equal(t, 1, cfClient.GetServiceInstanceParamsCallCount())
				require.Equal(t, 1, cfClient.GetServicePlanByGUIDCallCount())
				require.Equal(t, 1, cfClient.GetServiceByGuidCallCount())
				require.Equal(t, 1, r.LookupCallCount())
				require.Equal(t, 2, p.MarshalCallCount())
				s, ok := config.SummaryFromContext(ctx)
				require.True(t, ok)
				require.Equal(t, 0, s.ServiceSkippedCount())
				require.Equal(t, 1, s.ServiceSuccessCount())
			},
		},
		{
			name: "export a managed service with no bindings",
			cfClient: &cffakes.FakeClient{
				GetClientConfigStub: func() *cfclient.Config {
					return cfclient.DefaultConfig()
				},
				ListServiceBindingsByQueryStub: func(values url.Values) ([]cfclient.ServiceBinding, error) {
					return []cfclient.ServiceBinding{}, nil
				},
				ListSpaceServiceInstancesStub: func(spaceGUID string) ([]cfclient.ServiceInstance, error) {
					return []cfclient.ServiceInstance{{
						Name:            "some-instance",
						ServicePlanGuid: "some-plan-guid",
						Type:            "managed_service_instance",
						ServiceGuid:     "some-service-guid",
						Guid:            spaceGUID,
					}}, nil
				},
				GetServiceInstanceParamsStub: func(string) (map[string]interface{}, error) {
					return map[string]interface{}{"param1": "value1"}, nil
				},
				GetServicePlanByGUIDStub: func(string) (*cfclient.ServicePlan, error) {
					return &cfclient.ServicePlan{
						Name: "some-plan",
					}, nil
				},
				GetServiceByGuidStub: func(string) (cfclient.Service, error) {
					return cfclient.Service{
						Label: "some-service",
					}, nil
				},
			},
			fields: fields{
				cfg:    &config.Config{},
				holder: new(fakes.FakeClientHolder),
				registry: &fakes.FakeMigratorRegistry{
					LookupStub: func(org string, space string, instance *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (migrate.ServiceInstanceMigrator, bool, error) {
						return &fakes.FakeServiceInstanceMigrator{
							MigrateStub: func(ctx context.Context) (*cf.ServiceInstance, error) {
								return &cf.ServiceInstance{}, nil
							},
						}, true, nil
					},
				},
				parser: new(fakes.FakeServiceInstanceParser),
			},
			args: args{
				ctx: context.TODO(),
				org: cfclient.Org{
					Name: "some-org",
				},
				space: cfclient.Space{
					Name: "some-space",
				},
				dir: pwd + "/testdata",
			},
			wantErr: false,
			beforeFunc: func(ctx context.Context) context.Context {
				return config.ContextWithSummary(ctx, report.NewSummary(&bytes.Buffer{}))
			},
			afterFunc: func(ctx context.Context, fakeHolder *fakes.FakeClientHolder, r *fakes.FakeMigratorRegistry, p *fakes.FakeServiceInstanceParser) {
				cfClient := fakeHolder.SourceCFClient().(*cffakes.FakeClient)
				require.Equal(t, 1, cfClient.ListServiceBindingsByQueryCallCount())
				require.Equal(t, 1, cfClient.ListServiceKeysByQueryCallCount())
				require.Equal(t, 1, cfClient.ListSpaceServiceInstancesCallCount())
				require.Equal(t, 1, cfClient.GetServiceInstanceParamsCallCount())
				require.Equal(t, 1, cfClient.GetServicePlanByGUIDCallCount())
				require.Equal(t, 1, cfClient.GetServiceByGuidCallCount())
				require.Equal(t, 1, r.LookupCallCount())
				require.Equal(t, 1, p.MarshalCallCount())
				s, ok := config.SummaryFromContext(ctx)
				require.True(t, ok)
				require.Equal(t, 0, s.ServiceSkippedCount())
				require.Equal(t, 1, s.ServiceSuccessCount())
			},
		},
		{
			name: "export returns skipped migrations",
			cfClient: &cffakes.FakeClient{
				GetClientConfigStub: func() *cfclient.Config {
					return cfclient.DefaultConfig()
				},
				ListServiceBindingsByQueryStub: func(values url.Values) ([]cfclient.ServiceBinding, error) {
					return []cfclient.ServiceBinding{}, nil
				},
				ListSpaceServiceInstancesStub: func(spaceGUID string) ([]cfclient.ServiceInstance, error) {
					return []cfclient.ServiceInstance{{
						Name:            "some-instance",
						ServicePlanGuid: "some-plan-guid",
						Type:            "managed_service_instance",
						ServiceGuid:     "some-service-guid",
						Guid:            spaceGUID,
					}}, nil
				},
				GetServiceInstanceParamsStub: func(string) (map[string]interface{}, error) {
					return map[string]interface{}{"param1": "value1"}, nil
				},
				GetServicePlanByGUIDStub: func(string) (*cfclient.ServicePlan, error) {
					return &cfclient.ServicePlan{
						Name: "some-plan",
					}, nil
				},
				GetServiceByGuidStub: func(string) (cfclient.Service, error) {
					return cfclient.Service{
						Label: "some-service",
					}, nil
				},
			},
			fields: fields{
				cfg:    &config.Config{},
				holder: new(fakes.FakeClientHolder),
				registry: &fakes.FakeMigratorRegistry{
					LookupStub: func(org string, space string, instance *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (migrate.ServiceInstanceMigrator, bool, error) {
						return &fakes.FakeServiceInstanceMigrator{
							MigrateStub: func(ctx context.Context) (*cf.ServiceInstance, error) {
								return &cf.ServiceInstance{}, nil
							},
						}, false, nil
					},
				},
				parser: new(fakes.FakeServiceInstanceParser),
			},
			args: args{
				ctx: context.TODO(),
				org: cfclient.Org{
					Name: "some-org",
				},
				space: cfclient.Space{
					Name: "some-space",
				},
				dir: pwd + "/testdata",
			},
			wantErr: false,
			beforeFunc: func(ctx context.Context) context.Context {
				return config.ContextWithSummary(ctx, report.NewSummary(&bytes.Buffer{}))
			},
			afterFunc: func(ctx context.Context, fakeHolder *fakes.FakeClientHolder, r *fakes.FakeMigratorRegistry, p *fakes.FakeServiceInstanceParser) {
				cfClient := fakeHolder.SourceCFClient().(*cffakes.FakeClient)
				require.Equal(t, 1, cfClient.ListServiceBindingsByQueryCallCount())
				require.Equal(t, 1, cfClient.ListServiceKeysByQueryCallCount())
				require.Equal(t, 1, cfClient.ListSpaceServiceInstancesCallCount())
				require.Equal(t, 1, cfClient.GetServiceInstanceParamsCallCount())
				require.Equal(t, 1, cfClient.GetServicePlanByGUIDCallCount())
				require.Equal(t, 1, cfClient.GetServiceByGuidCallCount())
				require.Equal(t, 1, r.LookupCallCount())
				require.Equal(t, 0, p.MarshalCallCount())
				s, ok := config.SummaryFromContext(ctx)
				require.True(t, ok)
				require.Equal(t, 1, s.ServiceSkippedCount())
				require.Equal(t, 0, s.ServiceSuccessCount())
			},
		},
		{
			name: "export only migrates specific instances",
			cfClient: &cffakes.FakeClient{
				GetClientConfigStub: func() *cfclient.Config {
					return cfclient.DefaultConfig()
				},
				ListServiceBindingsByQueryStub: func(values url.Values) ([]cfclient.ServiceBinding, error) {
					return []cfclient.ServiceBinding{}, nil
				},
				ListSpaceServiceInstancesStub: func(spaceGUID string) ([]cfclient.ServiceInstance, error) {
					return []cfclient.ServiceInstance{
						{
							Name:            "some-instance1",
							ServicePlanGuid: "some-plan1-guid",
							Type:            "managed_service_instance",
							ServiceGuid:     "some-service1-guid",
							Guid:            spaceGUID,
						},
						{
							Name:            "some-instance2",
							ServicePlanGuid: "some-plan2-guid",
							Type:            "managed_service_instance",
							ServiceGuid:     "some-service2-guid",
							Guid:            spaceGUID,
						},
					}, nil
				},
				GetServiceInstanceParamsStub: func(string) (map[string]interface{}, error) {
					return map[string]interface{}{"param1": "value1"}, nil
				},
				GetServicePlanByGUIDStub: func(string) (*cfclient.ServicePlan, error) {
					return &cfclient.ServicePlan{
						Name: "some-plan1",
					}, nil
				},
				GetServiceByGuidStub: func(string) (cfclient.Service, error) {
					return cfclient.Service{
						Label: "some-service1",
					}, nil
				},
			},
			fields: fields{
				cfg:    &config.Config{Instances: []string{"some-instance2"}},
				holder: new(fakes.FakeClientHolder),
				registry: &fakes.FakeMigratorRegistry{
					LookupStub: func(org string, space string, instance *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (migrate.ServiceInstanceMigrator, bool, error) {
						return &fakes.FakeServiceInstanceMigrator{
							MigrateStub: func(ctx context.Context) (*cf.ServiceInstance, error) {
								return &cf.ServiceInstance{}, nil
							},
						}, true, nil
					},
				},
				parser: new(fakes.FakeServiceInstanceParser),
			},
			args: args{
				ctx: context.TODO(),
				org: cfclient.Org{
					Name: "some-org",
				},
				space: cfclient.Space{
					Name: "some-space",
				},
				dir: pwd + "/testdata",
			},
			wantErr: false,
			beforeFunc: func(ctx context.Context) context.Context {
				return config.ContextWithSummary(ctx, report.NewSummary(&bytes.Buffer{}))
			},
			afterFunc: func(ctx context.Context, fakeHolder *fakes.FakeClientHolder, r *fakes.FakeMigratorRegistry, p *fakes.FakeServiceInstanceParser) {
				cfClient := fakeHolder.SourceCFClient().(*cffakes.FakeClient)
				require.Equal(t, 1, cfClient.ListServiceBindingsByQueryCallCount())
				require.Equal(t, 1, cfClient.ListServiceKeysByQueryCallCount())
				require.Equal(t, 1, cfClient.ListSpaceServiceInstancesCallCount())
				require.Equal(t, 1, cfClient.GetServiceInstanceParamsCallCount())
				require.Equal(t, 1, cfClient.GetServicePlanByGUIDCallCount())
				require.Equal(t, 1, cfClient.GetServiceByGuidCallCount())
				require.Equal(t, 1, r.LookupCallCount())
				require.Equal(t, 1, p.MarshalCallCount())
				s, ok := config.SummaryFromContext(ctx)
				require.True(t, ok)
				require.Equal(t, 0, s.ServiceSkippedCount())
				require.Equal(t, 1, s.ServiceSuccessCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.args.ctx
			h := tt.fields.holder
			h.SourceCFClientStub = func() cf.Client {
				return tt.cfClient
			}
			e := migrate.NewServiceInstanceExporter(tt.fields.cfg, tt.fields.holder, tt.fields.registry, tt.fields.parser)

			if tt.beforeFunc != nil {
				ctx = tt.beforeFunc(ctx)
			}
			if err := e.ExportManagedServices(ctx, tt.args.org, tt.args.space, tt.args.om, tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("ExportManagedServices() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.afterFunc != nil {
				tt.afterFunc(ctx, tt.fields.holder, tt.fields.registry, tt.fields.parser)
			}
		})
	}
}

func TestServiceInstanceExporter_ExportUserProvidedServices(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	type fields struct {
		holder   *fakes.FakeClientHolder
		registry *fakes.FakeMigratorRegistry
		parser   *fakes.FakeServiceInstanceParser
	}
	type args struct {
		org   cfclient.Org
		space cfclient.Space
		dir   string
		ctx   context.Context
	}
	tests := []struct {
		name     string
		cfClient *cffakes.FakeClient
		fields   fields
		args     args
		wantErr  bool
	}{
		{
			name: "export a user provided service",
			cfClient: &cffakes.FakeClient{
				ListUserProvidedServiceInstancesByQueryStub: func(url.Values) ([]cfclient.UserProvidedServiceInstance, error) {
					return []cfclient.UserProvidedServiceInstance{{
						Name: "service-provided-instance",
					}}, nil
				},
			},
			fields: fields{
				registry: new(fakes.FakeMigratorRegistry),
				parser:   new(fakes.FakeServiceInstanceParser),
				holder:   new(fakes.FakeClientHolder),
			},
			args: args{
				ctx: context.TODO(),
				org: cfclient.Org{
					Name: "",
				},
				space: cfclient.Space{
					Name: "",
				},
				dir: pwd + "/testdata",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.holder.SourceCFClientStub = func() cf.Client {
				return tt.cfClient
			}
			e := migrate.DefaultServiceInstanceExporter{
				ClientHolder: tt.fields.holder,
				Registry:     tt.fields.registry,
				Parser:       tt.fields.parser,
			}
			if err := e.ExportUserProvidedServices(tt.args.ctx, tt.args.org, tt.args.space, tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("ExportUserProvidedServices() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, 1, tt.cfClient.ListUserProvidedServiceInstancesByQueryCallCount())
		})
	}
}
