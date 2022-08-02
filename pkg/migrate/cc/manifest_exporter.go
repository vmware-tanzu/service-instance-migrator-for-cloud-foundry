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

package cc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type DefaultManifestExporter struct {
	config *config.Config
	holder ClientHolder
}

func NewManifestExporter(cfg *config.Config, h ClientHolder) *DefaultManifestExporter {
	return &DefaultManifestExporter{
		config: cfg,
		holder: h,
	}
}

func (m *DefaultManifestExporter) ExportAppManifest(org cfclient.Org, space cfclient.Space, app cfclient.App) (cf.Application, error) {
	log.Infof("Writing app manifest for %s/%s/%s", org.Name, space.Name, app.Name)

	manifestApp := cf.Application{}
	manifestApp.Name = app.Name
	manifestApp.Env = app.Environment
	manifestApp.HealthCheckType = app.HealthCheckType
	manifestApp.HealthCheckHTTPEndpoint = app.HealthCheckHttpEndpoint
	manifestApp.Instances = int64(app.Instances)
	manifestApp.Command = app.Command
	manifestApp.Memory = getSizeString(int64(app.Memory))
	manifestApp.DiskQuota = getSizeString(int64(app.DiskQuota))
	manifestApp.Timeout = int64(app.HealthCheckTimeout)

	var (
		res *http.Response
		err error
	)
	client := m.holder.SourceCFClient()

	err = client.DoWithRetry(func() error {
		req := client.NewRequest(http.MethodGet, fmt.Sprintf("/v3/apps/%s", app.Guid))
		res, err = client.DoRequest(req)
		if err == nil {
			if res.StatusCode >= 500 && res.StatusCode <= 599 {
				if res.Body != nil {
					_ = res.Body.Close()
				}
				return cf.ErrRetry
			}
		}
		return err
	})
	if err != nil {
		cfErr := &cfclient.CloudFoundryHTTPError{}
		if errors.As(err, cfErr) {
			if cfErr.StatusCode == http.StatusNotFound {
				return cf.Application{}, nil
			}
		}
		return cf.Application{}, err
	}

	var v3app cfclient.V3App
	if err = json.NewDecoder(res.Body).Decode(&v3app); err != nil {
		if res.Body != nil {
			_ = res.Body.Close()
		}
		return cf.Application{}, err
	}
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}

	manifestApp.Buildpacks = v3app.Lifecycle.BuildpackData.Buildpacks
	if v3app.Lifecycle.Type == "docker" {
		manifestApp.Buildpacks = nil
		manifestApp.Docker = struct {
			Image    string "yaml:\"image,omitempty\""
			Username string "yaml:\"username,omitempty\""
		}{
			Image:    app.DockerImage,
			Username: app.DockerCredentials.Username,
		}
	}

	manifestApp.Stack = v3app.Lifecycle.BuildpackData.Stack

	err = client.DoWithRetry(func() error {
		var err error
		req := client.NewRequest(http.MethodGet, fmt.Sprintf("/v3/apps/%s/routes", app.Guid))
		res, err = client.DoRequest(req)
		if err == nil {
			if res.StatusCode >= 500 && res.StatusCode <= 599 {
				if res.Body != nil {
					_ = res.Body.Close()
				}
				return cf.ErrRetry
			}
		}
		return err
	})
	if err != nil {
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
		return cf.Application{}, err
	}

	var routesResponse = struct {
		Resources []struct {
			URL string `json:"url"`
		} `json:"resources"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&routesResponse); err != nil {
		if res.Body != nil {
			_ = res.Body.Close()
		}
		return cf.Application{}, err
	}
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}

	manifestApp.NoRoute = len(routesResponse.Resources) == 0
	manifestApp.Routes = make([]struct {
		Route string `yaml:"route,omitempty"`
	}, 0)

	var routes []string
	for _, route := range routesResponse.Resources {
		routes = append(routes, route.URL)
	}
	routeMapper := &routeMapper{
		DomainsToReplace: m.config.DomainsToReplace,
	}
	adjustedRoutes := routeMapper.AdjustRoutes(routes)
	for _, adjustedRoute := range adjustedRoutes {
		manifestApp.Routes = append(manifestApp.Routes, struct {
			Route string `yaml:"route,omitempty"`
		}{adjustedRoute})
	}

	err = client.DoWithRetry(func() error {
		req := client.NewRequest(http.MethodGet, fmt.Sprintf("/v2/apps/%s/service_bindings", app.Guid))
		res, err = client.DoRequest(req)
		if err == nil {
			if res.StatusCode >= 500 && res.StatusCode <= 599 {
				if res.Body != nil {
					_ = res.Body.Close()
				}
				return cf.ErrRetry
			}
		}
		return err
	})
	if err != nil {
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
		return cf.Application{}, err
	}

	var bindings cfclient.ServiceBindingsResponse
	if err = json.NewDecoder(res.Body).Decode(&bindings); err != nil {
		if res.Body != nil {
			_ = res.Body.Close()
		}
		return cf.Application{}, err
	}
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}

	manifestApp.Services = make([]string, 0, len(bindings.Resources))
	for _, binding := range bindings.Resources {
		var res *http.Response
		err := client.DoWithRetry(func() error {
			req := client.NewRequest(http.MethodGet, binding.Entity.ServiceInstanceUrl)
			res, err = client.DoRequest(req)
			if err == nil {
				if res.StatusCode >= 500 && res.StatusCode <= 599 {
					if res.Body != nil {
						_ = res.Body.Close()
					}
					return cf.ErrRetry
				}
			}
			return err
		})
		if err != nil {
			if res != nil && res.Body != nil {
				_ = res.Body.Close()
			}
			return cf.Application{}, err
		}

		var si cfclient.ServiceInstanceResource
		if err = json.NewDecoder(res.Body).Decode(&si); err != nil {
			if res.Body != nil {
				_ = res.Body.Close()
			}
			return cf.Application{}, err
		}
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}

		manifestApp.Services = append(manifestApp.Services, si.Entity.Name)
	}

	return manifestApp, nil
}

func getSizeString(size int64) string {
	suffix := "M"
	if size >= 1024 {
		size = size / 1024
		suffix = "G"
	}

	return fmt.Sprintf("%d%s", size, suffix)
}
