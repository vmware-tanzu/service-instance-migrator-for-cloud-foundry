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

package credhub_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/credhub"
)

func TestImport(t *testing.T) {
	instance := &cf.ServiceInstance{
		Name:        "some-instance",
		GUID:        "some-guid",
		Credentials: map[string]interface{}{"key": "service.cf1.example.com"},
	}
	type args struct {
		instance *cf.ServiceInstance
		config   *config.Migration
		dryRun   bool
		ctx      context.Context
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		want    *cf.ServiceInstance
		wantErr error
	}{
		{
			name: "imports instance with domains replaced",
			args: args{
				instance: instance,
				config:   &config.Migration{},
				ctx: config.ContextWithConfig(context.TODO(), &config.Config{
					DomainsToReplace: map[string]string{"cf1.example.com": "cf2.example.com"},
				}),
			},
			step: credhub.SetCredentials(instance),
			want: &cf.ServiceInstance{
				Name:        "some-instance",
				GUID:        "some-guid",
				Credentials: map[string]interface{}{"key": "service.cf2.example.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flow.Sequence(tt.step).Run(tt.args.ctx, tt.args.config, tt.args.dryRun)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			}
			require.Equal(t, res, tt.want)
		})
	}
}
