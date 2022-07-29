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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"net/url"
	"regexp"
	"strconv"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type OrgExporter struct {
	ClientHolder            ClientHolder
	ExcludedOrgs            []string
	IncludedOrgs            []string
	ServiceInstanceExporter ServiceInstanceExporter
}

func NewOrgExporter(e ServiceInstanceExporter, h ClientHolder, includedOrgs []string, excludedOrgs []string) *OrgExporter {
	return &OrgExporter{
		ServiceInstanceExporter: e,
		ClientHolder:            h,
		IncludedOrgs:            includedOrgs,
		ExcludedOrgs:            excludedOrgs,
	}
}

func (e OrgExporter) Export(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
	g, gctx := errgroup.WithContext(ctx)

	for _, o := range orgs {

		if e.ExcludeOrg(o) || !e.IncludeOrg(o) {
			log.Debugf("Excluding org %q", o)
			continue
		}

		log.Infof("Exporting org %q", o)
		client := e.ClientHolder.SourceCFClient()
		org, err := client.GetOrgByName(o)
		if err != nil {
			return err
		}

		page := 1
		resultsPerPage := 50
		for {
			params := url.Values{
				"page":             []string{strconv.Itoa(page)},
				"results-per-page": []string{strconv.Itoa(resultsPerPage)},
				"q":                []string{"organization_guid:" + org.Guid},
			}
			var err error
			var spaces []cfclient.Space
			spaces, err = client.ListSpacesByQuery(params)
			if err != nil {
				return err
			}
			for _, space := range spaces {
				log.Infof("Exporting space %q", space.Name)
				func(o cfclient.Org, s cfclient.Space) {
					g.Go(func() error {
						log.Infof("Exporting %s/%s managed services", o.Name, s.Name)
						return e.ServiceInstanceExporter.ExportManagedServices(gctx, o, s, om, dir)
					})
				}(org, space)

				func(o cfclient.Org, s cfclient.Space) {
					g.Go(func() error {
						log.Infof("Exporting %s/%s user provided services", o.Name, s.Name)
						return e.ServiceInstanceExporter.ExportUserProvidedServices(gctx, o, s, dir)
					})
				}(org, space)
			}

			if len(spaces) < resultsPerPage {
				break
			}

			page++
		}
	}
	log.Debugln("Waiting for export to finish...")
	err := g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Warnln("context was canceled")
		}
		return err
	}
	log.Infoln("Finished exporting services")
	return nil
}

func (e OrgExporter) ExportAll(ctx context.Context, om config.OpsManager, dir string) error {
	g, gctx := errgroup.WithContext(ctx)
	client := e.ClientHolder.SourceCFClient()

	spaces, err := client.ListSpaces()
	if err != nil {
		return fmt.Errorf("error listing all spaces: %w", err)
	}

	for _, space := range spaces {
		org, err := client.GetOrgByGuid(space.OrganizationGuid)
		if err != nil {
			return fmt.Errorf("error getting the org %s for space %s: %w", space.OrganizationGuid, space.Name, err)
		}

		if e.ExcludeOrg(org.Name) || !e.IncludeOrg(org.Name) {
			log.Debugf("Excluding %s/%s", space.Name, org.Name)
			continue
		}

		func(o cfclient.Org, s cfclient.Space) {
			g.Go(func() error {
				log.Infof("Exporting %s/%s managed services", o.Name, s.Name)
				return e.ServiceInstanceExporter.ExportManagedServices(gctx, o, s, om, dir)
			})
		}(org, space)

		func(o cfclient.Org, s cfclient.Space) {
			g.Go(func() error {
				log.Infof("Exporting %s/%s user provided services", o.Name, s.Name)
				return e.ServiceInstanceExporter.ExportUserProvidedServices(gctx, o, s, dir)
			})
		}(org, space)
	}

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

func (e OrgExporter) ExcludeOrg(orgName string) bool {
	for _, re := range e.ExcludedOrgs {
		match, _ := regexp.Match(re, []byte(orgName))
		if match {
			return true
		}
	}

	return false
}

func (e OrgExporter) IncludeOrg(orgName string) bool {
	if len(e.IncludedOrgs) == 0 {
		return true
	}

	for _, re := range e.IncludedOrgs {
		match, _ := regexp.Match(re, []byte(orgName))
		if match {
			return true
		}
	}

	return false
}
