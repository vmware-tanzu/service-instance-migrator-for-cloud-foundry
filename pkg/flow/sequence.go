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
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Result

type Result interface{}

//counterfeiter:generate -o fakes . Flow

type Flow interface {
	Run(ctx context.Context, data interface{}, dryRun bool) (Result, error)
}

type StepFunc func(ctx context.Context, data interface{}, dryRun bool) (Result, error)

func (fn StepFunc) Run(ctx context.Context, data interface{}, dryRun bool) (Result, error) {
	return fn(ctx, data, dryRun)
}

func RunWith(step Flow, ctx context.Context, data interface{}, dryRun bool) (Result, error) {
	return step.Run(ctx, data, dryRun)
}

func Sequence(steps ...StepFunc) Flow {
	return StepFunc(func(ctx context.Context, data interface{}, dryRun bool) (Result, error) {
		var res Result
		var err error
		for _, step := range steps {
			res, err = step(ctx, data, dryRun)
			if err != nil {
				return res, err
			}
		}
		return res, nil
	})
}
