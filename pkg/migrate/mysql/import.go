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

package mysql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func NewImportSequence(api, org, space string, instance *cf.ServiceInstance, om config.OpsManager, executor exec.Executor) flow.Flow {
	cfHome, err := os.MkdirTemp("", instance.GUID)
	if err != nil {
		panic("failed to create CF_HOME")
	}

	log.Debugf("Creating import with CF_HOME='%s' for %s %s service running in %s/%s", cfHome, instance.Name, instance.Service, org, space)

	return flow.ProgressBarSequence(
		fmt.Sprintf("Importing %s", instance.Name),
		flow.StepWithProgressBar(cf.LoginTargetFoundation(executor, om, api, org, space, cfHome), flow.WithDisplay("Logging into target foundation")),
		flow.StepWithProgressBar(cf.CreateServiceInstance(executor, cfHome, *instance), flow.WithDisplay("Creating service instance")),
		flow.StepWithProgressBar(cf.GetServiceInstance(executor, cfHome, instance, 15*time.Minute, 10*time.Second), flow.WithDisplay("Waiting for service instance to create")),
		flow.StepWithProgressBar(TransferBackup(executor, om, instance), flow.WithDisplay("Transferring backup")),
		flow.StepWithProgressBar(RestoreBackup(executor, om, instance), flow.WithDisplay("Restoring from backup")),
	)
}

func TransferBackup(e exec.Executor, om config.OpsManager, instance *cf.ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		log.Infof("Transferring backup")
		if _, err := os.Stat(instance.BackupFile); os.IsNotExist(err) {
			return exec.Result{}, fmt.Errorf("failed to transfer backup: %q, file does not exist", instance.BackupFile)
		}
		log.Debugf("Transferring %q to %q", instance.BackupFile, fmt.Sprintf("service-instance_%s", instance.GUID))
		return bosh.Run(e, ctx, om, "-d", fmt.Sprintf("service-instance_%s", instance.GUID), "scp",
			fmt.Sprintf("%s mysql/0:/tmp", instance.BackupFile))
	}
}

func RestoreBackup(e exec.Executor, om config.OpsManager, instance *cf.ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		log.Infof("Restoring backup")
		log.Debugf("Restoring %q on %q with %q encryption key", instance.BackupFile, fmt.Sprintf("service-instance_%s", instance.GUID), instance.BackupEncryptionKey)
		_, file := filepath.Split(instance.BackupFile)
		result, err := bosh.Run(e, ctx, om, "-d", fmt.Sprintf("service-instance_%s", instance.GUID), "ssh", "mysql/0", "-c",
			fmt.Sprintf("\"sudo mysql-restore --encryption-key %s --restore-file %s\"", instance.BackupEncryptionKey, filepath.Join("/tmp", file)))
		if err != nil {
			if strings.Contains(result.Status.Output, "Restore is permitted only in a non-empty service instance") {
				log.Warnln("failed to restore backup, restore is permitted only in a empty service instance")
				return result, nil
			}
			return result, err
		}

		return result, err
	}
}
