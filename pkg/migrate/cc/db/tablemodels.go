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

import "database/sql"

type serviceInstanceRow struct {
	GUID             string         `db:"guid"`
	Name             string         `db:"name"`
	Credentials      sql.NullString `db:"credentials"`
	GatewayName      sql.NullString `db:"gateway_name"`
	GatewayData      sql.NullString `db:"gateway_data"`
	SpaceID          int            `db:"space_id"`
	ServicePlanID    sql.NullInt64  `db:"service_plan_id"`
	Salt             sql.NullString `db:"salt"`
	DashboardURL     sql.NullString `db:"dashboard_url"`
	IsGatewayService bool           `db:"is_gateway_service"`
	SyslogDrainURL   sql.NullString `db:"syslog_drain_url"`
	Tags             sql.NullString `db:"tags"`
	RouteServiceURL  sql.NullString `db:"route_service_url"`
}

type serviceUsageEventRow struct {
	GUID                string         `db:"guid"`
	State               string         `db:"state"`
	OrganizationGUID    string         `db:"org_guid"`
	SpaceGUID           string         `db:"space_guid"`
	SpaceName           string         `db:"space_name"`
	ServiceInstanceGUID string         `db:"service_instance_guid"`
	ServiceInstanceName string         `db:"service_instance_name"`
	ServiceInstanceType string         `db:"service_instance_type"`
	ServicePlanGUID     sql.NullString `db:"service_plan_guid"`
	ServicePlanName     sql.NullString `db:"service_plan_name"`
	ServiceGUID         sql.NullString `db:"service_guid"`
	ServiceLabel        sql.NullString `db:"service_label"`
}

type serviceBindingRow struct {
	GUID                string         `db:"guid"`
	Credentials         string         `db:"credentials"`
	Salt                sql.NullString `db:"salt"`
	SyslogDrainURL      sql.NullString `db:"syslog_drain_url"`
	VolumeMounts        sql.NullString `db:"volume_mounts"`
	VolumeMountsSalt    sql.NullString `db:"volume_mounts_salt"`
	AppGUID             string         `db:"app_guid"`
	ServiceInstanceGUID string         `db:"service_instance_guid"`
	Type                sql.NullString `db:"type"`
}

type idResponse struct {
	ID int `db:"id"`
}

type sharesResponse struct {
	ServiceInstanceGUID string `db:"service_instance_guid"`
	SpaceGUID           string `db:"target_space_guid"`
}
