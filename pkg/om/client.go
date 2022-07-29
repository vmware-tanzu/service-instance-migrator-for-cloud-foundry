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
	"errors"
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . CertAppender

type CertAppender interface {
	AppendCertsFromPEM(pemCerts []byte) (ok bool)
}

//counterfeiter:generate -o fakes . OpsManager

type OpsManager interface {
	GetBOSHCredentials() (BoshCredentials, error)
	GetCertificateAuthorities() ([]httpclient.CA, error)
	GetDeployedProductCredentials(deployedProductGUID string, credentialRef string) (httpclient.DeployedProductCredential, error)
	GetStagedProductProperties(product string) (map[string]httpclient.ResponseProperty, error)
	ListDeployedProducts() ([]httpclient.DeployedProduct, error)
}

//counterfeiter:generate -o fakes . OpsManagerFactory

type OpsManagerFactory interface {
	New(config Config) (OpsManager, error)
}

//counterfeiter:generate -o fakes . UAAFactory

type UAAFactory interface {
	New(config uaa.Config) (uaa.UAA, error)
}

//counterfeiter:generate -o fakes . Client

type Client interface {
	BoshEnvironment() (string, string, string, error)
	CertificateAuthorities() ([]httpclient.CA, error)
	DeployedProductCredentials(deployedProductGUID string, credentialRef string) (httpclient.DeployedProductCredential, error)
	StagedProductProperties(deployedProductGUID string) (map[string]httpclient.ResponseProperty, error)
	DeployedProduct(productType string) (string, error)
}

//counterfeiter:generate -o fakes . ClientFactory

type ClientFactory interface {
	New(url string,
		trustedCertPEM []byte,
		certAppender CertAppender,
		opsManagerFactory OpsManagerFactory,
		uaaFactory UAAFactory,
		boshAuth config.Authentication) (Client, error)
}

type ClientFactoryFunc func(url string, trustedCertPEM []byte, certAppender CertAppender, opsManagerFactory OpsManagerFactory, uaaFactory UAAFactory, auth config.Authentication) (Client, error)

func (f ClientFactoryFunc) New(
	url string,
	trustedCertPEM []byte,
	certAppender CertAppender,
	opsManagerFactory OpsManagerFactory,
	uaaFactory UAAFactory,
	auth config.Authentication,
) (Client, error) {
	return f(url, trustedCertPEM, certAppender, opsManagerFactory, uaaFactory, auth)
}

type ClientImpl struct {
	URL               string
	TrustedCertPEM    []byte
	opsManagerFactory OpsManagerFactory
	uaaFactory        UAAFactory
	Auth              config.Authentication
}

func NewClient() ClientFactoryFunc {
	return New
}

func New(
	url string,
	trustedCertPEM []byte,
	certAppender CertAppender,
	opsManagerFactory OpsManagerFactory,
	uaaFactory UAAFactory,
	auth config.Authentication,
) (Client, error) {

	if certAppender != nil {
		certAppender.AppendCertsFromPEM(trustedCertPEM)
	}

	return &ClientImpl{
		URL:               url,
		TrustedCertPEM:    trustedCertPEM,
		Auth:              auth,
		uaaFactory:        uaaFactory,
		opsManagerFactory: opsManagerFactory,
	}, nil
}

func (c *ClientImpl) OpsMan() (OpsManager, error) {
	cfg, err := c.opsManConfig()
	if err != nil {
		return opsManagerImpl{}, err
	}
	return c.opsManagerFactory.New(cfg)
}

func (c *ClientImpl) BoshEnvironment() (string, string, string, error) {
	o, err := c.OpsMan()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create ops manager client: %w", err)
	}

	creds, err := o.GetBOSHCredentials()
	if err != nil {
		return "", "", "", fmt.Errorf("cannot get bosh environment variables from ops manager: %w", err)
	}

	return creds.Environment, creds.Client, creds.ClientSecret, nil
}

