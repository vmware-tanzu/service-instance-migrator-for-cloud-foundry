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
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/s3"
)

type MigratorFactory struct {
	l  config.Loader
	h  ClientHolder
	mh *MigratorHelper
	e  exec.Executor
	sf cc.CloudControllerServiceFactory
}

func NewMigratorFactory(l config.Loader, h ClientHolder, mh *MigratorHelper, e exec.Executor, sf cc.CloudControllerServiceFactory) *MigratorFactory {
	return &MigratorFactory{
		l:  l,
		h:  h,
		mh: mh,
		e:  e,
		sf: sf,
	}
}

func (f *MigratorFactory) New(org, space string, si *cf.ServiceInstance, om config.OpsManager, l config.Loader, h ClientHolder, dir string, isExport bool) (ServiceInstanceMigrator, error) {
	cfg, err := f.mh.GetReader().GetMigration()
	if err != nil {
		return nil, err
	}

	switch ServiceType(si.Type) {
	case ManagedService:
		switch Service(si.Service) {
		case MySQLService:
			mySQLFlow, err := f.buildMySQLFlow(org, space, si, om, dir, isExport)
			return mysql.NewMigrator(mySQLFlow, f.mh.GetReader()), err
		case CredHubService:
			credhubFlow, err := f.buildCredhubFlow(org, space, si, om, h, isExport)
			return credhub.NewMigrator(credhubFlow, f.mh.GetReader()), err
		case CustomSQLServerService, SQLServerService, ECSBucketService:
			ccFlow, err := f.buildCCFlow(org, space, si, om, l, isExport)
			return cc.NewMigrator(ccFlow), err
		default:
			if cfg.UseDefaultMigrator {
				serviceFlow, err := f.buildManagedServiceFlow(org, space, si, isExport)
				return NewManagedServiceMigrator(serviceFlow), err
			}
		}
	case UserProvidedService:
		serviceFlow, err := f.buildUserProvidedServiceFlow(org, space, si, isExport)
		return NewUserProvidedServiceMigrator(serviceFlow), err
	}

	log.Warnf("Service instance %q is not supported, service: %q, type: %q", si.Name, si.Service, si.Type)

	return nil, nil
}

func (f *MigratorFactory) buildMySQLFlow(org, space string, si *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (flow.Flow, error) {
	d, err := s3.NewDownloader(f.mh.GetReader())
	if err != nil {
		return nil, err
	}

	api := f.h.CFClient(isExport).GetClientConfig().ApiAddress

	var sequence flow.Flow
	if isExport {
		sequence = mysql.NewExportSequence(api, org, space, si, om, d, f.e, dir)
	} else {
		sequence = mysql.NewImportSequence(api, org, space, si, om, f.e)
	}
	return sequence, nil
}

func (f *MigratorFactory) buildCredhubFlow(org, space string, si *cf.ServiceInstance, om config.OpsManager, h ClientHolder, isExport bool) (flow.Flow, error) {
	var sequence flow.Flow
	if isExport {
		sequence = credhub.NewExportSequence(org, space, si, om, h, f.e)
	} else {
		sequence = credhub.NewImportSequence(org, space, si, om, h, f.e)
	}

	return sequence, nil
}

func (f *MigratorFactory) buildCCFlow(org, space string, si *cf.ServiceInstance, om config.OpsManager, l config.Loader, isExport bool) (flow.Flow, error) {
	m, _ := f.mh.GetMigratorType(si.Service)
	ccConfig := l.CCDBConfig(m.String(), isExport).(*cc.Config)
	if ccConfig == nil {
		return nil, fmt.Errorf("failed to find ccdb config for %s", si.Service)
	}

	if err := ccConfig.Validate(isExport); err != nil {
		return nil, fmt.Errorf("migration config validation failed for %s, %w", m, err)
	}

	svc, err := f.sf.NewCloudControllerService(ccConfig, isExport)
	if err != nil {
		return nil, err
	}

	var sequence flow.Flow
	if isExport {
		sequence = cc.NewExportSequence(org, space, svc, si, f.e, om, &ccConfig.SourceCloudControllerDatabase)
	} else {
		sequence = cc.NewImportSequence(org, space, svc, si, ccConfig.TargetCloudControllerDatabase.EncryptionKey, f.e, om, &ccConfig.TargetCloudControllerDatabase)
	}

	return sequence, nil
}

func (f *MigratorFactory) buildManagedServiceFlow(org, space string, si *cf.ServiceInstance, isExport bool) (flow.Flow, error) {
	return NewManagedServiceFlow(org, space, f.h, si, isExport), nil
}

func (f *MigratorFactory) buildUserProvidedServiceFlow(org, space string, si *cf.ServiceInstance, isExport bool) (flow.Flow, error) {
	return NewUserProvidedServiceFlow(org, space, f.h, si, isExport), nil
}
