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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"strings"
)

type DefaultMigratorRegistry struct {
	migratorFactory Factory
	helper          *MigratorHelper
	cfg             *config.Config
	cfgLoader       config.Loader
	clientHolder    ClientHolder
}

func NewMigratorRegistry(
	factory Factory,
	helper *MigratorHelper,
	cfg *config.Config,
	cfgLoader config.Loader,
	clientHolder ClientHolder,
) *DefaultMigratorRegistry {
	return &DefaultMigratorRegistry{
		migratorFactory: factory,
		helper:          helper,
		cfg:             cfg,
		cfgLoader:       cfgLoader,
		clientHolder:    clientHolder,
	}
}

func (r *DefaultMigratorRegistry) Lookup(org, space string, si *cf.ServiceInstance, om config.OpsManager, dir string, isExport bool) (ServiceInstanceMigrator, bool, error) {
	if !r.shouldMigrate(r.cfg, si.Service) {
		log.Debugf("Skipping %s", si.Service)
		return nil, false, nil
	}

	migrator, err := r.migratorFactory.New(org, space, si, om, r.cfgLoader, r.clientHolder, dir, isExport)
	if err != nil {
		return nil, false, err
	}

	return migrator, true, nil
}

func (r *DefaultMigratorRegistry) shouldMigrate(cfg *config.Config, service string) bool {
	if len(cfg.Services) == 0 {
		return true
	}

	for _, s := range cfg.Services {
		if strings.ToLower(s) == service {
			return true
		}
		if migrator, ok := r.helper.GetMigratorType(service); ok {
			if migrator.String() == strings.ToLower(s) {
				return true
			}
		}
	}

	return false
}
