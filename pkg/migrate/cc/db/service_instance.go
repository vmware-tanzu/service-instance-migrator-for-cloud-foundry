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

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/crypto"
)

var ErrUnsupportedOperation = errors.New("unsupported operation")

// CreateServiceInstance will insert a service instance (and associated service usage event)
// in the CC DB
func (d *CloudController) CreateServiceInstance(si cfclient.ServiceInstance, targetSpace cfclient.Space, targetPlan cfclient.ServicePlan, targetService cfclient.Service, key string) error {

	var spaceRows []idResponse
	var servicePlanRows []idResponse

	if err := d.DB.Select(&spaceRows, fmt.Sprintf(GetIDFromGUIDLimitToOneTemplateQuery, "spaces"), targetSpace.Guid); err != nil || len(spaceRows) == 0 {
		if err == nil {
			err = fmt.Errorf("could not find space with id %s", targetSpace.Guid)
		}
		return err
	}

	if err := d.DB.Select(&servicePlanRows, fmt.Sprintf(GetIDFromGUIDLimitToOneTemplateQuery, "service_plans"), targetPlan.Guid); err != nil || len(servicePlanRows) == 0 {
		if err == nil {
			err = fmt.Errorf("could not find service plan with id %s", targetPlan.Guid)
		}
		return err
	}

	spaceID := spaceRows[0].ID
	planID := servicePlanRows[0].ID

	tx, err := d.DB.Beginx()
	if err != nil {
		return err
	}

	salt, err := d.GenerateSalt()
	if err != nil {
		return err
	}

	serializedCreds := ""
	if si.Credentials != nil {
		c, marshalErr := json.Marshal(si.Credentials)
		if marshalErr != nil {
			return marshalErr
		}
		serializedCreds = string(c)
	}

	serializedTags := ""
	if si.Tags != nil {
		c, marshalErr := json.Marshal(si.Tags)
		if marshalErr != nil {
			return marshalErr
		}

		serializedTags = string(c)
	}

	encrypted, err := crypto.Encrypt(serializedCreds, salt, key)
	if err != nil {
		return err
	}

	siRow := &serviceInstanceRow{
		GUID:             si.Guid,
		Name:             si.Name,
		Credentials:      NewNullString(encrypted),
		IsGatewayService: true,
		GatewayData:      NewNullString(""),
		GatewayName:      NewNullString(""),
		SpaceID:          spaceID,
		ServicePlanID:    NewNullInt64(int64(planID)),
		Salt:             NewNullString(salt),
		DashboardURL:     NewNullString(si.DashboardUrl),
		SyslogDrainURL:   NewNullString(""),
		Tags:             NewNullString(serializedTags),
		RouteServiceURL:  NewNullString(""),
	}

	_, err = tx.NamedExec(CreateServiceInstanceSQLStatement, siRow)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return err
	}

	guid := uuid.NewV4()

	eventRow := &serviceUsageEventRow{
		GUID:                guid.String(),
		State:               "CREATED",
		OrganizationGUID:    targetSpace.OrganizationGuid,
		SpaceGUID:           targetSpace.Guid,
		SpaceName:           targetSpace.Name,
		ServiceInstanceGUID: si.Guid,
		ServiceInstanceName: si.Name,
		ServiceInstanceType: "managed_service_instance",
		ServicePlanGUID:     NewNullString(targetPlan.Guid),
		ServicePlanName:     NewNullString(targetPlan.Name),
		ServiceGUID:         NewNullString(targetService.Guid),
		ServiceLabel:        NewNullString(targetService.Label),
	}

	_, err = tx.NamedExec(CreateServiceUsageEventSQLStatement, eventRow)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return err
	}

	return nil
}

func (d *CloudController) ServiceInstanceExists(serviceInstanceGUID string) (bool, error) {
	var rows []idResponse
	if err := d.DB.Select(&rows, fmt.Sprintf(GetIDFromGUIDLimitToOneTemplateQuery, "service_instances"), serviceInstanceGUID); err != nil || len(rows) == 0 {
		if err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("could not find service_instance with id %s", serviceInstanceGUID))
		}
		return false, nil
	}
	return true, nil
}

