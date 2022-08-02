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

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func LoginSourceFoundation(e exec.Executor, om config.OpsManager, api, org, space, cfHome string) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		log.Debugf("Logging into '%s/%s' source api: %q, CF_HOME='%s'", org, space, api, cfHome)
		if err := om.Validate(); err != nil {
			return exec.Result{}, err
		}

		err := loginFoundation(ctx, api, org, space, cfHome, e, om)
		if err != nil {
			return exec.Result{}, fmt.Errorf("failed to login to source foundation: %w", err)
		}

		return exec.Result{}, nil
	}
}

func LoginTargetFoundation(e exec.Executor, om config.OpsManager, api, org, space, cfHome string) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		log.Debugf("Logging into '%s/%s' target api: %q, CF_HOME='%s'", org, space, api, cfHome)
		if err := om.Validate(); err != nil {
			return exec.Result{}, err
		}

		err := loginFoundation(ctx, api, org, space, cfHome, e, om)
		if err != nil {
			return exec.Result{}, fmt.Errorf("failed to login to target foundation: %w", err)
		}

		return exec.Result{}, nil
	}
}

func loginFoundation(ctx context.Context, api, org, space, cfHome string, e exec.Executor, om config.OpsManager) error {
	lines := []string{
		fmt.Sprintf(`products="$(OM_CLIENT_ID='%s' OM_CLIENT_SECRET='%s' OM_USERNAME='%s' OM_PASSWORD='%s' om -t %s -k curl -s -p /api/v0/staged/products)"`,
			om.ClientID,
			om.ClientSecret,
			om.Username,
			om.Password,
			om.URL),
		`product_guid="$(echo "$products" | jq -r '.[] | select(.type == "cf") | .guid')"`,
		fmt.Sprintf(`admin_credentials="$(OM_CLIENT_ID='%s' OM_CLIENT_SECRET='%s' OM_USERNAME='%s' OM_PASSWORD='%s' om -t %s -k curl -s -p /api/v0/deployed/products/"$product_guid"/credentials/.uaa.admin_credentials)"`,
			om.ClientID,
			om.ClientSecret,
			om.Username,
			om.Password,
			om.URL),
		`username="$(echo "$admin_credentials" | jq -r .credential.value.identity)"`,
		`password="$(echo "$admin_credentials" | jq -r .credential.value.password)"`,
		fmt.Sprintf(`CF_HOME='%s' cf login -a "%s" -u "$username" -p "$password" -o %s -s %s --skip-ssl-validation`, cfHome, api, org, space),
	}

	_, err := e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))

	return err
}
