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
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
)

func TestLoginFoundation(t *testing.T) {
	fakeSourceScriptExecutor := new(fakes.FakeExecutor)
	fakeTargetScriptExecutor := new(fakes.FakeExecutor)
	type args struct {
		config   *config.Migration
		dryRun   bool
		executor *fakes.FakeExecutor
		omConfig config.OpsManager
		api      string
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		wantErr error
		want    io.Reader
	}{
		{
			name: "login source foundation",
			args: args{
				dryRun: false,
				omConfig: config.OpsManager{
					URL:          "https://opsman.source.example.com",
					Username:     "fake-user",
					Password:     "fake-password",
					ClientID:     "fake-client-id",
					ClientSecret: "fake-client-secret",
				},
				executor: fakeSourceScriptExecutor,
				api:      "https://api.sys.source.example.com",
			},
			want: strings.NewReader(strings.Join([]string{
				`products="$(OM_CLIENT_ID='fake-client-id' OM_CLIENT_SECRET='fake-client-secret' OM_USERNAME='fake-user' OM_PASSWORD='fake-password' om -t https://opsman.source.example.com -k curl -s -p /api/v0/staged/products)"`,
				`product_guid="$(echo "$products" | jq -r '.[] | select(.type == "cf") | .guid')"`,
				`admin_credentials="$(OM_CLIENT_ID='fake-client-id' OM_CLIENT_SECRET='fake-client-secret' OM_USERNAME='fake-user' OM_PASSWORD='fake-password' om -t https://opsman.source.example.com -k curl -s -p /api/v0/deployed/products/"$product_guid"/credentials/.uaa.admin_credentials)"`,
				`username="$(echo "$admin_credentials" | jq -r .credential.value.identity)"`,
				`password="$(echo "$admin_credentials" | jq -r .credential.value.password)"`,
				`CF_HOME='.cf' cf login -a "https://api.sys.source.example.com" -u "$username" -p "$password" -o my-org -s my-space --skip-ssl-validation`,
			}, "\n")),
		},
		{
			name: "login target foundation",
			args: args{
				dryRun: false,
				omConfig: config.OpsManager{
					URL:          "https://opsman.target.example.com",
					Username:     "fake-user",
					Password:     "fake-password",
					ClientID:     "fake-client-id",
					ClientSecret: "fake-client-secret",
				},
				executor: fakeTargetScriptExecutor,
				api:      "https://api.sys.target.example.com",
			},
			want: strings.NewReader(strings.Join([]string{
				`products="$(OM_CLIENT_ID='fake-client-id' OM_CLIENT_SECRET='fake-client-secret' OM_USERNAME='fake-user' OM_PASSWORD='fake-password' om -t https://opsman.target.example.com -k curl -s -p /api/v0/staged/products)"`,
				`product_guid="$(echo "$products" | jq -r '.[] | select(.type == "cf") | .guid')"`,
				`admin_credentials="$(OM_CLIENT_ID='fake-client-id' OM_CLIENT_SECRET='fake-client-secret' OM_USERNAME='fake-user' OM_PASSWORD='fake-password' om -t https://opsman.target.example.com -k curl -s -p /api/v0/deployed/products/"$product_guid"/credentials/.uaa.admin_credentials)"`,
				`username="$(echo "$admin_credentials" | jq -r .credential.value.identity)"`,
				`password="$(echo "$admin_credentials" | jq -r .credential.value.password)"`,
				`CF_HOME='.cf' cf login -a "https://api.sys.target.example.com" -u "$username" -p "$password" -o my-org -s my-space --skip-ssl-validation`,
			}, "\n")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			login := LoginTargetFoundation(tt.args.executor, tt.args.omConfig, tt.args.api, "my-org", "my-space", ".cf")
			if _, err := flow.Sequence(login).Run(context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
		require.Equal(t, 1, tt.args.executor.ExecuteCallCount())
		_, got := tt.args.executor.ExecuteArgsForCall(0)
		require.Equal(t, tt.want, got)
	}
}
