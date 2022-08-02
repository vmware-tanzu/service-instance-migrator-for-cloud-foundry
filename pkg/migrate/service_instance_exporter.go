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

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/validation"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type DefaultServiceInstanceExporter struct {
	ClientHolder ClientHolder
	Registry     MigratorRegistry
	Parser       ServiceInstanceParser
	cfg          *config.Config
}

func NewServiceInstanceExporter(cfg *config.Config, h ClientHolder, registry MigratorRegistry, parser ServiceInstanceParser) *DefaultServiceInstanceExporter {
	return &DefaultServiceInstanceExporter{
		ClientHolder: h,
		Registry:     registry,
		Parser:       parser,
		cfg:          cfg,
	}
}

func (e *DefaultServiceInstanceExporter) ExportManagedServices(ctx context.Context, org cfclient.Org, space cfclient.Space, om config.OpsManager, dir string) error {
	g, gctx := errgroup.WithContext(ctx)
	client := e.ClientHolder.SourceCFClient()

	// variable behavior using the service instances api -- does not always return the correct service instances
	// similar issue: https://github.com/cloudfoundry/cloud_controller_ng/issues/2474
	// changed to using the spaces api to fetch the service instances by space
	instances, err := client.ListSpaceServiceInstances(space.Guid)
	if err != nil {
		return fmt.Errorf("error getting managed service instances for '%s/%s': %w", org.Name, space.Name, err)
	}

	if len(instances) == 0 {
		log.Warnf("No service instances found in org: %q, space: %q", org.Name, space.Name)
		return nil
	}

	for _, instance := range instances {
		func(instance cfclient.ServiceInstance) {
			g.Go(func() error {
				if !e.shouldMigrate(instance) {
					return nil
				}

				bindings, err := client.ListServiceBindingsByQuery(url.Values{"q": []string{fmt.Sprintf("service_instance_guid:%s", instance.Guid)}})
				if err != nil {
					return fmt.Errorf("could not fetch service bindings for instance %s: %w", instance.Guid, err)
				}

				serviceKeys, err := client.ListServiceKeysByQuery(url.Values{"q": []string{fmt.Sprintf("service_instance_guid:%s", instance.Guid)}})
				if err != nil {
					return fmt.Errorf("could not fetch service keys for instance %s: %w", instance.Guid, err)
				}

				params, err := client.GetServiceInstanceParams(instance.Guid)
				if err != nil && !cfclient.IsServiceFetchInstanceParametersNotSupportedError(err) {
					return fmt.Errorf("could not fetch service instance params for instance %s: %w", instance.Name, err)
				}

				servicePlan, err := client.GetServicePlanByGUID(instance.ServicePlanGuid)
				if err != nil {
					return fmt.Errorf("error retrieving service plan for instance %s: %w", instance.Name, err)
				}

				svc, err := client.GetServiceByGuid(servicePlan.ServiceGuid)
				if err != nil {
					return fmt.Errorf("error retrieving service for plan %s: %w", servicePlan.Name, err)
				}

				si := &cf.ServiceInstance{
					Name:            instance.Name,
					GUID:            instance.Guid,
					Type:            instance.Type,
					Tags:            strings.Join(instance.Tags, ","),
					Params:          params,
					Plan:            servicePlan.Name,
					Credentials:     instance.Credentials,
					ServiceBindings: convertBindings(bindings),
					ServiceKeys:     convertServiceKeys(serviceKeys),
					Service:         svc.Label,
				}

				migrator, migrate, err := e.Registry.Lookup(org.Name, space.Name, si, om, dir, true)

				if err != nil {
					return fmt.Errorf("failed to find a valid migrator for instance %s: %w", si.Name, err)
				}

				dryRun := false
				if cfg, ok := config.FromContext(ctx); ok {
					dryRun = cfg.DryRun
				}

				if !migrate || dryRun {
					if summary, ok := config.SummaryFromContext(gctx); ok {
						summary.AddSkippedService(org.Name, space.Name, si.Name, si.Service, nil)
					}
					return nil
				}

				if migrator == nil {
					if summary, ok := config.SummaryFromContext(gctx); ok {
						summary.AddSkippedService(org.Name, space.Name, si.Name, si.Service, nil)
					}
					return nil
				}

				err = migrator.Validate(si, true)
				if err != nil {
					return err
				}

				log.Infof("Exporting service %s from %s/%s", si.Service, org.Name, space.Name)

				migrated, err := migrator.Migrate(gctx)
				if errors.Is(err, db.ErrUnsupportedOperation) {
					log.Warnf("unsupported db error %v, skipped exporting service instance %s", err, si.Name)
					if summary, ok := config.SummaryFromContext(gctx); ok {
						summary.AddSkippedService(org.Name, space.Name, si.Name, si.Service, err)
					}
					return nil
				}

				var validationErr *validation.MigrationError
				if errors.As(err, &validationErr) && len(si.ServiceBindings) == 0 {
					log.Warnf("validation error %s, skipped exporting service instance %s", validationErr.Error(), si.Name)
					if summary, ok := config.SummaryFromContext(gctx); ok {
						summary.AddSkippedService(org.Name, space.Name, si.Name, si.Service, validationErr)
					}
					return nil
				}

				if migrated != nil {
					for _, app := range migrated.AppManifest.Applications {
						filename := strings.ReplaceAll(app.Name, "/", "-") + "_manifest"
						err = marshalAppManifest(e.Parser, filename, &migrated.AppManifest, org, space, dir)
						if err != nil {
							log.Errorf("Failed to save manifest: %s, error: %v", filename, err)
						}
					}
				}

				if err != nil {
					if summary, ok := config.SummaryFromContext(gctx); ok {
						summary.AddFailedService(org.Name, space.Name, si.Name, si.Service, err)
					}
					return fmt.Errorf("failed to migrate %s: %w", si.Name, err)
				}

				err = marshalServiceInstance(e.Parser, si, org, space, dir)
				if err != nil {
					log.Fatalf("cannot save instance %v, error %v", si, err)
				}

				log.Debugf("Finished exporting %q", si.Name)
				if summary, ok := config.SummaryFromContext(gctx); ok {
					summary.AddSuccessfulService(org.Name, space.Name, si.Name, si.Service)
				}

				return nil
			})
		}(instance)
	}
	gerr := g.Wait()
	if gerr != nil {
		return gerr
	}

	return nil
}

