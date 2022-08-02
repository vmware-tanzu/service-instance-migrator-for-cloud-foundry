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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
)

type ExportMigratorFactory struct {
	Config       *config.Config
	ClientHolder migrate.ClientHolder
}

type ImportMigratorFactory struct {
	Config       *config.Config
	ClientHolder migrate.ClientHolder
}

func NewExportMigratorFactory(cfg *config.Config, h migrate.ClientHolder) ExportMigratorFactory {
	return ExportMigratorFactory{
		Config:       cfg,
		ClientHolder: h,
	}
}

func NewImportMigratorFactory(cfg *config.Config, l migrate.ClientHolder) ImportMigratorFactory {
	return ImportMigratorFactory{
		Config:       cfg,
		ClientHolder: l,
	}
}

func (f ExportMigratorFactory) NewOrgExporter(exporter migrate.ServiceInstanceExporter) OrgExporter {
	return migrate.NewOrgExporter(
		exporter,
		f.ClientHolder,
		f.Config.IncludedOrgs,
		f.Config.ExcludedOrgs,
	)
}

func (f ExportMigratorFactory) NewSpaceExporter(exporter migrate.ServiceInstanceExporter) SpaceExporter {
	return migrate.NewSpaceExporter(
		exporter,
		f.ClientHolder,
	)
}

func (f ImportMigratorFactory) NewOrgImporter(importer migrate.ServiceInstanceImporter) OrgImporter {
	return migrate.NewOrgImporter(
		migrate.NewSpaceImporter(importer),
		f.Config.IncludedOrgs,
		f.Config.ExcludedOrgs,
	)
}

func (f ImportMigratorFactory) NewSpaceImporter(importer migrate.ServiceInstanceImporter) SpaceImporter {
	return migrate.NewSpaceImporter(importer)
}
