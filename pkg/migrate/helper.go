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

package migrate

import (
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"sync"
)

type (
	Migrator    int64
	Service     string
	ServiceType string
)

const (
	ECS Migrator = iota
	MySQL
	SQLServer
	CredHub
)

const (
	ECSBucketService       Service = "ecs-bucket"
	MySQLService           Service = "p.mysql"
	SQLServerService       Service = "SQLServer"
	CustomSQLServerService Service = "MSSQL-Broker"
	CredHubService         Service = "credhub"
)

const (
	ManagedService      ServiceType = "managed_service_instance"
	UserProvidedService ServiceType = "user_provided_service_instance"
)

func (m Migrator) String() string {
	switch m {
	case ECS:
		return "ecs"
	case MySQL:
		return "mysql"
	case SQLServer:
		return "sqlserver"
	case CredHub:
		return "credhub"
	}
	return "unknown"
}

func (m Migrator) HasCCDBConfig() bool {
	switch m {
	case ECS:
		return true
	case MySQL:
		return false
	case SQLServer:
		return true
	case CredHub:
		return false
	}
	return false
}

func (s Service) String() string {
	return string(s)
}

func (t ServiceType) String() string {
	return string(t)
}

type MigratorHelper struct {
	migrationReader    config.MigrationReader
	supportedMigrators map[Service]Migrator
	mu                 sync.Mutex
}

func NewMigratorHelper(mr config.MigrationReader) *MigratorHelper {
	return &MigratorHelper{
		migrationReader: mr,
	}
}

func (h *MigratorHelper) GetMigratorConfig(data interface{}, si *cf.ServiceInstance) (interface{}, error) {
	cfg, err := h.migrationReader.GetMigration()
	if err != nil {
		return nil, err
	}

	var migratorType Migrator
	var ok bool
	if migratorType, ok = h.GetMigratorType(si.Service); !ok {
		return nil, fmt.Errorf("service %q of type %q not supported", si.Service, si.Type)
	}

	return config.NewMapDecoder(data).Decode(*cfg, migratorType.String()), nil
}

func (h *MigratorHelper) GetMigratorType(service string) (Migrator, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.supportedMigrators) == 0 {
		h.loadSupportedMigrators()
	}

	// return the value used as a key to look up the supported migrator config
	if m, ok := h.supportedMigrators[Service(service)]; ok {
		return m, true
	}

	return 0, false
}

func (h *MigratorHelper) GetReader() config.MigrationReader {
	return h.migrationReader
}

func (h *MigratorHelper) IsCCDBMigrator(service string) bool {
	if migratorType, ok := h.GetMigratorType(service); ok {
		return migratorType.HasCCDBConfig()
	}
	return false
}

func (h *MigratorHelper) loadSupportedMigrators() {
	h.supportedMigrators = make(map[Service]Migrator)
	h.supportedMigrators[CustomSQLServerService] = SQLServer
	h.supportedMigrators[SQLServerService] = SQLServer
	h.supportedMigrators[ECSBucketService] = ECS
	h.supportedMigrators[MySQLService] = MySQL
	h.supportedMigrators[CredHubService] = CredHub
}
