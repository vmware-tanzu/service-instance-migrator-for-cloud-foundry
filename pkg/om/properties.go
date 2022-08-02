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
	"crypto/x509"
	"fmt"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	boshcli "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/cli"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/net"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"
)

const (
	CFTypeString = "cf"
)

type boshPropertiesBuilder struct {
	Client     Client
	OpsManager config.OpsManager
	Properties *config.BoshProperties
}

type cfPropertiesBuilder struct {
	Client     Client
	Properties *config.CFProperties
	Deployment string
}

type ccdbPropertiesBuilder struct {
	Client         Client
	CCDBProperties *config.CCDBProperties
	BoshProperties *config.BoshProperties
	BoshFactory    bosh.ClientFactory
	CredhubFactory credhub.ClientFactory
	OpsManager     config.OpsManager
}

func Environment(boshProperties *config.BoshProperties, cfProperties *config.CFProperties, ccdbProperties *config.CCDBProperties) config.EnvProperties {
	return config.EnvProperties{
		BoshProperties: boshProperties,
		CFProperties:   cfProperties,
		CCDBProperties: ccdbProperties,
	}
}

func NewCFPropertiesBuilder(client Client, cfDeployment string) *cfPropertiesBuilder {
	return &cfPropertiesBuilder{Client: client, Properties: &config.CFProperties{}, Deployment: cfDeployment}
}

func NewBOSHPropertiesBuilder(client Client, opsManager config.OpsManager) *boshPropertiesBuilder {
	return &boshPropertiesBuilder{Client: client, Properties: &config.BoshProperties{}, OpsManager: opsManager}
}

func NewCCDBPropertiesBuilder(client Client, opsman config.OpsManager, boshFactory bosh.ClientFactory, credhubFactory credhub.ClientFactory, boshProperties *config.BoshProperties) *ccdbPropertiesBuilder {
	return &ccdbPropertiesBuilder{Client: client, BoshFactory: boshFactory, CredhubFactory: credhubFactory, BoshProperties: boshProperties, CCDBProperties: &config.CCDBProperties{}, OpsManager: opsman}
}

func (b *boshPropertiesBuilder) Build() *config.BoshProperties {
	deployedProductGUID, err := b.Client.DeployedProduct(CFTypeString)
	if err != nil {
		log.Fatalln(err)
	}

	env, clientID, clientSecret, err := b.Client.BoshEnvironment()
	if err != nil {
		log.Fatalln(err)
	}

	authorities, err := b.Client.CertificateAuthorities()
	if err != nil {
		log.Fatalln(err)
	}

	scheme, host, port, _, err := net.ParseURL(env)
	if err != nil {
		log.Fatalln(err)
	}

	boshURL := fmt.Sprintf("%s://%s", scheme, host)
	if port != 443 && port != 0 {
		boshURL = fmt.Sprintf("%s://%s:%d", scheme, host, port)
	}

	b.Properties.URL = boshURL
	b.Properties.AllProxy = AllProxyURL(b.OpsManager)
	b.Properties.ClientID = clientID
	b.Properties.ClientSecret = clientSecret
	b.Properties.RootCA = authorities
	b.Properties.Deployment = deployedProductGUID

	return b.Properties
}

func (b *cfPropertiesBuilder) Build() *config.CFProperties {
	if b.Deployment == "" {
		deployedProductGUID, err := b.Client.DeployedProduct(CFTypeString)
		if err != nil {
			log.Fatalln(err)
		}
		b.Deployment = deployedProductGUID
	}

	uaaAdminCreds, err := b.Client.DeployedProductCredentials(b.Deployment, ".uaa.admin_credentials")
	if err != nil {
		log.Fatalln(err)
	}

	props, err := b.Client.StagedProductProperties(b.Deployment)
	if err != nil {
		log.Fatalln(err)
	}

	b.Properties.URL = "https://api." + props[".cloud_controller.system_domain"].Value.(string)
	b.Properties.Username = uaaAdminCreds.Credential.Value["identity"]
	b.Properties.Password = uaaAdminCreds.Credential.Value["password"]
	return b.Properties
}

