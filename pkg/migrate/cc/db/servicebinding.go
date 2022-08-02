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
	"database/sql"
	"encoding/json"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/crypto"

	"github.com/cloudfoundry-community/go-cfclient"
)

// CreateServiceBinding will inject a service binding into the DB
func (d *CloudController) CreateServiceBinding(binding cfclient.ServiceBinding, appGUID string, encryptionKey string) error {
	var (
		encryptedVolumeMounts sql.NullString
		volumeMountsSalt      sql.NullString
	)

	credentials := binding.Credentials
	if credentials == nil {
		credentials = map[string]interface{}{}
	}

	sourceCredentials, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	salt, err := d.GenerateSalt()
	if err != nil {
		return err
	}

	wrappedSalt := NewNullString(salt)

	encryptedCredentials, err := crypto.Encrypt(string(sourceCredentials), salt, encryptionKey)
	if err != nil {
		return err
	}

	if binding.VolumeMounts != nil {
		volMountsJSON, merr := json.Marshal(binding.VolumeMounts)
		if merr != nil {
			return merr
		}

		encryptedMounts, crypterr := crypto.Encrypt(string(volMountsJSON), salt, encryptionKey)
		if crypterr != nil {
			return crypterr
		}

		encryptedVolumeMounts = NewNullString(encryptedMounts)
		volumeMountsSalt = wrappedSalt
	}

	row := &serviceBindingRow{
		AppGUID:             appGUID,
		Credentials:         encryptedCredentials,
		GUID:                binding.Guid,
		Salt:                wrappedSalt,
		ServiceInstanceGUID: binding.ServiceInstanceGuid,
		SyslogDrainURL:      NewNullString(binding.SyslogDrainUrl),
		Type:                NewNullString("app"),
		VolumeMounts:        encryptedVolumeMounts,
		VolumeMountsSalt:    volumeMountsSalt,
	}

	_, err = d.DB.NamedExec(CreateServiceBindingSQLStatement, row)
	return err
}
