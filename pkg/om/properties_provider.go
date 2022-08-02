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
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type PropertiesProvider struct {
	cfg            *config.Config
	opsman         Client
	boshFactory    bosh.ClientFactory
	credhubFactory credhub.ClientFactory
}

func NewPropertiesProvider(cfg *config.Config, opsmanConfig config.OpsManager, opsmanFactory ClientFactory, boshFactory bosh.ClientFactory, credhubFactory credhub.ClientFactory) *PropertiesProvider {
	return &PropertiesProvider{cfg: cfg, opsman: newOpsManClient(opsmanFactory, opsmanConfig), boshFactory: boshFactory, credhubFactory: credhubFactory}
}

func (l PropertiesProvider) Environment(bb config.BoshPropertiesBuilder, cfb config.CFPropertiesBuilder, ccb config.CCDBPropertiesBuilder) config.EnvProperties {
	return Environment(bb.Build(), cfb.Build(), ccb.Build())
}

func (l PropertiesProvider) SourceBoshPropertiesBuilder() config.BoshPropertiesBuilder {
	if err := l.cfg.Foundations.Source.Validate(); err != nil {
		panic(fmt.Sprintf("Error validating config.\n\nPlease check foundations.source config: %q", l.cfg.ConfigFile))
	}
	return NewBOSHPropertiesBuilder(l.opsman, l.cfg.Foundations.Source)
}

func (l PropertiesProvider) TargetBoshPropertiesBuilder() config.BoshPropertiesBuilder {
	if err := l.cfg.Foundations.Target.Validate(); err != nil {
		panic(fmt.Sprintf("Error validating config.\n\nPlease check foundations.target config: %q", l.cfg.ConfigFile))
	}
	return NewBOSHPropertiesBuilder(l.opsman, l.cfg.Foundations.Target)
}

func (l PropertiesProvider) SourceCFPropertiesBuilder() config.CFPropertiesBuilder {
	if err := l.cfg.SourceBosh.Validate(); err != nil {
		if err := l.cfg.Foundations.Source.Validate(); err != nil {
			panic(fmt.Sprintf("Error validating config.\n\nPlease check source_bosh or foundations.source config: %q", l.cfg.ConfigFile))
		}
	}
	return NewCFPropertiesBuilder(l.opsman, l.cfg.SourceBosh.Deployment)
}

func (l PropertiesProvider) TargetCFPropertiesBuilder() config.CFPropertiesBuilder {
	if err := l.cfg.TargetBosh.Validate(); err != nil {
		if err := l.cfg.Foundations.Target.Validate(); err != nil {
			panic(fmt.Sprintf("Error validating config.\n\nPlease check target_bosh or foundations.target config: %q", l.cfg.ConfigFile))
		}
	}
	return NewCFPropertiesBuilder(l.opsman, l.cfg.TargetBosh.Deployment)
}

func (l PropertiesProvider) SourceCCDBPropertiesBuilder(b config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
	if err := l.cfg.Foundations.Source.Validate(); err != nil {
		panic(fmt.Sprintf("Error validating config.\n\nPlease check foundations.source config: %q", l.cfg.ConfigFile))
	}
	return NewCCDBPropertiesBuilder(l.opsman, l.cfg.Foundations.Source, l.boshFactory, l.credhubFactory, b.Build())
}

func (l PropertiesProvider) TargetCCDBPropertiesBuilder(b config.BoshPropertiesBuilder) config.CCDBPropertiesBuilder {
	if err := l.cfg.Foundations.Target.Validate(); err != nil {
		panic(fmt.Sprintf("Error validating config.\n\nPlease check foundations.target config: %q", l.cfg.ConfigFile))
	}
	return NewCCDBPropertiesBuilder(l.opsman, l.cfg.Foundations.Target, l.boshFactory, l.credhubFactory, b.Build())
}

func newOpsManClient(c ClientFactory, cfg config.OpsManager) Client {
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
