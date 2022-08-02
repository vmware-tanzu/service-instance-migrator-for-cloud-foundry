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

package cf

type ServiceInstance struct {
	Name                string                 `yaml:"name,omitempty"`
	GUID                string                 `yaml:"guid,omitempty"`
	Type                string                 `yaml:"type,omitempty"`
	Tags                string                 `yaml:"tags,omitempty"`
	Params              map[string]interface{} `yaml:"params,omitempty"`
	RouteServiceUrl     string                 `yaml:"route_service_url,omitempty"`
	SyslogDrainUrl      string                 `yaml:"syslog_drain_url,omitempty"`
	DashboardURL        string                 `yaml:"dashboard_url,omitempty"`
	Service             string                 `yaml:"service,omitempty"`
	Plan                string                 `yaml:"plan,omitempty"`
	Credentials         map[string]interface{} `json:"credentials,omitempty"`
	ServiceBindings     []ServiceBinding       `yaml:"service_bindings,omitempty"`
	ServiceKeys         []ServiceKey           `yaml:"service_keys,omitempty"`
	BackupID            string                 `yaml:"backup_id,omitempty"`
	BackupDate          string                 `yaml:"backup_date,omitempty"`
	BackupTime          string                 `yaml:"backup_time,omitempty"`
	BackupFile          string                 `yaml:"backup_file,omitempty"`
	BackupEncryptionKey string                 `yaml:"backup_encryption_key,omitempty"`
	Apps                map[string]string      `yaml:"apps,omitempty"`
	AppManifest         Manifest               `yaml:"app_manifest"`
}

type ServiceBinding struct {
	Guid                string                 `yaml:"guid,omitempty"`
	Name                string                 `yaml:"name,omitempty"`
	AppGuid             string                 `yaml:"app_guid,omitempty"`
	ServiceInstanceGuid string                 `yaml:"service_instance_guid,omitempty"`
	Credentials         map[string]interface{} `yaml:"credentials,omitempty"`
	BindingOptions      interface{}            `yaml:"binding_options,omitempty"`
	GatewayData         interface{}            `yaml:"gateway_data,omitempty"`
	GatewayName         string                 `yaml:"gateway_name,omitempty"`
	SyslogDrainUrl      string                 `yaml:"syslog_drain_url,omitempty"`
	VolumeMounts        interface{}            `yaml:"volume_mounts,omitempty"`
	AppUrl              string                 `yaml:"app_url,omitempty"`
	ServiceInstanceUrl  string                 `yaml:"service_instance_url,omitempty"`
}

type ServiceKey struct {
	Name                string                 `yaml:"name"`
	Guid                string                 `yaml:"guid"`
	ServiceInstanceGuid string                 `yaml:"service_instance_guid"`
	Credentials         map[string]interface{} `yaml:"credentials"`
	ServiceInstanceUrl  string                 `yaml:"service_instance_url"`
}

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	Name       string   `yaml:"name"`
	Buildpacks []string `yaml:"buildpacks"`
	Command    string   `yaml:"command"`
	DiskQuota  string   `yaml:"disk_quota"`
	Docker     struct {
		Image    string `yaml:"image,omitempty"`
		Username string `yaml:"username,omitempty"`
	} `yaml:"docker,omitempty"`
	Env                     map[string]interface{} `yaml:"env"`
	HealthCheckType         string                 `yaml:"health-check-type"`
	HealthCheckHTTPEndpoint string                 `yaml:"health-check-http-endpoint,omitempty"`
	Instances               int64                  `yaml:"instances"`
	Memory                  string                 `yaml:"memory"`
	NoRoute                 bool                   `yaml:"no-route,omitempty"`
	Routes                  []struct {
		Route string `yaml:"route,omitempty"`
	} `yaml:"routes,omitempty"`
	Services []string `yaml:"services"`
	Stack    string   `yaml:"stack"`
	Timeout  int64    `yaml:"timeout,omitempty"`
}
