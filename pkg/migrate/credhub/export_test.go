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
	"bytes"
	"context"
	"errors"
	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/fakes"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	execfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
)

func TestRetrieveCredhubCredentials(t *testing.T) {
	fakeScriptExecutor := new(execfakes.FakeExecutor)
	type args struct {
		boshClient     *fakes.FakeClient
		config         *config.Migration
		si             *cf.ServiceInstance
		credsExtractor func(string) (map[string]interface{}, error)
		dryRun         bool
		want1          string
		want2          string
		want3          string
		want4          string
		om             config.OpsManager
	}
	tests := []struct {
		name               string
		args               args
		wantErr            error
		fakeScriptExecutor *execfakes.FakeExecutor
	}{
		{
			name: "retrieves credhub credentials",
			args: args{
				boshClient: &fakes.FakeClient{
					FindDeploymentStub: func(s string) (director.DeploymentResp, bool, error) {
						return director.DeploymentResp{
							Name: "cf-7de431470b92530a463b",
						}, true, nil
					},
				},
				dryRun: false,
				si: &cf.ServiceInstance{
					GUID: "some-guid",
					ServiceBindings: []cf.ServiceBinding{
						{
							Credentials: map[string]interface{}{
								"credhub-ref": "/credhub-service-broker/credhub/99f0c44c-1cb5-4c5f-91a8-873090e035aa/credentials",
							},
						},
					},
				},
				om: config.OpsManager{
					URL:          "opsman.tas1.example.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "",
					ClientSecret: "",
					PrivateKey:   "opsman-private-key",
					IP:           "10.1.1.1",
					SshUser:      "ubuntu",
				},
				config:         &config.Migration{},
				credsExtractor: credsExtractor,
				want1:          "ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas1.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas1.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh deps --column=name | grep '^cf-' | tr -d '\\t\\n'",
				want2:          "OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t 'opsman.tas1.example.com' -k curl -s -p /api/v0/deployed/products/cf-7de431470b92530a463b/credentials/.uaa.credhub_admin_client_client_credentials | jq -r .credential.value.password",
				want3:          "ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas1.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas1.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh -d 'cf-7de431470b92530a463b' --column=instance --column=process is -p | grep credhub | awk '{print $1}' | head -1 | tr -d '\\t\\n'",
				want4: `ssh_key_path=$(mktemp)
cat "opsman-private-key" >"$ssh_key_path"
chmod 0600 "${ssh_key_path}"
bosh_ca_path=$(mktemp)
bosh_ca_cert="$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas1.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')"
echo "$bosh_ca_cert" >"$bosh_ca_path"
chmod 0600 "${bosh_ca_path}"
creds="$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas1.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)"
bosh_all="$(echo "$creds" | jq -r .credential | tr ' ' '\n' | grep '=')"
bosh_client="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT=')"
bosh_env="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_ENVIRONMENT=')"
bosh_secret="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT_SECRET=')"
bosh_ca_cert="BOSH_CA_CERT=$bosh_ca_path"
bosh_proxy="BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}"
bosh_gw_host="BOSH_GW_HOST=10.1.1.1"
bosh_gw_user="BOSH_GW_USER=ubuntu"
bosh_gw_private_key="BOSH_GW_PRIVATE_KEY=${ssh_key_path}"
trap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT
/usr/bin/env "$bosh_client" "$bosh_env" "$bosh_secret" "$bosh_ca_cert" "$bosh_proxy" "$bosh_gw_host" "$bosh_gw_user" "$bosh_gw_private_key" bosh -d 'cf-7de431470b92530a463b' ssh 'control/some-guid' -c '
export CREDHUB_SECRET="credhub-secret"
export NAME="/credhub-service-broker/credhub/99f0c44c-1cb5-4c5f-91a8-873090e035aa/credentials"
ACCESS_TOKEN=$(curl -k -d "client_id=credhub_admin_client&client_secret=$CREDHUB_SECRET&grant_type=client_credentials&token_format=jwt" https://uaa.service.cf.internal:8443/oauth/token -s -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Accept: application/json" | grep -Eo "access_token"[^,]* | grep -Eo [^:]*$ | tr -d "\"")
curl -k "https://credhub.service.cf.internal:8844/api/v1/data?name=$NAME&current=true" -s -X GET -H "Host: credhub.service.cf.internal" -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN"
'`,
			},
			fakeScriptExecutor: fakeScriptExecutor,
		},
		{
			name: "returns error when credhub-ref does not exist",
			args: args{
				boshClient: &fakes.FakeClient{
					FindDeploymentStub: func(s string) (director.DeploymentResp, bool, error) {
						return director.DeploymentResp{
							Name: "cf",
						}, true, nil
					},
				},
				dryRun: false,
				si: &cf.ServiceInstance{
					GUID: "some-guid",
				},
				om: config.OpsManager{
					URL:          "opsman.tas1.example.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "",
					ClientSecret: "",
					PrivateKey:   "opsman-private-key",
					IP:           "10.1.1.1",
					SshUser:      "ubuntu",
				},
				config: &config.Migration{},
			},
			wantErr:            errors.New("failed to find credhub-ref in service binding for instance guid 'some-guid'"),
			fakeScriptExecutor: fakeScriptExecutor,
		},
	}
	for _, tt := range tests {
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(0, exec.Result{
			Status: &exec.Status{
				Output: `credhub-secret`,
			},
		}, nil)
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(1, exec.Result{
			Status: &exec.Status{
				Output: `control/some-guid`,
			},
		}, nil)
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(2, exec.Result{
			Status: &exec.Status{
				Output: `
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | Unauthorized use is strictly prohibited. All access and activity
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | is subject to logging and monitoring.
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stdout | {"data":[{"type":"json","version_created_at":"2022-01-29T00:31:50Z","id":"2e841518-6394-4771-b5f6-3ad355755279","name":"/credhub-service-broker/credhub/2fd5b24c-3d95-4594-af75-8eb418cf556f/credentials","metadata":null,"value":{"service":{"serviceinfo":{"password":"pwd","url":"https://svc.example.com","username":"user"}}}}]}Connection to 192.168.2.23 closed.
`,
			},
		}, nil)
		t.Run(tt.name, func(t *testing.T) {
			_, err := flow.RunWith(RetrieveCredhubCredentials(tt.args.boshClient, tt.args.om, tt.fakeScriptExecutor, tt.args.si, tt.args.credsExtractor), context.TODO(), tt.args.config, tt.args.dryRun)
			if err != nil && tt.wantErr == nil {
				require.NoError(t, err)
			} else if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				require.EqualError(t, err, tt.wantErr.Error())
			}
		})
		require.Equal(t, 3, tt.fakeScriptExecutor.ExecuteCallCount())
		_, got2 := tt.fakeScriptExecutor.ExecuteArgsForCall(0)
		require.Equal(t, tt.args.want2, copyFrom(t, got2).String())
		_, got3 := tt.fakeScriptExecutor.ExecuteArgsForCall(1)
		require.Equal(t, tt.args.want3, copyFrom(t, got3).String())
		_, got4 := tt.fakeScriptExecutor.ExecuteArgsForCall(2)
		require.Equal(t, tt.args.want4, copyFrom(t, got4).String())
	}
}

