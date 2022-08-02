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

package config

import (
	"context"
	"github.com/vbauerster/mpb/v7"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/report"
)

type key int

const (
	configKey key = iota
	reportSummaryKey
	progressKey
)

func ContextWithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

func FromContext(ctx context.Context) (*Config, bool) {
	cfg, ok := ctx.Value(configKey).(*Config)
	if cfg == nil {
		return cfg, false
	}
	return cfg, ok
}

func ContextWithSummary(ctx context.Context, summary *report.Summary) context.Context {
	return context.WithValue(ctx, reportSummaryKey, summary)
}

func SummaryFromContext(ctx context.Context) (*report.Summary, bool) {
	summary, ok := ctx.Value(reportSummaryKey).(*report.Summary)
	if summary == nil {
		return summary, false
	}
	return summary, ok
}

func ContextWithProgress(ctx context.Context, p *mpb.Progress) context.Context {
	return context.WithValue(ctx, progressKey, p)
}

func ProgressFromContext(ctx context.Context) (*mpb.Progress, bool) {
	progress, ok := ctx.Value(progressKey).(*mpb.Progress)
	if progress == nil {
		return progress, false
	}
	return progress, ok
}
