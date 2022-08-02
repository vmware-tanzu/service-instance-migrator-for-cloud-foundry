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
	"fmt"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/validation"
)

type ManagedServiceInstanceImporter struct {
	Registry MigratorRegistry
}

func NewServiceInstanceImporter(registry MigratorRegistry) ManagedServiceInstanceImporter {
	return ManagedServiceInstanceImporter{
		Registry: registry,
	}
}

func (i ManagedServiceInstanceImporter) ImportManagedService(ctx context.Context, org, space string, si *cf.ServiceInstance, om config.OpsManager, dir string) error {
	migrator, migrate, err := i.Registry.Lookup(org, space, si, om, dir, false)
	if err != nil {
		return fmt.Errorf("failed to find a valid migrator for instance %s: %w", si.Name, err)
	}

	dryRun := false
	if cfg, ok := config.FromContext(ctx); ok {
		dryRun = cfg.DryRun
	}

	if !migrate || dryRun {
		if summary, ok := config.SummaryFromContext(ctx); ok {
			summary.AddSkippedService(org, space, si.Name, si.Service, nil)
		}
		return nil
	}

	if migrator == nil {
		if summary, ok := config.SummaryFromContext(ctx); ok {
			summary.AddSkippedService(org, space, si.Name, si.Service, nil)
		}
		return nil
	}

	err = migrator.Validate(si, false)
	if err != nil {
		return err
	}

	log.Infof("Importing %q to %s/%s", si.Name, org, space)
	_, err = migrator.Migrate(ctx)
	if err != nil {
		var validationErr *validation.MigrationError
		if errors.As(err, &validationErr) && len(si.ServiceBindings) == 0 {
			log.Warnf("validation error %s, skipped migrating service instance %s", validationErr.Error(), si.Name)
			if summary, ok := config.SummaryFromContext(ctx); ok {
				summary.AddSkippedService(org, space, si.Name, si.Service, validationErr)
			}
			return nil
		}
		log.Errorf("error migrating service instance %s", si.Name)
		if summary, ok := config.SummaryFromContext(ctx); ok {
			summary.AddFailedService(org, space, si.Name, si.Service, err)
		}
		return errors.Wrap(err, fmt.Sprintf("failed to migrate %s", si.Name))
	}

	log.Debugf("Finished importing %q", si.Name)

	if summary, ok := config.SummaryFromContext(ctx); ok {
		summary.AddSuccessfulService(org, space, si.Name, si.Service)
	}

	return nil
}