func copyFrom(t *testing.T, r io.Reader) *bytes.Buffer {
	dst := &bytes.Buffer{}
	_, err := io.Copy(dst, r)
	require.NoError(t, err)
	return dst
}

func Test_credsExtractor(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "returns credentials as a map",
			args: args{
				input: `control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | Unauthorized use is strictly prohibited. All access and activity
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | is subject to logging and monitoring.
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stdout | {"data":[{"type":"json","version_created_at":"2022-01-29T00:31:50Z","id":"2e841518-6394-4771-b5f6-3ad355755279","name":"/credhub-service-broker/credhub/2fd5b24c-3d95-4594- f75-8eb418cf556f/credentials","metadata":null,"value":{"serviceinfo":{"password":"pwd","url":"https://svc.example.com","username":"user"}}}]}Connection to 192.168.2.23 closed.
`,
			},
			want: map[string]interface{}{
				"serviceinfo": map[string]interface{}{
					"password": "pwd",
					"url":      "https://svc.example.com",
					"username": "user",
				},
			},
			wantErr: false,
		},
		{
			name: "does not error if input is nested",
			args: args{
				input: `control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | Unauthorized use is strictly prohibited. All access and activity
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | is subject to logging and monitoring.
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stdout | {"data":[{"type":"json","version_created_at":"2022-01-29T00:31:50Z","id":"2e841518-6394-4771-b5f6-3ad355755279","name":"/credhub-service-broker/credhub/2fd5b24c-3d95-4594-af75-8eb418cf556f/credentials","metadata":null,"value":{"service":{"serviceinfo":{"password":"pwd","url":"https://svc.example.com","username":"user"}}}}]}Connection to 192.168.2.23 closed.
`,
			},
			want: map[string]interface{}{
				"service": map[string]interface{}{
					"serviceinfo": map[string]interface{}{
						"password": "pwd",
						"url":      "https://svc.example.com",
						"username": "user",
					},
				}},
			wantErr: false,
		},
		{
			name: "does not error if data is large",
			args: args{
				input: `control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | Unauthorized use is strictly prohibited. All access and activity
control/1bc628b5-c094-4164-b2bd-39db5c4553d9: stderr | is subject to logging and monitoring.
credhub/47595943-3c37-4e13-add9-b091aaaa7e1e: stdout | {"data":[{"type":"json","version_created_at":"2021-02-16T19:00:38Z","id":"8141bf22-12d4-4fc6-a7f9-775798e6e015","name":"/credhub-service-broker/credhub/aabccb7a-5f04-46d3-9ab3-af7db83ee5d1/credentials","value":{"name":"acqauth","value":"dwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqxdrxicwsazmqxybrleuqebntvmqajsaqmxcytqrzgridgrathhtqftvpzxogycrixxfwdfsdsdasffsadasdasdasdwedfsdfjsdlkfjlasdjlasdjlaksjdwejaskdfxmdcasdjasdlasdasldjaskdjlasjdaalskdjasjkdalksjdas"}}]}Connection to 10.59.196.177 closed.
`,
			},
			want: map[string]interface{}{
				"name":  "acqauth",
				"value": "dwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqsdrxicwsazmqxybrleuqebntvmqajsaqmscytqrzgridgrathhtqftvpzxogycrixxdwuhkqvbylxykekthkzobldbtqjwrtxjsnkthbobjaqzfpsicpzsyifbfrtzdxzfisvothudglfaypfgncrplajygwigvcdssqxdrxicwsazmqxybrleuqebntvmqajsaqmxcytqrzgridgrathhtqftvpzxogycrixxfwdfsdsdasffsadasdasdasdwedfsdfjsdlkfjlasdjlasdjlaksjdwejaskdfxmdcasdjasdlasdasldjaskdjlasjdaalskdjasjkdalksjdas",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := credsExtractor(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("credsExtractor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("credsExtractor() got = %v, want %v", got, tt.want)
			}
		})
	}
}
