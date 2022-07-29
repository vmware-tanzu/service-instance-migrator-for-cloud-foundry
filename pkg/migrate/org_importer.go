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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"regexp"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"golang.org/x/sync/errgroup"
)

type OrgImporter struct {
	IncludedOrgs []string
	ExcludedOrgs []string
	*SpaceImporter
}

func NewOrgImporter(spaceImporter *SpaceImporter, includedOrgs []string, excludedOrgs []string) *OrgImporter {
	return &OrgImporter{
		SpaceImporter: spaceImporter,
		IncludedOrgs:  includedOrgs,
		ExcludedOrgs:  excludedOrgs,
	}
}

func (i OrgImporter) Import(ctx context.Context, om config.OpsManager, dir string, orgs ...string) error {
	g, gctx := errgroup.WithContext(ctx)

	err := i.importOrg(gctx, g, om, dir, orgs...)
	if err != nil {
		return err
	}

	return g.Wait()
}

func (i OrgImporter) ImportAll(ctx context.Context, om config.OpsManager, dir string) error {
	g, gctx := errgroup.WithContext(ctx)

	err := i.importOrg(gctx, g, om, dir)
	if err != nil {
		return err
	}

	return g.Wait()
}

func (i OrgImporter) importOrg(ctx context.Context, g *errgroup.Group, om config.OpsManager, dir string, orgs ...string) error {
	serviceInstanceMap, err := i.createServiceInstances(dir, orgs...)
	if err != nil {
		return err
	}

	for org, spaces := range serviceInstanceMap {
		if i.ExcludeOrg(org) || !i.IncludeOrg(org) {
			log.Infof("Excluding %q", org)
			continue
		}

		for space, instances := range spaces {
			func(org, space string, instances []*cf.ServiceInstance) {
				for _, instance := range instances {
					func(si *cf.ServiceInstance) {
						g.Go(func() error {
							if !i.shouldMigrate(ctx, si) {
								return nil
							}
							err := i.ImportManagedService(ctx, org, space, si, om, dir)
							if err != nil {
								return err
							}

							return nil
						})
					}(instance)
				}
			}(org, space, instances)
		}
	}

	return nil
}

func (i OrgImporter) ExcludeOrg(orgName string) bool {
	for _, re := range i.ExcludedOrgs {
		match, _ := regexp.Match(re, []byte(orgName))
		if match {
			return true
		}
	}

	return false
}

func (i OrgImporter) IncludeOrg(orgName string) bool {
	if len(i.IncludedOrgs) == 0 {
		return true
	}

	for _, re := range i.IncludedOrgs {
		match, _ := regexp.Match(re, []byte(orgName))
		if match {
			return true
		}
	}

	return false
}
