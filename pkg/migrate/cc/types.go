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
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . CloudControllerServiceFactory

type CloudControllerServiceFactory interface {
	NewCloudControllerService(cfg *Config, isExport bool) (Service, error)
}

//counterfeiter:generate -o fakes . CloudControllerRepositoryFactory

type CloudControllerRepositoryFactory interface {
	NewCCDB(bool) (db.Repository, error)
}

//counterfeiter:generate -o fakes . Service

type Service interface {
	Create(org, space string, instance *cf.ServiceInstance, encryptionKey string) error
	Delete(org, space string, instance *cf.ServiceInstance) error
	ServiceInstanceExists(org, space, name string) (bool, error)
	CreateServiceKey(si cf.ServiceInstance, key cf.ServiceKey) error
	CreateApp(org, space, name string) (string, error)
	CreateServiceBinding(binding *cf.ServiceBinding, appGUID string, encryptionKey string) error
	FindAppByGUID(guid string) (string, error)
	DownloadManifest(org, space, appName string) (cf.Application, error)
}

//counterfeiter:generate -o fakes . ManifestExporter

type ManifestExporter interface {
	ExportAppManifest(org cfclient.Org, space cfclient.Space, app cfclient.App) (cf.Application, error)
}

//counterfeiter:generate -o fakes . ClientHolder

type ClientHolder interface {
	SourceCFClient() cf.Client
	CFClient(toSource bool) cf.Client
}