func (b *ccdbPropertiesBuilder) Build() *config.CCDBProperties {
	setCCDBHost(b.CCDBProperties, *b.BoshProperties, b.BoshFactory, b.BoshProperties.Deployment)
	setCCDBCredentials(*b.BoshProperties, b.CCDBProperties, b.CredhubFactory, b.BoshProperties.Deployment)

	ccdbEncCreds, err := b.Client.DeployedProductCredentials(b.BoshProperties.Deployment, ".cloud_controller.db_encryption_credentials")
	if err != nil {
		log.Fatalln(err)
	}
	b.CCDBProperties.EncryptionKey = ccdbEncCreds.Credential.Value["password"]

	b.CCDBProperties.SSHHost = b.OpsManager.Hostname
	b.CCDBProperties.SSHUsername = b.OpsManager.SshUser
	b.CCDBProperties.SSHPrivateKey = b.OpsManager.PrivateKey
	return b.CCDBProperties
}

func setCCDBHost(properties *config.CCDBProperties, boshProperties config.BoshProperties, boshFactory bosh.ClientFactory, cfDeployment string) {
	boshClient := newBoshClient(boshFactory, boshProperties)
	vm, found, err := boshClient.FindVM(cfDeployment, "proxy")
	if err != nil {
		log.Fatalln(err)
	}

	if !found {
		log.Fatalf("could not find vm with running process: %v\n", "proxy")
	}

	if len(vm.IPs) > 0 {
		properties.Host = vm.IPs[0]
	}
}

func setCCDBCredentials(boshProperties config.BoshProperties, ccdbProperties *config.CCDBProperties, credhubFactory credhub.ClientFactory, cfDeployment string) {
	client := newCredhubClient(credhubFactory, boshProperties)
	credhubCreds, err := client.GetCreds(fmt.Sprintf("/p-bosh/%s/cc-db-credentials", cfDeployment))
	if err != nil {
		log.Fatalf("failed to get ccdb credentials from credhub, %v", err)
	}

	data := credhubCreds["data"]
	value := data[0]["value"].(map[string]interface{})
	ccdbProperties.Username = value["username"].(string)
	ccdbProperties.Password = value["password"].(string)
}

func newCredhubClient(c credhub.ClientFactory, properties config.BoshProperties) credhub.Client {
	scheme, host, _, _, err := net.ParseURL(properties.URL)
	if err != nil {
		log.Fatalln(err)
	}
	var trustedCert string
	if len(properties.RootCA) > 0 {
		trustedCert = properties.RootCA[len(properties.RootCA)-1].CertPEM
	}

	return c.New(
		fmt.Sprintf("%s://%s", scheme, host),
		"8844",
		"8443",
		properties.AllProxy,
		[]byte(trustedCert),
		properties.ClientID,
		properties.ClientSecret)
}

func AllProxyURL(opsman config.OpsManager) string {
	if err := opsman.ValidateSSH(); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("ssh+socks5://%s@%s:22?private-key=%s", opsman.SshUser, opsman.Hostname, opsman.PrivateKey)
}

func newBoshClient(b bosh.ClientFactory, properties config.BoshProperties) bosh.Client {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("error getting a certificate pool to append our trusted cert to: %v", err)
	}
	var trustedCert string
	if len(properties.RootCA) > 0 {
		trustedCert = properties.RootCA[len(properties.RootCA)-1].CertPEM
	}

	auth := config.Authentication{}
	scheme, host, _, _, err := net.ParseURL(properties.URL)
	if err != nil {
		log.Fatalln(err)
	}
	auth.UAA.URL = fmt.Sprintf("%s://%s:%d", scheme, host, 8443)
	auth.UAA.ClientCredentials.ID = properties.ClientID
	auth.UAA.ClientCredentials.Secret = properties.ClientSecret

	client, err := b.New(
		properties.URL,
		properties.AllProxy,
		[]byte(trustedCert),
		certPool,
		boshcli.NewFactory(),
		uaa.NewFactory(),
		auth)
	if err != nil {
		log.Fatalf("error creating bosh client: %v", err)
	}

	return client
}
