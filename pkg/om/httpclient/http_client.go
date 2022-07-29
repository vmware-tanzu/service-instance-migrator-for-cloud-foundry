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

package httpclient

import (
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
)

type credential struct {
	Credential string `json:"credential"`
}

type CertificateAuthorities struct {
	CAs []CA `json:"certificate_authorities"`
}

type CA struct {
	GUID      string `json:"guid"`
	Issuer    string `json:"issuer"`
	CreatedOn string `json:"created_on"`
	ExpiresOn string `json:"expires_on"`
	Active    bool   `json:"active"`
	CertPEM   string `json:"cert_pem"`
}

type DeployedProduct struct {
	Type string
	GUID string
}

type DeployedProductCredential struct {
	Credential Credential `json:"credential"`
}

type Credential struct {
	Type  string            `json:"type"`
	Value map[string]string `json:"value"`
}

type ResponseProperty struct {
	Value          interface{}
	SelectedOption string `yaml:"selected_option,omitempty"`
	Configurable   bool
	IsCredential   bool   `yaml:"credential"`
	Type           string `yaml:"type"`
}

type OpsManHTTPClient struct {
	clientRequest ClientRequest
}

func NewHTTPClient(
	endpoint string,
	httpClient *httpclient.DefaultHTTPClient,
) OpsManHTTPClient {
	return OpsManHTTPClient{
		clientRequest: NewClientRequest(endpoint, httpClient),
	}
}

func (c OpsManHTTPClient) GetBOSHCredentials() (string, error) {
	var credential credential
	err := c.clientRequest.Get("/api/v0/deployed/director/credentials/bosh_commandline_credentials", &credential)
	if err != nil {
		return "", fmt.Errorf("failed to get bosh commandline credentials: %w", err)
	}
	return credential.Credential, nil
}

func (c OpsManHTTPClient) ListCertificateAuthorities() ([]CA, error) {
	var authorities CertificateAuthorities
	err := c.clientRequest.Get("/api/v0/certificate_authorities", &authorities)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate authorities: %w", err)
	}
	return authorities.CAs, nil
}

func (c OpsManHTTPClient) ListDeployedProductCredentials(deployedProductGUID string, credentialRef string) (DeployedProductCredential, error) {
	var credential DeployedProductCredential
	err := c.clientRequest.Get(fmt.Sprintf("/api/v0/deployed/products/%s/credentials/%s", deployedProductGUID, credentialRef), &credential)
	if err != nil {
		return DeployedProductCredential{}, fmt.Errorf("failed to get deployed product credentials: %w", err)
	}
	return credential, nil
}

func (c OpsManHTTPClient) GetStagedProductProperties(product string) (map[string]ResponseProperty, error) {
	var propertiesResponse struct {
		Properties map[string]ResponseProperty
	}
	err := c.clientRequest.Get(fmt.Sprintf("/api/v0/staged/products/%s/properties?redact=true", product), &propertiesResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get staged product properties: %w", err)
	}
	return propertiesResponse.Properties, nil
}

func (c OpsManHTTPClient) ListDeployedProducts() ([]DeployedProduct, error) {
	var deployedProducts []DeployedProduct
	err := c.clientRequest.Get("/api/v0/deployed/products", &deployedProducts)
	if err != nil {
		return []DeployedProduct{}, fmt.Errorf("failed to get deployed products: %w", err)
	}
	return deployedProducts, nil
}
