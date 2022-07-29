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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"

	"github.com/cloudfoundry/bosh-cli/director"
)

type DirectorClient struct {
	clientRequest     ClientRequest
	taskClientRequest TaskClientRequest
}

func NewClient(
	endpoint string,
	httpClient *httpclient.DefaultHTTPClient,
	taskReporter director.TaskReporter,
	fileReporter director.FileReporter,
) DirectorClient {
	clientRequest := NewClientRequest(endpoint, httpClient, fileReporter)
	taskClientRequest := NewTaskClientRequest(clientRequest, taskReporter, 500*time.Millisecond)
	return DirectorClient{clientRequest, taskClientRequest}
}

func (c DirectorClient) WithContext(contextId string) DirectorClient {
	clientRequest := c.clientRequest.WithContext(contextId)

	taskClientRequest := c.taskClientRequest
	taskClientRequest.clientRequest = clientRequest

	return DirectorClient{clientRequest, taskClientRequest}
}

func (c DirectorClient) DeploymentsWithoutConfigs() ([]director.DeploymentResp, error) {
	var deps []director.DeploymentResp

	err := c.clientRequest.Get("/deployments?exclude_configs=true", &deps)
	if err != nil {
		return deps, fmt.Errorf("error finding deployments: %w", err)
	}

	return deps, nil
}

func (c DirectorClient) Info() (director.InfoResp, error) {
	var info director.InfoResp

	err := c.clientRequest.Get("/info", &info)
	if err != nil {
		return info, fmt.Errorf("error fetching info: %w", err)
	}

	return info, nil
}

func (c DirectorClient) DeploymentVMInfos(deploymentName string) ([]director.VMInfo, error) {
	return c.deploymentResourceInfos(deploymentName, "vms")
}

func (c DirectorClient) deploymentResourceInfos(deploymentName string, resourceType string) ([]director.VMInfo, error) {
	if len(deploymentName) == 0 {
		return nil, fmt.Errorf("expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/%s?format=full", deploymentName, resourceType)

	_, resultBytes, err := c.taskClientRequest.GetResult(path)
	if err != nil {
		return nil, fmt.Errorf(
			"error listing deployment '%s' %s infos: %w", deploymentName, resourceType, err)
	}

	var resps []director.VMInfo

	for _, piece := range strings.Split(string(resultBytes), "\n") {
		if len(piece) == 0 {
			continue
		}

		var resp director.VMInfo

		err := json.Unmarshal([]byte(piece), &resp)
		if err != nil {
			return nil, fmt.Errorf(
				"error unmarshaling %s info response: '%s': %w", strings.TrimSuffix(resourceType, "s"), piece, err)
		}

		resp.Deployment = deploymentName

		if len(resp.DiskIDs) == 0 && resp.DiskID != "" {
			resp.DiskIDs = []string{resp.DiskID}
		}

		resp.VMCreatedAt, err = director.TimeParser{}.Parse(resp.VMCreatedAtRaw)
		if err != nil {
			return resps, fmt.Errorf("error converting created_at '%s' to time: %w", resp.VMCreatedAtRaw, err)
		}

		resps = append(resps, resp)
	}

	return resps, nil
}