func (c *ClientImpl) CertificateAuthorities() ([]httpclient.CA, error) {
	o, err := c.OpsMan()
	if err != nil {
		return []httpclient.CA{}, fmt.Errorf("failed to create ops manager client: %w", err)
	}

	authorities, err := o.GetCertificateAuthorities()
	if err != nil {
		return []httpclient.CA{}, fmt.Errorf("cannot get certificate authorities from ops manager: %w", err)
	}

	return authorities, err
}

func (c *ClientImpl) DeployedProductCredentials(deployedProductGUID string, credentialRef string) (httpclient.DeployedProductCredential, error) {
	o, err := c.OpsMan()
	if err != nil {
		return httpclient.DeployedProductCredential{}, fmt.Errorf("failed to create ops manager client: %w", err)
	}

	credentials, err := o.GetDeployedProductCredentials(deployedProductGUID, credentialRef)
	if err != nil {
		return httpclient.DeployedProductCredential{}, fmt.Errorf("cannot get deployed product credentials from ops manager, product guid %q: %w", deployedProductGUID, err)
	}

	return credentials, err
}

func (c *ClientImpl) StagedProductProperties(deployedProductGUID string) (map[string]httpclient.ResponseProperty, error) {
	o, err := c.OpsMan()
	if err != nil {
		return nil, fmt.Errorf("failed to create ops manager client: %w", err)
	}

	resp, err := o.GetStagedProductProperties(deployedProductGUID)
	if err != nil {
		return nil, fmt.Errorf("cannot get staged product properties from ops manager, product guid %q: %w", deployedProductGUID, err)
	}

	return resp, err
}

func (c *ClientImpl) DeployedProduct(productType string) (string, error) {
	o, err := c.OpsMan()
	if err != nil {
		return "", fmt.Errorf("failed to create ops manager client: %w", err)
	}

	products, err := o.ListDeployedProducts()
	if err != nil {
		return "", fmt.Errorf("cannot get a list of deployed products from ops manager, product type %q: %w", productType, err)
	}

	for _, product := range products {
		if product.Type == productType {
			return product.GUID, nil
		}
	}

	return "", fmt.Errorf("failed to find product for type %q", productType)
}

func (c *ClientImpl) opsManConfig() (Config, error) {
	cfg, err := NewConfigFromURL(c.URL)
	if err != nil {
		return Config{}, fmt.Errorf("failed to build ops manager config from url")
	}
	cfg.CACert = string(c.TrustedCertPEM)

	if c.Auth.UAA.IsSet() {
		var uaaClient, err = buildUAA(c.Auth.UAA.URL, c.Auth, cfg.CACert, c.uaaFactory)
		if err != nil {
			return Config{}, fmt.Errorf("failed to build UAA client: %w", err)
		}

		if c.Auth.UAA.ClientCredentials.IsSet() {
			cfg.TokenFunc = uaa.NewClientTokenSession(uaaClient).ClientCredentialsTokenFunc
		} else if c.Auth.UAA.UserCredentials.IsSet() {
			cfg.TokenFunc = uaa.NewClientTokenSession(uaaClient,
				uaa.WithPasswordCredentials(
					c.Auth.UAA.UserCredentials.Username,
					c.Auth.UAA.UserCredentials.Password,
				),
			).OwnerPasswordCredentialsTokenFunc
		}
	} else {
		return Config{}, errors.New("uaa auth must be set for opsman api authentication")
	}

	return cfg, nil
}

func buildUAA(uaaURL string, auth config.Authentication, CACert string, factory UAAFactory) (uaa.UAA, error) {
	uaaConfig, err := uaa.NewConfigFromURL(uaaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build UAA config from url: %w", err)
	}
	if auth.UAA.ClientCredentials.IsSet() {
		uaaConfig.ClientID = auth.UAA.ClientCredentials.ID
		uaaConfig.ClientSecret = auth.UAA.ClientCredentials.Secret
	} else if auth.UAA.UserCredentials.IsSet() {
		uaaConfig.ClientID = "opsman"
		uaaConfig.ClientSecret = ""
		uaaConfig.Username = auth.UAA.UserCredentials.Username
		uaaConfig.Password = auth.UAA.UserCredentials.Password
	}
	uaaConfig.CACert = CACert
	return factory.New(uaaConfig)
}
