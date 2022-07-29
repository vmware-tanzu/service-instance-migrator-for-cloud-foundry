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
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
)

func TestScriptRunner_Run(t *testing.T) {
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
			name: "execute credhub with args",
			args: args{
				ctx:      context.TODO(),
				executor: &fakes.FakeExecutor{},
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
credhub_server="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_ENVIRONMENT=' | sed 's/BOSH_ENVIRONMENT/CREDHUB_SERVER/g'):8844"
credhub_client="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT=' | sed 's/BOSH_CLIENT/CREDHUB_CLIENT/g')"
credhub_secret="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT_SECRET=' | sed 's/BOSH_CLIENT_SECRET/CREDHUB_SECRET/g')"
credhub_ca_cert="CREDHUB_CA_CERT=$bosh_ca_path"
credhub_proxy="CREDHUB_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}"
trap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT
/usr/bin/env "$credhub_server" "$credhub_client" "$credhub_secret" "$credhub_ca_cert" "$credhub_proxy" credhub arg1 arg2 arg3`,
		},
		{
			name: "execute credhub without args",
			args: args{
				ctx:      context.TODO(),
				executor: &fakes.FakeExecutor{},
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
				args: []string{},
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
credhub_server="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_ENVIRONMENT=' | sed 's/BOSH_ENVIRONMENT/CREDHUB_SERVER/g'):8844"
credhub_client="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT=' | sed 's/BOSH_CLIENT/CREDHUB_CLIENT/g')"
credhub_secret="$(echo "$bosh_all" | tr ' ' '\n' | grep 'BOSH_CLIENT_SECRET=' | sed 's/BOSH_CLIENT_SECRET/CREDHUB_SECRET/g')"
credhub_ca_cert="CREDHUB_CA_CERT=$bosh_ca_path"
credhub_proxy="CREDHUB_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}"
echo "export $credhub_server"
echo "export $credhub_client"
echo "export $credhub_secret"
echo "export $credhub_ca_cert"
echo "export $credhub_proxy"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Run(tt.args.executor, tt.args.ctx, tt.args.foundation, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
