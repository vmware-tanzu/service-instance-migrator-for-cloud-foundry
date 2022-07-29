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

package bosh

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
)

func TestScriptRunner_Run(t *testing.T) {
	fakeScriptExecutor1 := new(fakes.FakeExecutor)
	fakeScriptExecutor2 := new(fakes.FakeExecutor)
	type args struct {
		ctx        context.Context
		executor   *fakes.FakeExecutor
		foundation config.OpsManager
		args       []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "execute bosh with args",
			args: args{
				ctx:      context.TODO(),
				executor: fakeScriptExecutor1,
				foundation: config.OpsManager{
					URL:          "opsman.tas2.example.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "",
					ClientSecret: "",
					PrivateKey:   "opsman-private-key",
					IP:           "10.1.1.1",
					SshUser:      "ubuntu",
				},
				args: []string{"arg1", "arg2", "arg3"},
			},
			wantErr: false,
			want: `ssh_key_path=$(mktemp)
cat "opsman-private-key" >"$ssh_key_path"
chmod 0600 "${ssh_key_path}"
bosh_ca_path=$(mktemp)
bosh_ca_cert="$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')"
echo "$bosh_ca_cert" >"$bosh_ca_path"
chmod 0600 "${bosh_ca_path}"
creds="$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)"
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
/usr/bin/env "$bosh_client" "$bosh_env" "$bosh_secret" "$bosh_ca_cert" "$bosh_proxy" "$bosh_gw_host" "$bosh_gw_user" "$bosh_gw_private_key" bosh arg1 arg2 arg3`,
		},
		{
			name: "execute bosh without args",
			args: args{
				ctx:      context.TODO(),
				executor: fakeScriptExecutor2,
				foundation: config.OpsManager{
					URL:          "opsman.tas2.example.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "",
					ClientSecret: "",
					PrivateKey:   "opsman-private-key",
					IP:           "10.1.1.1",
					SshUser:      "ubuntu",
				},
				args: nil,
			},
			wantErr: false,
			want: `ssh_key_path=$(mktemp)
cat "opsman-private-key" >"$ssh_key_path"
chmod 0600 "${ssh_key_path}"
bosh_ca_path=$(mktemp)
bosh_ca_cert="$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')"
echo "$bosh_ca_cert" >"$bosh_ca_path"
chmod 0600 "${bosh_ca_path}"
creds="$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)"
bosh_all="$(echo "$creds" | jq -r .credential | tr ' ' '\n' | grep '=')"
bosh_client="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT=')"
bosh_env="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_ENVIRONMENT=')"
bosh_secret="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT_SECRET=')"
bosh_ca_cert="BOSH_CA_CERT=$bosh_ca_path"
bosh_proxy="BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}"
bosh_gw_host="BOSH_GW_HOST=10.1.1.1"
bosh_gw_user="BOSH_GW_USER=ubuntu"
bosh_gw_private_key="BOSH_GW_PRIVATE_KEY=${ssh_key_path}"
echo "export BOSH_ENV_NAME=10.1.1.1"
echo "export $bosh_client"
echo "export $bosh_env"
echo "export $bosh_secret"
echo "export $bosh_ca_cert"
echo "export $bosh_proxy"
echo "export $bosh_gw_host"
echo "export $bosh_gw_user"
echo "export $bosh_gw_private_key"
echo "export CREDHUB_SERVER=\"\${BOSH_ENVIRONMENT}:8844\""
echo "export CREDHUB_PROXY=\"\${BOSH_ALL_PROXY}\""
echo "export CREDHUB_CLIENT=\"\${BOSH_CLIENT}\""
echo "export CREDHUB_SECRET=\"\${BOSH_CLIENT_SECRET}\""
echo "export CREDHUB_CA_CERT=\"\${BOSH_CA_CERT}\""`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Run(tt.args.executor, tt.args.ctx, tt.args.foundation, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		require.Equal(t, 1, tt.args.executor.ExecuteCallCount())
		_, got := tt.args.executor.ExecuteArgsForCall(0)
		require.Equal(t, tt.want, copyFrom(t, got).String())
	}
}

func copyFrom(t *testing.T, r io.Reader) *bytes.Buffer {
	dst := &bytes.Buffer{}
	_, err := io.Copy(dst, r)
	require.NoError(t, err)
	return dst
}
