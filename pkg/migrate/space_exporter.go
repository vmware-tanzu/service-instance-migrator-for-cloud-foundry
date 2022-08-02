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
	"errors"
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"

	"github.com/cloudfoundry-community/go-cfclient"
	"golang.org/x/sync/errgroup"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type SpaceExporter struct {
	ClientHolder ClientHolder
	ServiceInstanceExporter
}

func NewSpaceExporter(e ServiceInstanceExporter, h ClientHolder) *SpaceExporter {
	return &SpaceExporter{
		ClientHolder:            h,
		ServiceInstanceExporter: e,
	}
}

func (e SpaceExporter) Export(ctx context.Context, om config.OpsManager, dir string, orgName, spaceName string) error {
	g, gctx := errgroup.WithContext(ctx)
	client := e.ClientHolder.SourceCFClient()

	org, err := client.GetOrgByName(orgName)
	if err != nil {
		return fmt.Errorf("failed to find org %q: %w", orgName, err)
	}

	space, err := client.GetSpaceByName(spaceName, org.Guid)
	if err != nil {
		return fmt.Errorf("failed to find space %q: %w", spaceName, err)
	}

	func(o cfclient.Org, s cfclient.Space) {
		g.Go(func() error {
			log.Infof("Exporting %s/%s managed services", o.Name, s.Name)
			return e.ExportManagedServices(gctx, o, s, om, dir)
		})
	}(org, space)

	func(o cfclient.Org, s cfclient.Space) {
		g.Go(func() error {
			log.Infof("Exporting %s/%s user provided services", o.Name, s.Name)
			return e.ExportUserProvidedServices(gctx, o, s, dir)
		})
	}(org, space)

	log.Debugln("Waiting for export to finish...")
	err = g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Warnln("context was canceled")
		}
		return err
	}
	log.Infoln("Finished exporting services")
	return nil
}
