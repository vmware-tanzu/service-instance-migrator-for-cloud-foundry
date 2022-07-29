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
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
)

func TestCreateServiceInstance(t *testing.T) {
	fakeScriptExecutor := new(fakes.FakeExecutor)
	fakeScriptExecutorForCredentials := new(fakes.FakeExecutor)
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name               string
		args               args
		step               flow.StepFunc
		want1              string
		want2              string
		wantErr            error
		fakeScriptExecutor *fakes.FakeExecutor
	}{
		{
			name: "creates a service instance",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			fakeScriptExecutor: fakeScriptExecutor,
			step: CreateServiceInstance(fakeScriptExecutor,
				".cf",
				ServiceInstance{
					Name:    "some-instance",
					Service: "some-service",
					Plan:    "some-plan",
				}),
			want1: "CF_HOME='.cf' cf service 'some-instance' --guid",
			want2: "CF_HOME='.cf' cf create-service 'some-service' 'some-plan' 'some-instance'",
		},
		{
			name: "creates a service instance with credentials",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			fakeScriptExecutor: fakeScriptExecutorForCredentials,
			step: CreateServiceInstance(fakeScriptExecutorForCredentials,
				".cf",
				ServiceInstance{
					Name:    "some-instance",
					Service: "some-service",
					Plan:    "some-plan",
					Credentials: map[string]interface{}{
						"read-only": true,
					},
				}),
			want1: "CF_HOME='.cf' cf service 'some-instance' --guid",
			want2: "CF_HOME='.cf' cf create-service 'some-service' 'some-plan' 'some-instance' -c '{\"read-only\":true}'",
		},
	}
	for _, tt := range tests {
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(0, exec.Result{
			Output: `Showing info of service some-instance in org test-org / space test-space as some-user...

Service instance some-instance not found
FAILED`,
		}, nil)
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(1, exec.Result{
			Output: `Creating service instance some-instance in org test-org / space test-space as some-user...
OK

Create in progress. Use 'cf services' or 'cf service some-instance' to check operation status.`,
		}, nil)
		t.Run(tt.name, func(t *testing.T) {
			if _, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
		require.Equal(t, 2, tt.fakeScriptExecutor.ExecuteCallCount())
		_, got1 := tt.fakeScriptExecutor.ExecuteArgsForCall(0)
		require.Equal(t, tt.want1, copyFrom(t, got1).String())

		_, got2 := tt.fakeScriptExecutor.ExecuteArgsForCall(1)
		require.Equal(t, tt.want2, copyFrom(t, got2).String())
	}
}

func copyFrom(t *testing.T, r io.Reader) *bytes.Buffer {
	dst := &bytes.Buffer{}
	_, err := io.Copy(dst, r)
	require.NoError(t, err)
	return dst
}
