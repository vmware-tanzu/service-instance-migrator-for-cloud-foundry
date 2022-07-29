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

package om

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
			name: "execute om with args",
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
				args: []string{"curl -s -p /api/v0/deployed/products/cf-7de431470b92530a463b/credentials/.uaa.credhub_admin_client_client_credentials"},
			},
			wantErr: false,
			want:    `OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t 'opsman.tas2.example.com' -k curl -s -p /api/v0/deployed/products/cf-7de431470b92530a463b/credentials/.uaa.credhub_admin_client_client_credentials`,
		},
		{
			name: "execute om without args",
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
				args: []string{},
			},
			wantErr: false,
			want: `echo "export OM_TARGET='opsman.tas2.example.com'"
echo "export OM_CLIENT_ID=''"
echo "export OM_CLIENT_SECRET=''"
echo "export OM_USERNAME='admin'"
echo "export OM_PASSWORD='admin-password'"`,
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
