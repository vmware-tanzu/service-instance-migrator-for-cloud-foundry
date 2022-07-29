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
	"testing"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
)

func TestGetServiceInstance(t *testing.T) {
	fakeScriptExecutor := new(fakes.FakeExecutor)
	fakeScriptExecutor.ExecuteReturnsOnCall(0, exec.Result{
		Output: `Creating service instance some-instance in org test-org / space test-space as some-user...
OK

Create in progress. Use 'cf services' or 'cf service some-instance' to check operation status.`,
	}, nil)
	fakeScriptExecutor.ExecuteReturnsOnCall(1, exec.Result{
		Output: "succeeded",
	}, nil)
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		want    io.Reader
		wantErr error
	}{
		{
			name: "get a service instance",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			wantErr: nil,
			step: GetServiceInstance(fakeScriptExecutor,
				".cf",
				&ServiceInstance{
					Name:    "some-instance",
					Service: "some-service",
					Plan:    "some-plan",
				}, time.Second*1, time.Millisecond*100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
		require.GreaterOrEqual(t, fakeScriptExecutor.ExecuteCallCount(), 2)
		_, got1 := fakeScriptExecutor.ExecuteArgsForCall(0)
		require.Equal(t, "CF_HOME='.cf' cf service 'some-instance' | grep -i 'status:' | awk '{print $NF}'", copyFrom(t, got1).String())

		_, got2 := fakeScriptExecutor.ExecuteArgsForCall(fakeScriptExecutor.ExecuteCallCount() - 1)
		require.Equal(t, "CF_HOME='.cf' cf service 'some-instance' --guid", copyFrom(t, got2).String())
	}
}
