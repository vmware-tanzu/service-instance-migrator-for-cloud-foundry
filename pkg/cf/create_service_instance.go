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

package cf

import (
	"context"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func CreateServiceInstance(e exec.Executor, cfHome string, instance ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		res, err := getServiceInstance(ctx, e, cfHome, instance)
		if err != nil {
			// we expect an error if the instance doesn't exist
			log.Warnf("%s service instance not found", instance.Name)
		}
		if strings.Contains(res.Output, "FAILED") {
			return exec.Result{}, createServiceInstance(ctx, e, cfHome, instance)
		}

		return exec.Result{}, nil
	}
}

func createServiceInstance(ctx context.Context, e exec.Executor, cfHome string, instance ServiceInstance) error {
	log.Infof("Creating service instance")
	createCommand := fmt.Sprintf("CF_HOME='%s' cf create-service '%s' '%s' '%s'", cfHome, instance.Service, instance.Plan, instance.Name)
	if len(instance.Credentials) > 0 {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		credentialsBytes, err := json.Marshal(instance.Credentials)
		if err != nil {
			return fmt.Errorf("error encoding credentials: %w", err)
		}
		createCommand = fmt.Sprintf("CF_HOME='%s' cf create-service '%s' '%s' '%s' -c '%s'", cfHome, instance.Service, instance.Plan, instance.Name, strings.Trim(string(credentialsBytes), "\n"))
	}
	log.Debugf("Create service command: %q", createCommand)
	_, err := e.Execute(ctx, strings.NewReader(createCommand))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create service instance %q", instance.Name))
	}

	return nil
}
