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
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func Run(e exec.Executor, ctx context.Context, data interface{}, args ...string) (exec.Result, error) {
	foundation, ok := data.(config.OpsManager)
	if !ok {
		log.Fatal("failed to convert type to Foundation")
	}
	sshHost := foundation.Hostname
	if sshHost == "" {
		sshHost = foundation.IP
	}
	lines := []string{
		`ssh_key_path=$(mktemp)`,
		fmt.Sprintf(`cat "%s" >"$ssh_key_path"`, foundation.PrivateKey),
		`chmod 0600 "${ssh_key_path}"`,

		`bosh_ca_path=$(mktemp)`,
		fmt.Sprintf(`bosh_ca_cert="$(OM_CLIENT_ID='%s' OM_CLIENT_SECRET='%s' OM_USERNAME='%s' OM_PASSWORD='%s' om -t %s -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')"`,
			foundation.ClientID,
			foundation.ClientSecret,
			foundation.Username,
			foundation.Password,
			foundation.URL),
		`echo "$bosh_ca_cert" >"$bosh_ca_path"`,
		`chmod 0600 "${bosh_ca_path}"`,

		fmt.Sprintf(`creds="$(OM_CLIENT_ID='%s' OM_CLIENT_SECRET='%s' OM_USERNAME='%s' OM_PASSWORD='%s' om -t %s -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)"`,
			foundation.ClientID,
			foundation.ClientSecret,
			foundation.Username,
			foundation.Password,
			foundation.URL),
		`bosh_all="$(echo "$creds" | jq -r .credential | tr ' ' '\n' | grep '=')"`,
		`credhub_server="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_ENVIRONMENT=' | sed 's/BOSH_ENVIRONMENT/CREDHUB_SERVER/g'):8844"`,
		`credhub_client="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT=' | sed 's/BOSH_CLIENT/CREDHUB_CLIENT/g')"`,
		`credhub_secret="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT_SECRET=' | sed 's/BOSH_CLIENT_SECRET/CREDHUB_SECRET/g')"`,
		`credhub_ca_cert="CREDHUB_CA_CERT=$bosh_ca_path"`,
		fmt.Sprintf(`credhub_proxy="CREDHUB_PROXY=ssh+socks5://%s@%s:22?private-key=${ssh_key_path}"`, foundation.SshUser, sshHost),
	}

	if len(args) > 0 {
		lines = append(
			lines,
			`trap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT`,
			fmt.Sprintf(`/usr/bin/env "$credhub_server" "$credhub_client" "$credhub_secret" "$credhub_ca_cert" "$credhub_proxy" credhub %s`, strings.Join(args, " ")),
		)
	} else {
		lines = append(
			lines,
			`echo "export $credhub_server"`,
			`echo "export $credhub_client"`,
			`echo "export $credhub_secret"`,
			`echo "export $credhub_ca_cert"`,
			`echo "export $credhub_proxy"`,
		)
	}

	return e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
}
