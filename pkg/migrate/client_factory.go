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
	"crypto/x509"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
)

type ClientFactory struct {
	ConfigLoader     config.Loader
	BoshFactory      bosh.ClientFactory
	OpsmanFactory    om.ClientFactory
	OpsmanConfig     config.OpsManager
	sourceCFClient   cf.Client
	targetCFClient   cf.Client
	sourceOMClient   om.Client
	targetOMClient   om.Client
	sourceBOSHClient bosh.Client
	targetBOSHClient bosh.Client
}

func NewClientFactory(
	configLoader config.Loader,
	boshFactory bosh.ClientFactory,
	opsmanFactory om.ClientFactory,
	opsManConfig config.OpsManager,
) *ClientFactory {
	return &ClientFactory{ConfigLoader: configLoader, BoshFactory: boshFactory, OpsmanFactory: opsmanFactory, OpsmanConfig: opsManConfig}
}

func (f *ClientFactory) CFClient(toSource bool) cf.Client {
	if toSource {
		return f.SourceCFClient()
	}
	return f.TargetCFClient()
}

func (f *ClientFactory) SourceCFClient() cf.Client {
	if f.sourceCFClient == nil {
		cfg := f.ConfigLoader.SourceApiConfig()
		f.sourceCFClient = newCFClient(cfg)
	}
	return f.sourceCFClient
}

func (f *ClientFactory) TargetCFClient() cf.Client {
	if f.targetCFClient == nil {
		cfg := f.ConfigLoader.TargetApiConfig()
		f.targetCFClient = newCFClient(cfg)
	}
	return f.targetCFClient
}

func (f *ClientFactory) SourceOpsManClient() om.Client {
	if f.sourceOMClient == nil {
		f.sourceOMClient = newOpsManClient(f.OpsmanFactory, f.OpsmanConfig)
	}
	return f.sourceOMClient
}

func (f *ClientFactory) TargetOpsManClient() om.Client {
	if f.targetOMClient == nil {
		f.targetOMClient = newOpsManClient(f.OpsmanFactory, f.OpsmanConfig)
	}
	return f.targetOMClient
}

func (f *ClientFactory) SourceBoshClient() bosh.Client {
	if f.sourceBOSHClient == nil {
		cfg := f.ConfigLoader.SourceBoshConfig()
		f.sourceBOSHClient = newBoshClient(f.BoshFactory, *cfg)
	}
	return f.sourceBOSHClient
}

func (f *ClientFactory) TargetBoshClient() bosh.Client {
	if f.targetBOSHClient == nil {
		cfg := f.ConfigLoader.TargetBoshConfig()
		f.targetBOSHClient = newBoshClient(f.BoshFactory, *cfg)
	}
	return f.targetBOSHClient
}

func newBoshClient(b bosh.ClientFactory, cfg config.Bosh) bosh.Client {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("error getting a certificate pool to append our trusted cert to: %v", err)
	}

	client, err := b.New(
		cfg.URL,
		cfg.AllProxy,
		[]byte(cfg.TrustedCert),
		certPool,
		cfg.Authentication)
	if err != nil {
		log.Fatalf("error creating bosh client: %v", err)
	}

	return client
}

func newCFClient(cfg *config.CloudController) cf.Client {
	client, err := cf.NewClient(&cf.Config{
		URL:          cfg.URL,
		Username:     cfg.Username,
		Password:     cfg.Password,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		SSLDisabled:  true,
	})
	if err != nil {
		log.Fatalf("error creating cf client %v", err)
	}
	return client
}

func newOpsManClient(c om.ClientFactory, cfg config.OpsManager) om.Client {
	var auth config.Authentication
	if len(cfg.ClientSecret) > 0 {
		auth = config.Authentication{
			UAA: config.UAAAuthentication{
				URL: cfg.URL + "/uaa",
				ClientCredentials: config.ClientCredentials{
					ID:     cfg.ClientID,
					Secret: cfg.ClientSecret,
				},
			},
		}
	} else {
		auth = config.Authentication{
			UAA: config.UAAAuthentication{
				URL: cfg.URL + "/uaa",
				UserCredentials: config.UserCredentials{
					Username: cfg.Username,
					Password: cfg.Password,
				},
			},
		}
	}
	client, err := c.New(
		cfg.URL,
		nil,
		nil,
		auth)
	if err != nil {
		log.Fatalf("error creating opsman client: %v", err)
	}
	return client
}
