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

package migrate

import (
	"context"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . ServiceInstanceExporter

type ServiceInstanceExporter interface {
	ExportManagedServices(ctx context.Context, org cfclient.Org, space cfclient.Space, om config.OpsManager, dir string) error
	ExportUserProvidedServices(ctx context.Context, org cfclient.Org, space cfclient.Space, dir string) error
}

//counterfeiter:generate -o fakes . ServiceInstanceImporter

type ServiceInstanceImporter interface {
	ImportManagedService(ctx context.Context, org string, space string, instance *cf.ServiceInstance, om config.OpsManager, dir string) error
}

//counterfeiter:generate -o fakes . ServiceInstanceParser

type ServiceInstanceParser interface {
	Unmarshal(out interface{}, fd io.FileDescriptor) error
	Marshal(in interface{}, fd io.FileDescriptor) error
}

//counterfeiter:generate -o fakes . MigratorRegistry

type MigratorRegistry interface {
	Lookup(org, space string, si *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (ServiceInstanceMigrator, bool, error)
}

//counterfeiter:generate -o fakes . Validator

type Validator interface {
	Validate(*cf.ServiceInstance, bool) error
}

//counterfeiter:generate -o fakes . ServiceInstanceMigrator

type ServiceInstanceMigrator interface {
	Migrate(ctx context.Context) (*cf.ServiceInstance, error)
	Validator
}

//counterfeiter:generate -o fakes . ClientHolder

type ClientHolder interface {
	SourceBoshClient() bosh.Client
	TargetBoshClient() bosh.Client
	SourceOpsManClient() om.Client
	TargetOpsManClient() om.Client
	SourceCFClient() cf.Client
	TargetCFClient() cf.Client
	CFClient(toSource bool) cf.Client
}

//counterfeiter:generate -o fakes . Factory

type Factory interface {
	New(org, space string, si *cf.ServiceInstance, om config.OpsManager, l config.Loader, h ClientHolder, dir string, isExport bool) (ServiceInstanceMigrator, error)
}
