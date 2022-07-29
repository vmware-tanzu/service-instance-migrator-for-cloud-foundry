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
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
)

type DefaultCloudControllerService struct {
	Client           cf.Client
	Database         db.Repository
	ManifestExporter ManifestExporter
}

func NewCloudControllerService(db db.Repository, cf cf.Client, manifestExporter ManifestExporter) DefaultCloudControllerService {
	return DefaultCloudControllerService{
		Client:           cf,
		Database:         db,
		ManifestExporter: manifestExporter,
	}
}

func (m DefaultCloudControllerService) DownloadManifest(org, space, appName string) (cf.Application, error) {
	var app cfclient.App

	targetOrg, err := m.Client.GetOrgByName(org)
	if err != nil {
		return cf.Application{}, errors.Wrap(err, fmt.Sprintf("could not find org %q", org))
	}

	targetSpace, err := m.Client.GetSpaceByName(space, targetOrg.Guid)
	if err != nil {
		return cf.Application{}, errors.Wrap(err, fmt.Sprintf("could not find space %q in org %q", space, targetOrg.Name))
	}

	app, err = m.Client.AppByName(appName, targetSpace.Guid, targetOrg.Guid)
	if err != nil {
		return cf.Application{}, fmt.Errorf("failed to find app: %s in org: %s, space: %s, %w", appName, org, space, err)
	}

	return m.ManifestExporter.ExportAppManifest(targetOrg, targetSpace, app)
}

func (m DefaultCloudControllerService) Delete(org, space string, instance *cf.ServiceInstance) error {
	targetOrg, err := m.Client.GetOrgByName(org)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not find org %q", org))
	}

	targetSpace, err := m.Client.GetSpaceByName(space, targetOrg.Guid)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not find space %q in org %q", space, targetOrg.Name))
	}

	isDeleted, err := m.Database.DeleteServiceInstance(targetSpace.Guid, instance.GUID)
	if err != nil {
		return err
	}

	if !isDeleted {
		return fmt.Errorf("failed to delete service instance %s", instance.GUID)
	}

	return nil
}

func (m DefaultCloudControllerService) Create(org, space string, instance *cf.ServiceInstance, encryptionKey string) error {
	targetOrg, err := m.Client.GetOrgByName(org)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not find org %q", org))
	}

	targetSpace, err := m.Client.GetSpaceByName(space, targetOrg.Guid)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not find space %q in org %q", space, targetOrg.Name))
	}

	targetPlans, err := m.Client.ListServicePlans()
	if err != nil {
		return errors.Wrap(err, "could not get service plans from target foundation")
	}

	var targetPlan cfclient.ServicePlan
	var foundPlan bool
	for _, plan := range targetPlans {
		if plan.Name == instance.Plan {
			targetPlan = plan
			foundPlan = true
			break
		}
	}
	if !foundPlan {
		return fmt.Errorf("failed to find a plan %s for service instance %s", instance.Plan, instance.GUID)
	}

	targetServices, err := m.Client.ListServices()
	if err != nil {
		return errors.Wrap(err, "could not get service plans from target foundation")
	}

	var targetService cfclient.Service
	var foundService bool
	for _, service := range targetServices {
		if service.Label == instance.Service {
			targetService = service
			foundService = true
			break
		}
	}
	if !foundService {
		return fmt.Errorf("failed to find %s service for service instance %q with guid %q", instance.Service, instance.Name, instance.GUID)
	}

	si := cfclient.ServiceInstance{
		Name:          instance.Name,
		Credentials:   instance.Credentials,
		DashboardUrl:  instance.DashboardURL,
		LastOperation: cfclient.LastOperation{},
		Tags:          strings.Split(instance.Tags, ","),
		Guid:          instance.GUID,
	}

	exists, err := m.Database.ServiceInstanceExists(si.Guid)
	if err != nil {
		return err
	}

	if !exists {
		err = m.Database.CreateServiceInstance(si, targetSpace, targetPlan, targetService, encryptionKey)
		if err != nil {
			return errors.Wrap(err, "failed to create service instance")
		}
	}

	return nil
}

func (m DefaultCloudControllerService) CreateServiceKey(si cf.ServiceInstance, key cf.ServiceKey) error {
	_, err := m.Client.CreateServiceKey(cfclient.CreateServiceKeyRequest{
		Name:                key.Name,
		ServiceInstanceGuid: key.ServiceInstanceGuid,
		Parameters:          si.Params,
	})

	return err
}

func (m DefaultCloudControllerService) CreateServiceBinding(binding *cf.ServiceBinding, appGUID string, encryptionKey string) error {
	b := cfclient.ServiceBinding{
		Guid:                binding.Guid,
		Name:                binding.Name,
		AppGuid:             binding.AppGuid,
		ServiceInstanceGuid: binding.ServiceInstanceGuid,
		Credentials:         binding.Credentials,
		BindingOptions:      binding.BindingOptions,
		GatewayData:         binding.GatewayData,
		GatewayName:         binding.GatewayName,
		SyslogDrainUrl:      binding.SyslogDrainUrl,
		VolumeMounts:        binding.VolumeMounts,
		AppUrl:              binding.AppUrl,
		ServiceInstanceUrl:  binding.ServiceInstanceUrl,
	}
	err := m.Database.CreateServiceBinding(b, appGUID, encryptionKey)
	if err != nil {
		return errors.Wrapf(err, "failed to create service binding %q", b.Name)
	}

	return nil
}

func (m DefaultCloudControllerService) CreateApp(org, space, name string) (string, error) {
	targetOrg, err := m.Client.GetOrgByName(org)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not find org %q", org))
	}

	targetSpace, err := m.Client.GetSpaceByName(space, targetOrg.Guid)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not find space %q in org %q", space, targetOrg.Name))
	}

	// lookup app in case it's already been pushed
	a, _ := m.Client.AppByName(name, targetSpace.Guid, targetOrg.Guid)
	if a.Guid != "" {
		return a.Guid, nil
	}

	app, err := m.Client.CreateApp(cfclient.AppCreateRequest{
		Name:      name,
		SpaceGuid: targetSpace.Guid,
		State:     cfclient.APP_STOPPED,
	})

	return app.Guid, err
}

func (m DefaultCloudControllerService) FindAppByGUID(guid string) (string, error) {
	app, err := m.Client.GetAppByGuidNoInlineCall(guid)
	return app.Name, err
}

func (m DefaultCloudControllerService) ServiceInstanceExists(org, space, instanceName string) (bool, error) {
	targetOrg, err := m.Client.GetOrgByName(org)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("could not find org %q", org))
	}

	targetSpace, err := m.Client.GetSpaceByName(space, targetOrg.Guid)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("could not find space %q in org %q", space, targetOrg.Name))
	}

	sis, err := m.Client.ListServiceInstancesByQuery(url.Values{"q": []string{
		fmt.Sprintf("organization_guid:%s", targetOrg.Guid),
		fmt.Sprintf("space_guid:%s", targetSpace.Guid),
		fmt.Sprintf("name:%s", instanceName),
	}})
	if err != nil {
		return false, err
	}

	if len(sis) > 0 {
		return true, nil
	}

	return false, nil
}
