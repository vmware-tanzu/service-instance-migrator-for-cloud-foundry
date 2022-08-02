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
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	cffakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/fakes"
)

func TestDefaultManifestExporter_ExportAppManifest(t *testing.T) {
	type fields struct {
		config *config.Config
		h      *fakes.FakeClientHolder
	}
	type args struct {
		org   cfclient.Org
		space cfclient.Space
		app   cfclient.App
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       cf.Application
		wantErr    bool
		cfClient   *cffakes.FakeClient
		beforeFunc func(*fakes.FakeClientHolder, *cffakes.FakeClient)
	}{
		{
			name: "exports app manifest when no error occurs",
			fields: fields{
				config: &config.Config{
					DomainsToReplace: map[string]string{
						"apps.cf1.example.com": "apps.cf2.example.com",
					},
				},
				h: new(fakes.FakeClientHolder),
			},
			args: args{
				org: cfclient.Org{
					Name: "some-org",
				},
				space: cfclient.Space{
					Name: "some-space",
				},
				app: cfclient.App{
					Name: "some-app",
				},
			},
			want: cf.Application{
				Name:       "some-app",
				Buildpacks: []string{"java_buildpack"},
				Command:    "",
				DiskQuota:  "0M",
				Docker: struct {
					Image    string `yaml:"image,omitempty"`
					Username string `yaml:"username,omitempty"`
				}{
					Image:    "",
					Username: "",
				},
				Env:                     nil,
				HealthCheckType:         "",
				HealthCheckHTTPEndpoint: "",
				Instances:               0,
				Memory:                  "0M",
				NoRoute:                 false,
				Routes: []struct {
					Route string `yaml:"route,omitempty"`
				}{
					{
						Route: "a-hostname.a-domain.com/some_path",
					},
				},
				Services: []string{"name-1508"},
				Stack:    "cflinuxfs3",
			},
			wantErr: false,
			cfClient: &cffakes.FakeClient{
				NewRequestStub: func(s string, s2 string) *cfclient.Request {
					return &cfclient.Request{}
				},
				DoWithRetryStub: func(f func() error) error {
					return f()
				},
			},
			beforeFunc: func(fakeHolder *fakes.FakeClientHolder, fakeClient *cffakes.FakeClient) {
				fakeClient.DoRequestReturnsOnCall(0, &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(v3TestApp)),
				}, nil)
				fakeClient.DoRequestReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(v3TestAppRoutes)),
				}, nil)
				fakeClient.DoRequestReturnsOnCall(2, &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(v2TestAppServiceBindings)),
				}, nil)
				fakeClient.DoRequestReturnsOnCall(3, &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(v2TestAppServiceInstances)),
				}, nil)
				fakeHolder.SourceCFClientStub = func() cf.Client {
					return fakeClient
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := cc.NewManifestExporter(
				tt.fields.config,
				tt.fields.h,
			)
			tt.beforeFunc(tt.fields.h, tt.cfClient)
			got, err := m.ExportAppManifest(tt.args.org, tt.args.space, tt.args.app)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportAppManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExportAppManifest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var v3TestApp = `{
  "guid": "1cb006ee-fb05-47e1-b541-c34179ddc446",
  "name": "my_app",
  "state": "STOPPED",
  "created_at": "2016-03-17T21:41:30Z",
  "updated_at": "2016-06-08T16:41:26Z",
  "lifecycle": {
    "type": "buildpack",
    "data": {
      "buildpacks": ["java_buildpack"],
      "stack": "cflinuxfs3"
    }
  },
  "relationships": {
    "space": {
      "data": {
        "guid": "2f35885d-0c9d-4423-83ad-fd05066f8576"
      }
    }
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446"
    },
    "space": {
      "href": "https://api.example.org/v3/spaces/2f35885d-0c9d-4423-83ad-fd05066f8576"
    },
    "processes": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/processes"
    },
    "packages": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/packages"
    },
    "environment_variables": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/environment_variables"
    },
    "current_droplet": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/droplets/current"
    },
    "droplets": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/droplets"
    },
    "tasks": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/tasks"
    },
    "start": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/actions/start",
      "method": "POST"
    },
    "stop": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/actions/stop",
      "method": "POST"
    },
    "revisions": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/revisions"
    },
    "deployed_revisions": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/revisions/deployed"
    },
    "features": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/features"
    }
  },
  "metadata": {
    "labels": {},
    "annotations": {}
  }
}
`

var v3TestAppRoutes = `{
  "pagination": {
    "total_results": 3,
    "total_pages": 2,
    "first": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/routes?page=1&per_page=2"
    },
    "last": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/routes?page=2&per_page=2"
    },
    "next": {
      "href": "https://api.example.org/v3/apps/1cb006ee-fb05-47e1-b541-c34179ddc446/routes?page=2&per_page=2"
    },
    "previous": null
  },
  "resources": [
    {
      "guid": "cbad697f-cac1-48f4-9017-ac08f39dfb31",
      "protocol": "http",
      "created_at": "2019-05-10T17:17:48Z",
      "updated_at": "2019-05-10T17:17:48Z",
      "host": "a-hostname",
      "path": "/some_path",
      "url": "a-hostname.a-domain.com/some_path",
      "destinations": [
        {
          "guid": "385bf117-17f5-4689-8c5c-08c6cc821fed",
          "app": {
            "guid": "0a6636b5-7fc4-44d8-8752-0db3e40b35a5",
            "process": {
              "type": "web"
            }
          },
          "weight": null,
          "port": 8080,
          "protocol": "http1"
        },
        {
          "guid": "27e96a3b-5bcf-49ed-8048-351e0be23e6f",
          "app": {
            "guid": "f61e59fa-2121-4217-8c7b-15bfd75baf25",
            "process": {
              "type": "web"
            }
          },
          "weight": null,
          "port": 8080,
          "protocol": "http1"
        }
      ],
      "metadata": {
        "labels": {},
        "annotations": {}
      },
      "relationships": {
        "space": {
          "data": {
            "guid": "885a8cb3-c07b-4856-b448-eeb10bf36236"
          }
        },
        "domain": {
          "data": {
            "guid": "0b5f3633-194c-42d2-9408-972366617e0e"
          }
        }
      },
      "links": {
        "self": {
          "href": "https://api.example.org/v3/routes/cbad697f-cac1-48f4-9017-ac08f39dfb31"
        },
        "space": {
          "href": "https://api.example.org/v3/spaces/885a8cb3-c07b-4856-b448-eeb10bf36236"
        },
        "domain": {
          "href": "https://api.example.org/v3/domains/0b5f3633-194c-42d2-9408-972366617e0e"
        },
        "destinations": {
          "href": "https://api.example.org/v3/routes/cbad697f-cac1-48f4-9017-ac08f39dfb31/destinations"
        }
      }
    }
  ]
}
`

var v2TestAppServiceBindings = `{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "0b6e8fe9-b173-4845-a7aa-e093f1081c94",
        "url": "/v2/service_bindings/0b6e8fe9-b173-4845-a7aa-e093f1081c94",
        "created_at": "2016-06-08T16:41:43Z",
        "updated_at": "2016-06-08T16:41:26Z"
      },
      "entity": {
        "app_guid": "2a3820bb-febd-4c90-ab66-80faa4362142",
        "service_instance_guid": "92f0f510-dbb1-4c04-aa7c-28a8dc0797b4",
        "credentials": {
          "creds-key-72": "creds-val-72"
        },
        "binding_options": {

        },
        "gateway_data": null,
        "gateway_name": "",
        "syslog_drain_url": null,
        "volume_mounts": [

        ],
        "name": "prod-db",
        "last_operation": {
          "type": "create",
          "state": "succeeded",
          "description": "",
          "updated_at": "2018-02-28T16:25:19Z",
          "created_at": "2018-02-28T16:25:19Z"
        },
        "app_url": "/v2/apps/2a3820bb-febd-4c90-ab66-80faa4362142",
        "service_instance_url": "/v2/service_instances/92f0f510-dbb1-4c04-aa7c-28a8dc0797b4",
        "service_binding_parameters_url": "/v2/service_bindings/0b6e8fe9-b173-4845-a7aa-e093f1081c94/parameters"
      }
    }
  ]
}
`

var v2TestAppServiceInstances = `{
  "metadata": {
    "guid": "0d632575-bb06-4ea5-bb19-a451a9644d92",
    "url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92",
    "created_at": "2016-06-08T16:41:29Z",
    "updated_at": "2016-06-08T16:41:26Z"
  },
  "entity": {
    "name": "name-1508",
    "credentials": {
      "creds-key-38": "creds-val-38"
    },
    "service_guid": "a14baddf-1ccc-5299-0152-ab9s49de4422",
    "service_plan_guid": "779d2df0-9cdd-48e8-9781-ea05301cedb1",
    "space_guid": "38511660-89d9-4a6e-a889-c32c7e94f139",
    "gateway_data": null,
    "dashboard_url": null,
    "type": "managed_service_instance",
    "last_operation": {
      "type": "create",
      "state": "succeeded",
      "description": "service broker-provided description",
      "updated_at": "2016-06-08T16:41:29Z",
      "created_at": "2016-06-08T16:41:29Z"
    },
    "tags": [
      "accounting",
      "mongodb"
    ],
    "maintenance_info": {
      "version": "2.1.1",
      "description": "OS image update.\nExpect downtime."
    },
    "space_url": "/v2/spaces/38511660-89d9-4a6e-a889-c32c7e94f139",
    "service_url": "/v2/services/a14baddf-1ccc-5299-0152-ab9s49de4422",
    "service_plan_url": "/v2/service_plans/779d2df0-9cdd-48e8-9781-ea05301cedb1",
    "service_bindings_url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92/service_bindings",
    "service_keys_url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92/service_keys",
    "routes_url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92/routes",
    "shared_from_url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92/shared_from",
    "shared_to_url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92/shared_to",
    "service_instance_parameters_url": "/v2/service_instances/0d632575-bb06-4ea5-bb19-a451a9644d92/parameters"
  }
}
`
