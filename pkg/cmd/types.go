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

package cmd

import (
	"context"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . OrgImporter

type OrgImporter interface {
	Import(ctx context.Context, om config.OpsManager, dir string, org ...string) error
	ImportAll(ctx context.Context, om config.OpsManager, dir string) error
}

//counterfeiter:generate -o fakes . OrgExporter

type OrgExporter interface {
	Export(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error
	ExportAll(ctx context.Context, om config.OpsManager, dir string) error
}

//counterfeiter:generate -o fakes . SpaceImporter

type SpaceImporter interface {
	Import(ctx context.Context, om config.OpsManager, dir string, org, space string) error
}

//counterfeiter:generate -o fakes . SpaceExporter

type SpaceExporter interface {
	Export(ctx context.Context, om config.OpsManager, dir string, org, space string) error
}

//counterfeiter:generate -o fakes . ExporterFactory

type ExporterFactory interface {
	NewOrgExporter(exporter migrate.ServiceInstanceExporter) OrgExporter
	NewSpaceExporter(exporter migrate.ServiceInstanceExporter) SpaceExporter
}

//counterfeiter:generate -o fakes . ImporterFactory

type ImporterFactory interface {
	NewOrgImporter(importer migrate.ServiceInstanceImporter) OrgImporter
	NewSpaceImporter(importer migrate.ServiceInstanceImporter) SpaceImporter
}

type NoopPropertiesProvider struct{}
type NoopClientFactory struct{}
type NoopBoshPropertiesBuilder struct{}
type NoopCFPropertiesBuilder struct{}
type NoopCCDBPropertiesBuilder struct{}

func (s NoopPropertiesProvider) Environment(config.BoshPropertiesBuilder, config.CFPropertiesBuilder, config.CCDBPropertiesBuilder) config.EnvProperties {
	return config.EnvProperties{}
}
func (s NoopPropertiesProvider) SourceBoshPropertiesBuilder() config.BoshPropertiesBuilder {
	return NoopBoshPropertiesBuilder{}
}
func (s NoopPropertiesProvider) TargetBoshPropertiesBuilder() config.BoshPropertiesBuilder {
	return NoopBoshPropertiesBuilder{}
}
func (s NoopPropertiesProvider) SourceCFPropertiesBuilder() config.CFPropertiesBuilder {
	return NoopCFPropertiesBuilder{}
}
func (s NoopPropertiesProvider) TargetCFPropertiesBuilder() config.CFPropertiesBuilder {
	return NoopCFPropertiesBuilder{}
}
func (s NoopPropertiesProvider) SourceCCDBPropertiesBuilder(b config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
	return NoopCCDBPropertiesBuilder{}
}
func (s NoopPropertiesProvider) TargetCCDBPropertiesBuilder(b config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
	return NoopCCDBPropertiesBuilder{}
}

func (f NoopClientFactory) CFClient(bool) cf.Client       { return &cf.ClientImpl{} }
func (f NoopClientFactory) SourceCFClient() cf.Client     { return &cf.ClientImpl{} }
func (f NoopClientFactory) SourceOpsManClient() om.Client { return &om.ClientImpl{} }
func (f NoopClientFactory) SourceBoshClient() bosh.Client { return &bosh.ClientImpl{} }
func (f NoopClientFactory) TargetCFClient() cf.Client     { return &cf.ClientImpl{} }
func (f NoopClientFactory) TargetOpsManClient() om.Client { return &om.ClientImpl{} }
func (f NoopClientFactory) TargetBoshClient() bosh.Client { return &bosh.ClientImpl{} }

func (b NoopBoshPropertiesBuilder) Build() *config.BoshProperties { return &config.BoshProperties{} }
func (b NoopCFPropertiesBuilder) Build() *config.CFProperties     { return &config.CFProperties{} }
func (b NoopCCDBPropertiesBuilder) Build() *config.CCDBProperties { return &config.CCDBProperties{} }
