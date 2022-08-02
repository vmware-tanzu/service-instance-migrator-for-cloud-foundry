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
	"context"
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func NewImportSequence(org, space string, service Service, instance *cf.ServiceInstance, encryptionKey string, executor exec.Executor, manager config.OpsManager, controller *DatabaseConfig) flow.Flow {
	return flow.ProgressBarSequence(
		fmt.Sprintf("Importing %s", instance.Name),
		flow.StepWithProgressBar(SetCloudControllerDatabaseCredentials(executor, controller, manager), flow.WithDisplay("Setting cc credentials")),
		flow.StepWithProgressBar(Import(org, space, service, instance, encryptionKey), flow.WithDisplay("Creating service instance")),
	)
}

func Import(org, space string, service Service, instance *cf.ServiceInstance, encryptionKey string) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		exists, err := service.ServiceInstanceExists(org, space, instance.Name)
		if err != nil {
			return nil, err
		}

		if exists {
			return nil, fmt.Errorf("service instance name %q already exists", instance.Name)
		}

		if encryptionKey == "" {
			return nil, fmt.Errorf("ccdb encryption key is not set")
		}

		log.Debugf("Creating service instance %q in ccdb...", instance.GUID)
		err = service.Create(org, space, instance, encryptionKey)
		if err != nil {
			return nil, err
		}

		if len(instance.ServiceBindings) > 0 {
			for _, binding := range instance.ServiceBindings {
				if appName, ok := instance.Apps[binding.Guid]; ok {
					log.Debugf("Creating placeholder app %q in ccdb...", appName)
					appGuid, err := service.CreateApp(org, space, appName)
					if err != nil {
						// just log the error and keep going, so we can finish the migration without bindings
						log.Errorf("Could not create service binding: %q for app: %q: %s", binding.Name, appName, err.Error())
						continue
					}
					err = service.CreateServiceBinding(&binding, appGuid, encryptionKey)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		if cfg, ok := config.FromContext(ctx); ok {
			if cfg.IgnoreServiceKeys {
				return instance, nil
			}
		}

		var skErr error
		for _, key := range instance.ServiceKeys {
			err := service.CreateServiceKey(*instance, key)
			if err != nil {
				if skErr != nil {
					skErr = fmt.Errorf("failed to create service key %q, for service instance %q: %w", key, instance.Name, skErr)
				} else {
					skErr = fmt.Errorf("failed to create service key %q, for service instance %q: %w", key, instance.Name, err)
				}
			}
		}

		return instance, skErr
	}
}