func (e *DefaultServiceInstanceExporter) ExportUserProvidedServices(ctx context.Context, org cfclient.Org, space cfclient.Space, dir string) error {
	client := e.ClientHolder.SourceCFClient()

	spaceGuidParams := url.Values{"q": []string{fmt.Sprintf("space_guid:%s", space.Guid)}}
	upsInstances, err := client.ListUserProvidedServiceInstancesByQuery(spaceGuidParams)
	if err != nil {
		return fmt.Errorf("error getting user provided service instances for space %s instance org %s: %w", space.Name, org.Name, err)
	}

	for _, ups := range upsInstances {
		si := &cf.ServiceInstance{
			Name:            ups.Name,
			GUID:            ups.Guid,
			Type:            ups.Type,
			Tags:            strings.Join(ups.Tags, ","),
			SyslogDrainUrl:  ups.SyslogDrainUrl,
			RouteServiceUrl: ups.RouteServiceUrl,
			Credentials:     ups.Credentials,
			Service:         ups.Name,
		}
		fd := io.FileDescriptor{
			BaseDir:   dir,
			Name:      ups.Name,
			Org:       org.Name,
			Space:     space.Name,
			Extension: "yml",
		}
		err = e.Parser.Marshal(si, fd)
		if err != nil {
			if summary, ok := config.SummaryFromContext(ctx); ok {
				summary.AddFailedService(org.Name, space.Name, si.Name, si.Service, err)
			}
			log.Warnf("cannot save instance %q, error %s", si.Name, err)
			return err
		}

		if summary, ok := config.SummaryFromContext(ctx); ok {
			summary.AddSuccessfulService(org.Name, space.Name, si.Name, si.Service)
		}
	}
	return nil
}

func (e *DefaultServiceInstanceExporter) shouldMigrate(instance cfclient.ServiceInstance) bool {
	if len(e.cfg.Instances) == 0 {
		return true
	}

	for _, inst := range e.cfg.Instances {
		if inst == instance.Name {
			return true
		}
	}

	return false
}

func convertBindings(bindings []cfclient.ServiceBinding) []cf.ServiceBinding {
	configBindings := make([]cf.ServiceBinding, 0, len(bindings))
	for _, b := range bindings {
		configBindings = append(configBindings, convertBinding(b))
	}
	return configBindings
}

func convertBinding(binding cfclient.ServiceBinding) cf.ServiceBinding {
	return cf.ServiceBinding{
		Guid:                binding.Guid,
		Name:                binding.Name,
		AppGuid:             binding.AppGuid,
		ServiceInstanceGuid: binding.ServiceInstanceGuid,
		Credentials:         binding.Credentials.(map[string]interface{}),
		BindingOptions:      binding.BindingOptions,
		GatewayData:         binding.GatewayData,
		GatewayName:         binding.GatewayName,
		SyslogDrainUrl:      binding.SyslogDrainUrl,
		VolumeMounts:        binding.VolumeMounts,
	}
}

func convertServiceKeys(keys []cfclient.ServiceKey) []cf.ServiceKey {
	configKeys := make([]cf.ServiceKey, 0, len(keys))
	for _, k := range keys {
		configKeys = append(configKeys, convertServiceKey(k))
	}
	return configKeys
}

func convertServiceKey(key cfclient.ServiceKey) cf.ServiceKey {
	return cf.ServiceKey{
		Guid:                key.Guid,
		Name:                key.Name,
		ServiceInstanceGuid: key.ServiceInstanceGuid,
		Credentials:         key.Credentials.(map[string]interface{}),
		ServiceInstanceUrl:  key.ServiceInstanceUrl,
	}
}

func marshalServiceInstance(parser ServiceInstanceParser, si *cf.ServiceInstance, org cfclient.Org, space cfclient.Space, dir string) error {
	fd := io.FileDescriptor{
		BaseDir:   dir,
		Name:      si.Name,
		Org:       org.Name,
		Space:     space.Name,
		Extension: "yml",
	}

	return parser.Marshal(si, fd)
}

func marshalAppManifest(parser ServiceInstanceParser, filename string, m *cf.Manifest, org cfclient.Org, space cfclient.Space, dir string) error {
	fd := io.FileDescriptor{
		BaseDir:   dir,
		Name:      filename,
		Org:       org.Name,
		Space:     space.Name,
		Extension: "yml",
	}

	return parser.Marshal(m, fd)
}
