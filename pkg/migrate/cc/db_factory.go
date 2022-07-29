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

package cc

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
)

type DatabaseFactory struct {
	cfg *Config
}

func NewDatabaseFactory(cfg *Config) DatabaseFactory {
	return DatabaseFactory{
		cfg: cfg,
	}
}

func (f DatabaseFactory) NewCCDB(isExport bool) (db.Repository, error) {
	var cfg = f.cfg.TargetCloudControllerDatabase
	if isExport {
		cfg = f.cfg.SourceCloudControllerDatabase
	}

	ccdb, err := db.NewCCDBConnection(
		cfg.Host,
		cfg.Username,
		cfg.Password,
		cfg.SSHHost,
		cfg.SSHUsername,
		cfg.SSHPassword,
		cfg.SSHPrivateKey,
		cfg.TunnelRequired,
	)
	if err != nil {
		return nil, err
	}

	database := &db.CloudController{
		DB: ccdb,
	}

	return database, nil
}
