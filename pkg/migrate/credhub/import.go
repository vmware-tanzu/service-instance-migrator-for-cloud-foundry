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

package credhub

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func NewImportSequence(org, space string, instance *cf.ServiceInstance, om config.OpsManager, h ClientHolder, executor exec.Executor) flow.Flow {
	cfHome, err := os.MkdirTemp("", instance.GUID)
	if err != nil {
		panic("failed to create CF_HOME")
	}

	api := h.TargetCFClient().GetClientConfig().ApiAddress

	return flow.ProgressBarSequence(
		fmt.Sprintf("Importing %s", instance.Name),
		flow.StepWithProgressBar(SetCredentials(instance), flow.WithDisplay("Setting credentials")),
		flow.StepWithProgressBar(cf.LoginTargetFoundation(executor, om, api, org, space, cfHome), flow.WithDisplay("Logging into target foundation")),
		flow.StepWithProgressBar(cf.CreateServiceInstance(executor, cfHome, *instance), flow.WithDisplay("Creating service instance")),
	)
}

func SetCredentials(instance *cf.ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, data interface{}, dryRun bool) (flow.Result, error) {
		if len(instance.Credentials) > 0 {
			var domainsToReplace = make(map[string]string)
			if c, ok := config.FromContext(ctx); ok {
				domainsToReplace = c.DomainsToReplace
			}

			creds := make(map[string]interface{})
			for k, v := range instance.Credentials {
				creds[k] = v
				if val, ok := v.(string); ok {
					if newValue, replaced := replaceDomain(val, domainsToReplace); replaced {
						creds[k] = newValue
						log.Debugf("Replaced value %q with %q", val, newValue)
					}
				}
				log.Debugf("Added %q to creds", v)
			}
			instance.Credentials = creds
		}
		return instance, nil
	}
}

func replaceDomain(val string, domainsToReplace map[string]string) (string, bool) {
	for oldDomain, newDomain := range domainsToReplace {
		if strings.Contains(val, oldDomain) {
			return strings.ReplaceAll(val, oldDomain, newDomain), true
		}
	}
	return "", false
}
