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

package flow

import (
	"context"
	"fmt"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

type ProgressBarStep struct {
	stepFn  StepFunc
	display string
	bar     *mpb.Bar
}

func (p ProgressBarStep) String() string {
	return p.display
}

func (p ProgressBarStep) Bar() *mpb.Bar {
	return p.bar
}

type ProgressBarOption func(step *ProgressBarStep)

func StepWithProgressBar(step StepFunc, opts ...ProgressBarOption) *ProgressBarStep {
	pbs := &ProgressBarStep{stepFn: step}

	for _, o := range opts {
		o(pbs)
	}

	return pbs
}

func WithDisplay(display string) ProgressBarOption {
	return func(step *ProgressBarStep) {
		step.display = display
	}
}

func ProgressBarSequence(msg string, steps ...*ProgressBarStep) Flow {
	return StepFunc(func(ctx context.Context, data interface{}, dryRun bool) (Result, error) {
		var bar *mpb.Bar
		if p, ok := config.ProgressFromContext(ctx); ok {
			bar = p.AddBar(int64(len(steps)),
				mpb.PrependDecorators(
					Any(msg, steps, decor.WC{W: len(msg) + 1, C: decor.DSyncSpace}),
					OnComplete(msg, steps, decor.WC{W: len(msg) + 1, C: decor.DSyncSpace}),
				),
				mpb.AppendDecorators(decor.Percentage(decor.WCSyncSpace)),
			)
		}

		var res Result
		var err error
		for _, step := range steps {
			step.bar = bar
			res, err = step.stepFn(ctx, data, dryRun)
			if err != nil {
				return res, err
			}
			if bar != nil {
				bar.Increment()
			}
		}
		return res, nil
	})
}

func Any(msg string, steps []*ProgressBarStep, wcc ...decor.WC) decor.Decorator {
	return decor.Any(func(s decor.Statistics) string {
		if s.Current >= int64(len(steps)) {
			return ""
		}
		return msg
	}, wcc...)
}

func OnComplete(msg string, steps []*ProgressBarStep, wcc ...decor.WC) decor.Decorator {
	return decor.OnComplete(
		decor.Any(func(s decor.Statistics) string {
			return fmt.Sprintf("[\x1b[31m%s\x1b[0m]", steps[s.Current])
		}, wcc...), msg)
}
