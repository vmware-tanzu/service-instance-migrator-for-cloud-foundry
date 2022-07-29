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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	sio "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"golang.org/x/sync/errgroup"
)

type SpaceImporter struct {
	ServiceInstanceImporter
}

func NewSpaceImporter(serviceInstanceImporter ServiceInstanceImporter) *SpaceImporter {
	return &SpaceImporter{
		ServiceInstanceImporter: serviceInstanceImporter,
	}
}

func (i SpaceImporter) Import(ctx context.Context, om config.OpsManager, dir string, orgName, spaceName string) error {
	serviceInstanceMap, err := i.createServiceInstances(dir)
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)
	for org, spaces := range serviceInstanceMap {
		for space, instances := range spaces {
			importInstances := func(ctx context.Context, org, space string, instances []*cf.ServiceInstance, dir string) {
				for _, instance := range instances {
					func(ctx context.Context, org string, space string, si *cf.ServiceInstance, dir string) {
						g.Go(func() error {
							if si.Service != "" && si.Type != "" && si.Name != "" {
								if !i.shouldMigrate(ctx, si) {
									return nil
								}
								err := i.ImportManagedService(ctx, org, space, si, om, dir)
								if err != nil {
									return err
								}
							}
							return nil
						})
					}(gctx, org, space, instance, dir)
				}
			}
			if org == orgName && space == spaceName {
				importInstances(gctx, org, space, instances, dir)
			}
		}
	}
	return g.Wait()
}

func (i SpaceImporter) createServiceInstances(dir string, orgNames ...string) (map[string]map[string][]*cf.ServiceInstance, error) {
	orgs := make(map[string]map[string][]*cf.ServiceInstance)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fd, err := sio.NewFileDescriptor(path)
			var invalidExtErr *sio.InvalidFileExtensionError
			if err != nil {
				if errors.As(err, &invalidExtErr) {
					return nil
				}
				return err
			}
			instance := &cf.ServiceInstance{}
			err = sio.NewParser().Unmarshal(instance, fd)
			if err != nil {
				return err
			}

			addInstance := func(org, space string, instance *cf.ServiceInstance) {
				if vo, ok := orgs[org]; ok {
					if vs, ok := vo[space]; ok {
						vo[space] = append(vs, instance)
					} else {
						vo[space] = append(make([]*cf.ServiceInstance, 0), instance)
					}
				} else {
					orgs[org] = map[string][]*cf.ServiceInstance{
						space: append(make([]*cf.ServiceInstance, 0), instance),
					}
				}
			}

			org, space := sio.GetOrgSpace(path)
			if len(orgNames) == 0 {
				addInstance(org, space, instance)
				return nil
			}

			// only add the instance if its in one of the provided orgs
			for _, orgName := range orgNames {
				if org == orgName {
					addInstance(org, space, instance)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (i SpaceImporter) shouldMigrate(ctx context.Context, instance *cf.ServiceInstance) bool {
	if cfg, ok := config.FromContext(ctx); ok {

		if len(cfg.Instances) == 0 {
			return true
		}

		for _, inst := range cfg.Instances {
			if inst == instance.Name {
				return true
			}
		}
	}

	return false
}
