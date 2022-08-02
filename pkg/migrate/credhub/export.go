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
	"bufio"
	credhubcreds "code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
)

type CredentialsExtractor func(s string) (map[string]interface{}, error)

type ClientHolder interface {
	SourceCFClient() cf.Client
	TargetCFClient() cf.Client
	SourceBoshClient() bosh.Client
	TargetBoshClient() bosh.Client
}

func NewExportSequence(org, space string, instance *cf.ServiceInstance, om config.OpsManager, h ClientHolder, executor exec.Executor) flow.Flow {
	cfHome, err := os.MkdirTemp("", instance.GUID)
	if err != nil {
		panic("failed to create CF_HOME")
	}

	api := h.SourceCFClient().GetClientConfig().ApiAddress

	return flow.ProgressBarSequence(
		fmt.Sprintf("Exporting %s", instance.Name),
		flow.StepWithProgressBar(
			cf.LoginSourceFoundation(executor, om, api, org, space, cfHome),
			flow.WithDisplay("Logging into source foundation"),
		),
		flow.StepWithProgressBar(
			RetrieveCredhubCredentials(h.SourceBoshClient(), om, executor, instance, credsExtractor),
			flow.WithDisplay("Retrieving credhub credentials"),
		),
	)
}

func RetrieveCredhubCredentials(bc bosh.Client, om config.OpsManager, e exec.Executor, instance *cf.ServiceInstance, credsExtractor CredentialsExtractor) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		credhubRef, err := lookupCredhubRef(*instance)
		if err != nil {
			return exec.Result{}, err
		}

		deployment, found, err := bc.FindDeployment("^cf-")
		if err != nil {
			return exec.Result{}, err
		}
		if !found {
			return exec.Result{}, fmt.Errorf("cf deployment not found")
		}

		credhubAdminSecret, err := findCredhubAdminSecret(ctx, e, om, deployment.Name)
		if err != nil {
			return exec.Result{}, err
		}

		instanceName, err := findInstanceName(ctx, e, om, deployment.Name)
		if err != nil {
			return exec.Result{}, err
		}

		log.Debugf("Retrieving credhub credentials from bosh deployment %q, instance %q", deployment.Name, instanceName)

		res, err := bosh.Run(e, ctx, om, "-d", fmt.Sprintf("'%s'", deployment.Name), "ssh", fmt.Sprintf("'%s'", instanceName), "-c", fmt.Sprintf(`'
export CREDHUB_SECRET="%s"
export NAME="%s"
ACCESS_TOKEN=$(curl -k -d "client_id=credhub_admin_client&client_secret=$CREDHUB_SECRET&grant_type=client_credentials&token_format=jwt" https://uaa.service.cf.internal:8443/oauth/token -s -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Accept: application/json" | grep -Eo "access_token"[^,]* | grep -Eo [^:]*$ | tr -d "\"")
curl -k "https://credhub.service.cf.internal:8844/api/v1/data?name=$NAME&current=true" -s -X GET -H "Host: credhub.service.cf.internal" -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN"
'`, credhubAdminSecret, credhubRef))
		if err != nil {
			return res, errors.Wrap(err, fmt.Sprintf("failed to retrieve credhub credentials from '%s'", instanceName))
		}

		if len(res.Status.Output) == 0 {
			return res, fmt.Errorf("couldn't extract credentials, output is empty")
		}

		creds, err := credsExtractor(res.Status.Output)
		if err != nil {
			return res, err
		}
		instance.Credentials = creds

		return res, nil
	}
}

func lookupCredhubRef(instance cf.ServiceInstance) (interface{}, error) {
	for _, b := range instance.ServiceBindings {
		if val, ok := b.Credentials["credhub-ref"]; ok {
			return val, nil
		}
	}
	return "", fmt.Errorf("failed to find credhub-ref in service binding for instance guid '%s'", instance.GUID)
}

func findInstanceName(ctx context.Context, e exec.Executor, om config.OpsManager, deployment string) (string, error) {
	res, err := bosh.Run(e, ctx, om, "-d", fmt.Sprintf("'%s'", deployment), "--column=instance", "--column=process", "is", "-p", "|", "grep", "credhub", "|", "awk", "'{print $1}'", "|", "head", "-1", "|", "tr", "-d", "'\\t\\n'")
	if err != nil {
		return "", errors.Wrap(err, "failed to bosh instance running credhub")
	}

	return strings.TrimSuffix(res.Status.Output, "\t\n"), nil
}

func findCredhubAdminSecret(ctx context.Context, e exec.Executor, opsman config.OpsManager, deploymentName string) (string, error) {
	res, err := om.Run(e, ctx, opsman,
		fmt.Sprintf("curl -s -p /api/v0/deployed/products/%s/credentials/.uaa.credhub_admin_client_client_credentials | jq -r .credential.value.password", deploymentName))
	if err != nil {
		return "", errors.Wrap(err, "failed to find credhub admin credentials")
	}
	status := res.Status
	credhubAdminSecret := strings.TrimSuffix(status.Output, "\n")

	return credhubAdminSecret, nil
}

func credsExtractor(input string) (map[string]interface{}, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("couldn't extract encryption key, output is empty")
	}

	creds, err := readWithReadLine(input)
	if err != nil {
		return nil, err
	}

	subFields := strings.SplitAfter(creds, "]}")
	log.Debugf("Service instance creds: '%s'", subFields[0])

	var cred credhubcreds.JSON
	data := struct {
		Data []credhubcreds.JSON
	}{[]credhubcreds.JSON{cred}}
	err = json.Unmarshal([]byte(subFields[0]), &data)
	if err != nil {
		return nil, err
	}

	return data.Data[0].Value, nil
}

func readWithReadLine(input string) (string, error) {
	reader := bufio.NewReaderSize(strings.NewReader(input), 10*1024)
	fields := make([]string, 0, 9)

	for {
		line, err := read(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		fields = append(fields, string(line))
	}
	creds := fields[len(fields)-1]

	return creds, nil
}

func read(r *bufio.Reader) ([]byte, error) {
	var (
		isPrefix = true
		err      error
		line, ln []byte
	)

	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		fields := strings.Split(string(line), "| ")
		if len(fields) > 1 {
			ln = append(ln, []byte(fields[1])...)
		}
	}

	return ln, err
}
