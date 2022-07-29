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

package cli

import (
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director"
)

//counterfeiter:generate -o fakes . Deployment

type Deployment interface {
	VMInfos() ([]director.VMInfo, error)
}

type DirectorImpl struct {
	client DirectorClient
}

func (d DirectorImpl) IsAuthenticated() (bool, error) {
	r, err := d.client.Info()
	if err != nil {
		return false, err
	}

	authed := len(r.User) > 0

	return authed, nil
}

func (d DirectorImpl) Info() (director.Info, error) {
	r, err := d.client.Info()
	if err != nil {
		return director.Info{}, err
	}

	info := director.Info{
		Name:    r.Name,
		UUID:    r.UUID,
		Version: r.Version,

		User: r.User,
		Auth: director.UserAuthentication{
			Type:    r.Auth.Type,
			Options: r.Auth.Options,
		},

		Features: map[string]bool{},

		CPI:             r.CPI,
		StemcellOS:      r.StemcellOS,
		StemcellVersion: r.StemcellVersion,
	}

	for k, featResp := range r.Features {
		info.Features[k] = featResp.Status
	}

	return info, nil
}

func (d DirectorImpl) ListDeployments() ([]director.DeploymentResp, error) {
	return d.client.DeploymentsWithoutConfigs()
}

func (d DirectorImpl) FindDeployment(name string) (Deployment, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("expected non-empty deployment name")
	}

	return &DeploymentImpl{client: d.client, name: name}, nil
}
