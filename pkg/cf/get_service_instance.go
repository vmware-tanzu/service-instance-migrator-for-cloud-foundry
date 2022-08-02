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
	"time"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func GetServiceInstance(e exec.Executor, cfHome string, instance *ServiceInstance, duration time.Duration, pause time.Duration) flow.StepFunc {
	return func(ctx context.Context, config interface{}, dryRun bool) (flow.Result, error) {
		log.Infof("Waiting for '%s' service instance to become ready", instance.Name)
		res, err := executeForDuration(ctx, e, []string{
			fmt.Sprintf("CF_HOME='%s' cf service '%s' | grep -i 'status:' | awk '{print $NF}'", cfHome, instance.Name),
		}, duration, pause, func(result exec.Result) bool {
			return strings.Contains(result.Output, "succeeded")
		}, func(result exec.Result) error {
			if strings.Contains(result.Output, "failed") {
				return fmt.Errorf("'%s' is in a failed state", instance.Name)
			}
			return nil
		})
		if err != nil {
			return exec.Result{}, fmt.Errorf("failed to check service instance status: %w", err)
		}

		res, err = getServiceInstance(ctx, e, cfHome, *instance)
		if err != nil {
			return exec.Result{}, fmt.Errorf("failed to get service instance guid for '%s': %w", instance.Name, err)
		}

		instance.GUID = strings.TrimSuffix(res.Output, "\n")
		log.Debugf("Service instance %q created", instance.GUID)

		return exec.Result{}, nil
	}
}

func getServiceInstance(ctx context.Context, e exec.Executor, cfHome string, instance ServiceInstance) (exec.Result, error) {
	lines := []string{
		fmt.Sprintf("CF_HOME='%s' cf service '%s' --guid", cfHome, instance.Name),
	}
	return e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
}

func executeForDuration(ctx context.Context, s exec.Executor, lines []string, timeout time.Duration, pause time.Duration, doneCondition func(exec.Result) bool, errorCondition func(exec.Result) error) (exec.Result, error) {
	done := make(chan bool)
	var err error
	var res exec.Result
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			res, err = s.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
			if err != nil {
				cancel()
				return
			}

			if err = errorCondition(res); err != nil {
				cancel()
				return
			}

			if doneCondition(res) || res.DryRun {
				close(done)
				continue
			}

			time.Sleep(pause)
		}
	}()
	select {
	case <-done:
		return res, nil
	case <-child.Done():
		log.Debugln("context was cancelled")
		return res, err
	case <-time.After(timeout):
		close(done)
		return res, errors.New("timed out waiting for backup")
	}
}
