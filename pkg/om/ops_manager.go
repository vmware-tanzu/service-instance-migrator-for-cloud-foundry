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

package om

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
	"strings"
)

//counterfeiter:generate -o fakes . OpsManClient

type OpsManClient interface {
	GetBOSHCredentials() (string, error)
	ListCertificateAuthorities() ([]httpclient.CA, error)
	ListDeployedProductCredentials(deployedProductGUID string, credentialRef string) (httpclient.DeployedProductCredential, error)
	GetStagedProductProperties(product string) (map[string]httpclient.ResponseProperty, error)
	ListDeployedProducts() ([]httpclient.DeployedProduct, error)
}

type BoshCredentials struct {
	Client       string
	ClientSecret string
	Environment  string
}

type opsManagerImpl struct {
	client OpsManClient
}

func NewOpsManager(client OpsManClient) *opsManagerImpl {
	return &opsManagerImpl{client: client}
}

func (o opsManagerImpl) GetBOSHCredentials() (BoshCredentials, error) {
	credentials, err := o.client.GetBOSHCredentials()
	if err != nil {
		return BoshCredentials{}, err
	}

	keyValues := parseKeyValues(credentials)
	return BoshCredentials{
		Client:       keyValues["BOSH_CLIENT"],
		ClientSecret: keyValues["BOSH_CLIENT_SECRET"],
		Environment:  keyValues["BOSH_ENVIRONMENT"],
	}, nil
}

func (o opsManagerImpl) GetCertificateAuthorities() ([]httpclient.CA, error) {
	return o.client.ListCertificateAuthorities()
}

func (o opsManagerImpl) GetDeployedProductCredentials(deployedProductGUID string, credentialRef string) (httpclient.DeployedProductCredential, error) {
	return o.client.ListDeployedProductCredentials(deployedProductGUID, credentialRef)
}

func (o opsManagerImpl) GetStagedProductProperties(product string) (map[string]httpclient.ResponseProperty, error) {
	return o.client.GetStagedProductProperties(product)
}

func (o opsManagerImpl) ListDeployedProducts() ([]httpclient.DeployedProduct, error) {
	return o.client.ListDeployedProducts()
}

func parseKeyValues(credentials string) map[string]string {
	values := make(map[string]string)
	kvs := strings.Split(credentials, " ")
	for _, kv := range kvs {
		if strings.Contains(kv, "=") {
			k := strings.Split(kv, "=")[0]
			v := strings.Split(kv, "=")[1]
			values[k] = v
		}
	}
	return values
}
