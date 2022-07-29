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
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func SetCloudControllerDatabaseCredentials(e exec.Executor, cfg *DatabaseConfig, om config.OpsManager) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		var (
			deploymentName string
			err            error
		)

		if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" || cfg.EncryptionKey == "" {
			log.Debugf("Fetching ccdb creds for %s", om.Hostname)
			deploymentName, err = findDeploymentName(ctx, e, om, "^cf-")
			if err != nil {
				return exec.Result{}, err
			}
		}

		if cfg.Host == "" {
			databaseInstance, err := findInstanceName(ctx, e, om, deploymentName)
			if err != nil {
				return exec.Result{}, err
			}

			log.Debugf("Fetching ccdb ip address for %s", databaseInstance)
			ipAddress, err := findInstanceIPAddress(ctx, e, om, deploymentName, databaseInstance)
			if err != nil {
				return exec.Result{}, err
			}

			log.Debugf("ccdb ipAddress='%s'", ipAddress)
			cfg.Host = ipAddress
		}

		if cfg.Username == "" || cfg.Password == "" {
			username, password, err := getCredentials(ctx, e, om, deploymentName)
			if err != nil {
				return exec.Result{}, err
			}

			log.Debugf("ccdb username='%s'", username)
			log.Debugf("ccdb password='%s'", password)
			cfg.Username = username
			cfg.Password = password
		}

		if cfg.EncryptionKey == "" {
			encryptionKey, err := getEncryptionKey(ctx, e, om, deploymentName)
			if err != nil {
				return exec.Result{}, err
			}

			log.Debugf("ccdb encryptionKey='%s'", encryptionKey)
			cfg.EncryptionKey = encryptionKey
		}

		return exec.Result{}, nil
	}
}

func findInstanceName(ctx context.Context, e exec.Executor, om config.OpsManager, deployment string) (string, error) {
	res, err := bosh.Run(e, ctx, om, "-d", deployment, "--column=instance", "--column=process", "is", "-p", "|", "awk", "-F' '", "'$2==\"proxy\"'", "|", "awk", "'{print $1}'", "|", "head", "-1", "|", "tr", "-d", "'\\t\\n'")
	if err != nil {
		return "", errors.Wrap(err, "failed to bosh instance running credhub")
	}

	return strings.TrimSuffix(res.Status.Output, "\t\n"), nil
}

func findInstanceIPAddress(ctx context.Context, e exec.Executor, om config.OpsManager, deployment string, instance string) (string, error) {
	res, err := bosh.Run(e, ctx, om, "-d", deployment, "--column=instance", "--column=ips", "is", "--dns", "|", "grep", instance, "|", "awk", "'{print $2}'", "|", "tr", "-d", "'\\t\\n'")
	if err != nil {
		return "", errors.Wrap(err, "failed to bosh instance running credhub")
	}

	return strings.TrimSuffix(res.Status.Output, "\t\n"), nil
}

func findDeploymentName(ctx context.Context, e exec.Executor, om config.OpsManager, pattern string) (string, error) {
	res, err := bosh.Run(e, ctx, om, "deps", "--column=name", "|", "grep", "'"+pattern+"'", "|", "tr", "-d", "'\\t\\n'")
	if err != nil {
		return "", errors.Wrap(err, "failed to get deployments")
	}
	deployment := strings.TrimSuffix(res.Status.Output, "\t\n")
	match, _ := regexp.MatchString(pattern, deployment)
	if !match {
		return "", fmt.Errorf("failed to find deployment name with pattern %q", pattern)
	}

	return deployment, nil
}

func getCredentials(ctx context.Context, e exec.Executor, om config.OpsManager, deployment string) (string, string, error) {
	res, err := credhub.Run(e, ctx, om, "get", "-n", fmt.Sprintf("/p-bosh/%s/cc-db-credentials", deployment), "-q")
	if err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("failed to get creds from %s", deployment))
	}

	scanner := bufio.NewScanner(strings.NewReader(res.Status.Output))
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		lines := strings.Split(string(data), "\n")
		if atEOF {
			return 0, nil, io.EOF
		}
		for _, l := range lines {
			fields := strings.Split(l, ": ")
			if len(fields) > 1 {
				return len(l) + 1, []byte(fields[1]), nil
			}
		}
		return 0, nil, nil
	})

	fields := make([]string, 0, 3)
	for scanner.Scan() {
		if scanner.Text() != "" {
			text := scanner.Text()
			fields = append(fields, text)
		}
	}
	if scanner.Err() != nil {
		return "", "", scanner.Err()
	}

	username := strings.TrimSuffix(fields[2], "\t\n")
	password := strings.TrimSuffix(fields[0], "\t\n")

	return username, password, nil
}

func getEncryptionKey(ctx context.Context, e exec.Executor, om config.OpsManager, deployment string) (string, error) {
	res, err := credhub.Run(e, ctx, om, "get", "-n", fmt.Sprintf("/opsmgr/%s/cloud_controller/db_encryption_credentials", deployment), "-q", "-k", "password", "|", "tr", "-d", "'\\t\\n'")
	if err != nil {
		return "", errors.Wrap(err, "failed to get encryption key")
	}

	return strings.TrimSuffix(res.Status.Output, "\t\n"), nil
}
