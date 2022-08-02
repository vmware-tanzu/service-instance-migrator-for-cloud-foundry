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

package db

const (
	CreateServiceInstanceSQLStatement           = `INSERT INTO service_instances (guid, name, credentials, gateway_name, gateway_data, space_id, service_plan_id, salt, dashboard_url, is_gateway_service, syslog_drain_url, tags, route_service_url) VALUES (:guid, :name, :credentials, :gateway_name, :gateway_data, :space_id, :service_plan_id, :salt, :dashboard_url, :is_gateway_service, :syslog_drain_url, :tags, :route_service_url);`
	DeleteServiceInstanceSQLStatement           = `DELETE FROM service_instances WHERE guid=? AND space_id=?`
	CreateServiceBindingSQLStatement            = `INSERT INTO service_bindings (guid, credentials, salt, syslog_drain_url, volume_mounts, volume_mounts_salt, app_guid, service_instance_guid, type) VALUES (:guid, :credentials, :salt, :syslog_drain_url, :volume_mounts, :volume_mounts_salt, :app_guid, :service_instance_guid, :type)`
	DeleteServiceBindingsSQLStatement           = `DELETE FROM service_bindings WHERE service_instance_guid=?`
	DeleteServiceKeysSQLStatement               = `DELETE FROM service_keys WHERE service_instance_id=?`
	DeleteServiceInstanceOperationsSQLStatement = `DELETE FROM service_instance_operations WHERE service_instance_id=?`
	CreateServiceUsageEventSQLStatement         = `INSERT INTO service_usage_events (guid, state, org_guid, space_guid, space_name, service_instance_guid, service_instance_name, service_instance_type, service_plan_guid, service_plan_name, service_guid, service_label) VALUES (:guid, :state, :org_guid, :space_guid, :space_name, :service_instance_guid, :service_instance_name, :service_instance_type, :service_plan_guid, :service_plan_name, :service_guid, :service_label);`
	GetIDFromGUIDLimitToOneTemplateQuery        = `SELECT id FROM %s WHERE guid=? LIMIT 1`
	GetIDFromGUIDTemplateQuery                  = `SELECT id FROM %s WHERE guid=?`
	GetServiceInstanceSharesQuery               = `SELECT service_instance_guid, target_space_guid FROM service_instance_shares WHERE service_instance_guid=? AND target_space_guid=?`
)
