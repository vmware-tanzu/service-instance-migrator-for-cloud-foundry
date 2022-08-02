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

package flow_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vbauerster/mpb/v7"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"testing"
)

func TestProgressBarSequence(t *testing.T) {
	flowWriter1 := &bytes.Buffer{}
	flowWriter2 := &bytes.Buffer{}
	type args struct {
		format string
		steps  []*flow.ProgressBarStep
	}
	tests := []struct {
		name    string
		args    args
		display []string
		writer  *bytes.Buffer
	}{
		{
			name: "execution flow prints noop",
			args: args{
				format: "Exporting [%s]",
				steps: []*flow.ProgressBarStep{
					flow.StepWithProgressBar(func(ctx context.Context, data interface{}, dryRun bool) (flow.Result, error) {
						_, err := fmt.Fprintln(flowWriter1, data.(string)+" 1")
						return struct{}{}, err
					}),
					flow.StepWithProgressBar(func(ctx context.Context, data interface{}, dryRun bool) (flow.Result, error) {
						_, err := fmt.Fprintln(flowWriter1, data.(string)+" 2")
						return struct{}{}, err
					}),
				},
			},
			display: []string{"", ""},
			writer:  flowWriter1,
		},
		{
			name: "execution flow prints noop with display",
			args: args{
				format: "Exporting [%s]",
				steps: []*flow.ProgressBarStep{
					flow.StepWithProgressBar(func(ctx context.Context, data interface{}, dryRun bool) (flow.Result, error) {
						_, err := fmt.Fprintln(flowWriter2, data.(string)+" 1")
						return struct{}{}, err
					}, flow.WithDisplay("something 1")),
					flow.StepWithProgressBar(func(ctx context.Context, data interface{}, dryRun bool) (flow.Result, error) {
						_, err := fmt.Fprintln(flowWriter2, data.(string)+" 2")
						return struct{}{}, err
					}, flow.WithDisplay("something 2")),
				},
			},
			display: []string{"something 1", "something 2"},
			writer:  flowWriter2,
		},
	}
	p := mpb.New(mpb.WithWidth(64))
	ctx := config.ContextWithProgress(context.TODO(), p)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := flow.ProgressBarSequence(tt.args.format, tt.args.steps...)
			_, err := flow.RunWith(seq, ctx, "noop", false)
			assert.NoError(t, err)
			require.Equal(t, "noop 1\nnoop 2\n", tt.writer.String())
			assert.Equal(t, tt.args.steps[0].String(), tt.display[0])
			assert.Equal(t, tt.args.steps[1].String(), tt.display[1])
			assert.True(t, tt.args.steps[0].Bar().Completed())
			assert.True(t, tt.args.steps[1].Bar().Completed())
		})
	}
}