func (d *CloudController) DeleteServiceInstance(spaceGUID string, serviceInstanceGUID string) (bool, error) {
	tx, err := d.DB.Beginx()
	if err != nil {
		return false, err
	}

	var sharedInstancesRows []sharesResponse
	if err := d.DB.Select(&sharedInstancesRows, GetServiceInstanceSharesQuery, serviceInstanceGUID, spaceGUID); err != nil {
		return false, err
	}

	if len(sharedInstancesRows) > 0 {
		return false, fmt.Errorf("shared service instances cannot be deleted: %w", ErrUnsupportedOperation)
	}

	var spaceRows []idResponse
	var serviceInstanceIDs []idResponse

	if err := d.DB.Select(&spaceRows, fmt.Sprintf(GetIDFromGUIDLimitToOneTemplateQuery, "spaces"), spaceGUID); err != nil || len(spaceRows) == 0 {
		if err == nil {
			err = fmt.Errorf("could not find space with id %s", spaceGUID)
		}
		return false, err
	}

	if err := d.DB.Select(&serviceInstanceIDs, fmt.Sprintf(GetIDFromGUIDTemplateQuery, "service_instances"), serviceInstanceGUID); err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("could not find service instance with guid %s", serviceInstanceGUID))
	}

	if len(serviceInstanceIDs) == 0 {
		return false, fmt.Errorf("could not find service instance with guid %s", serviceInstanceGUID)
	}

	if len(serviceInstanceIDs) > 1 {
		panic("found more than one service instance")
	}

	err = d.deleteServiceBindings(tx, serviceInstanceGUID)
	if err != nil {
		return false, err
	}

	err = d.deleteServiceKeys(tx, serviceInstanceIDs[0].ID)
	if err != nil {
		return false, err
	}

	log.Debugf("Deleting service instance operations with instance id %d", serviceInstanceIDs[0].ID)
	res, err := tx.Exec(DeleteServiceInstanceOperationsSQLStatement, serviceInstanceIDs[0].ID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return false, fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return false, err
	}
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			if count == 0 {
				log.Debugf("no rows were deleted for query %q", DeleteServiceInstanceOperationsSQLStatement)
			}
		}
	}

	log.Debugf("Deleting service instance: %q in space id: %d", serviceInstanceGUID, spaceRows[0].ID)
	res, err = tx.Exec(DeleteServiceInstanceSQLStatement, serviceInstanceGUID, spaceRows[0].ID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return false, fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return false, err
	}
	if err == nil {
		count, err := res.RowsAffected()
		if err == nil {
			if count == 0 {
				log.Debugf("no rows were deleted for query %q", DeleteServiceInstanceSQLStatement)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Infoln("Transaction failed")
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return false, fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return false, err
	}

	return true, err
}

func (d *CloudController) deleteServiceBindings(tx *sqlx.Tx, serviceInstanceGUID string) error {
	log.Debugf("Deleting service binding for instance %q in ccdb...", serviceInstanceGUID)
	res, err := tx.Exec(DeleteServiceBindingsSQLStatement, serviceInstanceGUID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return err
	}

	count, err := res.RowsAffected()
	if err == nil {
		if count == 0 {
			log.Debugf("no rows were deleted for query %q", DeleteServiceBindingsSQLStatement)
		}
	}

	return err
}

func (d *CloudController) deleteServiceKeys(tx *sqlx.Tx, serviceInstanceID int) error {
	log.Debugf("Deleting service keys for instance %d in ccdb...", serviceInstanceID)
	res, err := tx.Exec(DeleteServiceKeysSQLStatement, serviceInstanceID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v: %w", err, rollbackErr)
		}

		return err
	}

	count, err := res.RowsAffected()
	if err == nil {
		if count == 0 {
			log.Debugf("no rows were deleted for query %q", DeleteServiceKeysSQLStatement)
		}
	}

	return err
}
