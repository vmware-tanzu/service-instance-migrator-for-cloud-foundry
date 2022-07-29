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

func NewExportSequence(org, space string, service Service, instance *cf.ServiceInstance, executor exec.Executor, manager config.OpsManager, controller *DatabaseConfig) flow.Flow {
	return flow.ProgressBarSequence(
		fmt.Sprintf("Exporting %s", instance.Name),
		flow.StepWithProgressBar(
			SetCloudControllerDatabaseCredentials(executor, controller, manager),
			flow.WithDisplay("Setting cc credentials"),
		),
		flow.StepWithProgressBar(
			Export(org, space, service, instance),
			flow.WithDisplay("Removing service instance"),
		),
	)
}

func Export(org, space string, service Service, instance *cf.ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		instance.Apps = make(map[string]string)
		for _, binding := range instance.ServiceBindings {
			if len(binding.AppGuid) > 0 {
				log.Debugf("Searching for app by guid %q in ccdb...", binding.AppGuid)
				appName, err := service.FindAppByGUID(binding.AppGuid)
				if err != nil {
					return instance, err
				}
				instance.Apps[binding.Guid] = appName
				manifest, err := service.DownloadManifest(org, space, appName)
				if err != nil {
					return instance, err
				}
				instance.AppManifest.Applications = append(instance.AppManifest.Applications, manifest)
			}
		}

		log.Debugf("Deleting service instance %q in ccdb...", instance.GUID)
		err := service.Delete(org, space, instance)
		if err != nil {
			return instance, err
		}

		return instance, err
	}
}
