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

package mysql

import (
	"context"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
)

type Migrator struct {
	sequence        flow.Flow
	migrationReader config.MigrationReader
}

func NewMigrator(flow flow.Flow, mr config.MigrationReader) *Migrator {
	return &Migrator{
		sequence:        flow,
		migrationReader: mr,
	}
}

func (m *Migrator) Validate(*cf.ServiceInstance, bool) error {
	return nil
}

func (m *Migrator) Migrate(ctx context.Context) (*cf.ServiceInstance, error) {
	dryRun := false
	if cfg, ok := config.FromContext(ctx); ok {
		dryRun = cfg.DryRun
	}

	data, err := m.migrationReader.GetMigration()
	if err != nil {
		return nil, err
	}

	var res flow.Result
	if res, err = m.sequence.Run(ctx, data, dryRun); err != nil {
		return nil, err
	}

	if si, ok := res.(*cf.ServiceInstance); ok {
		return si, nil
	}

	return nil, err
}
