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
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type UserProvidedServiceMigrator struct {
	sequence flow.Flow
}

func NewUserProvidedServiceMigrator(flow flow.Flow) UserProvidedServiceMigrator {
	return UserProvidedServiceMigrator{
		sequence: flow,
	}
}

func (m UserProvidedServiceMigrator) Migrate(ctx context.Context) (*cf.ServiceInstance, error) {
	dryRun := false
	if cfg, ok := config.FromContext(ctx); ok {
		dryRun = cfg.DryRun
	}

	var res flow.Result
	var err error
	if res, err = m.sequence.Run(ctx, nil, dryRun); err != nil {
		return nil, err
	}
	if si, ok := res.(*cf.ServiceInstance); ok {
		return si, nil
	}

	return nil, err
}

func (m UserProvidedServiceMigrator) Validate(si *cf.ServiceInstance, export bool) error {
	return nil
}

func NewUserProvidedServiceFlow(org, space string, h ClientHolder, instance *cf.ServiceInstance, isExport bool) flow.Flow {
	return flow.ProgressBarSequence(
		fmt.Sprintf("Migrating %s", instance.Name),
		flow.StepWithProgressBar(
			CreateUserProvidedService(org, space, instance, h, isExport),
			flow.WithDisplay("Creating user provided service"),
		),
	)
}

func CreateUserProvidedService(orgName, spaceName string, instance *cf.ServiceInstance, holder ClientHolder, isExport bool) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		if isExport {
			return nil, nil
		}

		client := holder.CFClient(isExport)

		org, err := client.GetOrgByName(orgName)
		if err != nil {
			return nil, fmt.Errorf("could not find org %q, %w", orgName, err)
		}

		space, err := client.GetSpaceByName(spaceName, org.Guid)
		if err != nil {
			return nil, fmt.Errorf("could not find find space %q in org %q, %w", spaceName, orgName, err)
		}

		var domainsToReplace map[string]string
		if cfg, ok := config.FromContext(ctx); ok {
			domainsToReplace = cfg.DomainsToReplace
		}

		log.Debugf("Domains to replace are %v", domainsToReplace)

		if instance.SyslogDrainUrl != "" {
			instance.SyslogDrainUrl = ReplaceDomain(instance.SyslogDrainUrl, domainsToReplace)
			log.Debugf("Set syslog drain url to %s", instance.SyslogDrainUrl)
		}

		if instance.RouteServiceUrl != "" {
			instance.RouteServiceUrl = ReplaceDomain(instance.RouteServiceUrl, domainsToReplace)
			log.Debugf("Set route service url to %s", instance.RouteServiceUrl)
		}

		creds := map[string]interface{}{}
		if len(instance.Credentials) > 0 {
			for k, v := range instance.Credentials {
				creds[k] = v
				if val, ok := v.(string); ok {
					creds[k] = ReplaceDomain(val, domainsToReplace)
				}
			}
			instance.Credentials = creds
			log.Debugf("Set creds to %v", instance.Credentials)
		}

		ups, err := client.ListUserProvidedServiceInstancesByQuery(url.Values{"q": []string{
			fmt.Sprintf("organization_guid:%s", org.Guid),
			fmt.Sprintf("space_guid:%s", space.Guid),
			fmt.Sprintf("name:%s", instance.Name),
		}})
		if err != nil {
			log.Errorf("Error looking up user provided services by name: %q in %s/%s, %v", instance.Name, orgName, spaceName, err)
			ups = []cfclient.UserProvidedServiceInstance{}
		}

		if !dryRun && len(ups) > 0 {
			log.Debugf("Updating user-provided-service %+v", instance)
			_, err = client.UpdateUserProvidedServiceInstance(ups[0].Guid, cfclient.UserProvidedServiceInstanceRequest{
				Name:            instance.Name,
				SpaceGuid:       space.Guid,
				Credentials:     instance.Credentials,
				Tags:            strings.Split(instance.Tags, ","),
				SyslogDrainUrl:  instance.SyslogDrainUrl,
				RouteServiceUrl: instance.RouteServiceUrl,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update user-provided-service: %q in %s/%s, %w", instance.Name, orgName, spaceName, err)
			}
		}

		if len(ups) == 0 {
			log.Debugf("Creating user-provided-service %+v", instance)
			_, err = client.CreateUserProvidedServiceInstance(cfclient.UserProvidedServiceInstanceRequest{
				Name:            instance.Name,
				SpaceGuid:       space.Guid,
				Credentials:     instance.Credentials,
				Tags:            strings.Split(instance.Tags, ","),
				SyslogDrainUrl:  instance.SyslogDrainUrl,
				RouteServiceUrl: instance.RouteServiceUrl,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create user-provided-service: %q in %s/%s, %w", instance.Name, orgName, spaceName, err)
			}
		}

		return instance, err
	}
}

func ReplaceDomain(val string, domainsToReplace map[string]string) string {
	for oldDomain, newDomain := range domainsToReplace {
		if strings.Contains(val, oldDomain) {
			return strings.ReplaceAll(val, oldDomain, newDomain)
		}
	}
	return val
}
