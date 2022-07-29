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
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/net"
)

type ConfigLoader struct {
	cfg                *config.Config
	migrationReader    config.MigrationReader
	propertiesProvider config.PropertiesProvider
}

func NewConfigLoader(
	cfg *config.Config,
	mr config.MigrationReader,
	pp config.PropertiesProvider,
) ConfigLoader {
	return ConfigLoader{
		cfg:                cfg,
		migrationReader:    mr,
		propertiesProvider: pp,
	}
}

func (l ConfigLoader) BuildSourceConfig() {
	boshPropertiesBuilder := l.propertiesProvider.SourceBoshPropertiesBuilder()
	cfPropertiesBuilder := l.propertiesProvider.SourceCFPropertiesBuilder()
	ccdbPropertiesBuilder := l.propertiesProvider.SourceCCDBPropertiesBuilder(boshPropertiesBuilder)
	l.toSource(l.propertiesProvider.Environment(boshPropertiesBuilder, cfPropertiesBuilder, ccdbPropertiesBuilder))
}

func (l ConfigLoader) BuildTargetConfig() {
	boshPropertiesBuilder := l.propertiesProvider.TargetBoshPropertiesBuilder()
	cfPropertiesBuilder := l.propertiesProvider.TargetCFPropertiesBuilder()
	ccdbPropertiesBuilder := l.propertiesProvider.TargetCCDBPropertiesBuilder(boshPropertiesBuilder)
	l.toTarget(l.propertiesProvider.Environment(boshPropertiesBuilder, cfPropertiesBuilder, ccdbPropertiesBuilder))
}

func (l ConfigLoader) BoshConfig(toSource bool) *config.Bosh {
	if toSource {
		return l.SourceBoshConfig()
	}
	return l.TargetBoshConfig()
}

func (l ConfigLoader) SourceBoshConfig() *config.Bosh {
	if !l.cfg.SourceBosh.IsSet() {
		log.Infoln("Source bosh config not set, so loading source config from opsman")
		boshClientProperties := l.propertiesProvider.SourceBoshPropertiesBuilder().Build()
		l.setBoshConfig(boshClientProperties, true)
		log.Debugf("Source bosh config: %+v", l.cfg.SourceBosh)
	}
	return &l.cfg.SourceBosh
}

func (l ConfigLoader) TargetBoshConfig() *config.Bosh {
	if !l.cfg.TargetBosh.IsSet() {
		log.Infoln("Target bosh config not set, so loading target config from opsman")
		boshClientProperties := l.propertiesProvider.TargetBoshPropertiesBuilder().Build()
		l.setBoshConfig(boshClientProperties, false)
		log.Debugf("Target bosh config: %+v", l.cfg.TargetBosh)
	}
	return &l.cfg.TargetBosh
}

func (l ConfigLoader) CFConfig(toSource bool) *config.CloudController {
	if toSource {
		return l.SourceApiConfig()
	}
	return l.TargetApiConfig()
}

func (l ConfigLoader) SourceApiConfig() *config.CloudController {
	if !l.cfg.SourceApi.IsSet() {
		log.Infoln("Source api config not set; loading source config from opsman")
		cfClientProperties := l.propertiesProvider.SourceCFPropertiesBuilder().Build()
		l.setApiConfig(cfClientProperties, true)
		log.Debugf("Source api config: %+v", l.cfg.SourceApi)
	}
	return &l.cfg.SourceApi
}

func (l ConfigLoader) TargetApiConfig() *config.CloudController {
	if !l.cfg.TargetApi.IsSet() {
		log.Infoln("Target api config not set, so loading target config from opsman")
		cfClientProperties := l.propertiesProvider.TargetCFPropertiesBuilder().Build()
		l.setApiConfig(cfClientProperties, false)
		log.Debugf("Target api config: %+v", l.cfg.TargetApi)
	}
	return &l.cfg.TargetApi
}

func (l ConfigLoader) CCDBConfig(m string, toSource bool) interface{} {
	if toSource {
		log.Infoln("Loading source ccdb config from opsman")
		return l.SourceCCDBConfig(m)
	}
	log.Infoln("Loading target ccdb config from opsman")
	return l.TargetCCDBConfig(m)
}

func (l ConfigLoader) SourceCCDBConfig(m string) interface{} {
	if l.cfg.Migration.Migrators == nil {
		l.cfg.Migration.Migrators = make([]config.Migrator, 0)
	}

	log.Debugf("Getting cc config from %#v", l.cfg)
	if cfg, ok := l.getCloudControllerConfig(m); ok {
		if cfg.SourceCloudControllerDatabase.IsSet() {
			log.Debugf("Source cfg is already set to %#v", cfg)
			return &cfg
		}
		ccdbProperties := l.propertiesProvider.SourceCCDBPropertiesBuilder(l.propertiesProvider.SourceBoshPropertiesBuilder()).Build()
		l.setCloudControllerConfig(m, *ccdbProperties, true)
		log.Debugf("Getting config for migrator type: %v", m)
		if cfg, ok := lookupValue(l.cfg.Migration, m).(cc.Config); ok {
			log.Debugf("Returning source migrator config %+v for: %v", cfg, m)
			return &cfg
		}
	}

	log.Debugf("Did not find a cc config for %s from %#v", m, l.cfg)
	return nil
}

func (l ConfigLoader) TargetCCDBConfig(m string) interface{} {
	if l.cfg.Migration.Migrators == nil {
		l.cfg.Migration.Migrators = make([]config.Migrator, 0)
	}

	log.Debugf("Getting cc config from %#v", l.cfg)
	if cfg, ok := l.getCloudControllerConfig(m); ok {
		if cfg.TargetCloudControllerDatabase.IsSet() {
			log.Debugf("Target cfg is already set to %#v", cfg)
			return &cfg
		}
		ccdbProperties := l.propertiesProvider.TargetCCDBPropertiesBuilder(l.propertiesProvider.TargetBoshPropertiesBuilder()).Build()
		l.setCloudControllerConfig(m, *ccdbProperties, false)
		if cfg, ok := lookupValue(l.cfg.Migration, m).(cc.Config); ok {
			log.Debugf("Returning target migrator config %+v for: %v", cfg, m)
			return &cfg
		}
	}

	log.Debugf("Did not find a cc config for %s", m)
	return nil
}

func (l ConfigLoader) toSource(env config.EnvProperties) {
	l.setBoshConfig(env.BoshProperties, true)
	l.setApiConfig(env.CFProperties, true)
	l.setMigrationConfig(*env.CCDBProperties, true)
}

func (l ConfigLoader) toTarget(env config.EnvProperties) {
	l.setBoshConfig(env.BoshProperties, false)
	l.setApiConfig(env.CFProperties, false)
	l.setMigrationConfig(*env.CCDBProperties, false)
}

func (l ConfigLoader) setBoshConfig(properties *config.BoshProperties, toSource bool) {
	var boshConfig = &l.cfg.TargetBosh
	var opsman = l.cfg.Foundations.Target
	if toSource {
		boshConfig = &l.cfg.SourceBosh
		opsman = l.cfg.Foundations.Source
	}
	scheme, host, port, _, err := net.ParseURL(properties.URL)
	if err != nil {
		log.Fatalln(err)
	}
	boshConfig.AllProxy = getAllProxyURL(opsman)
	boshURL := fmt.Sprintf("%s://%s", scheme, host)
	if port != 443 && port != 0 {
		boshURL = fmt.Sprintf("%s://%s:%d", scheme, host, port)
	}
	boshConfig.URL = boshURL
	if len(properties.RootCA) > 0 {
		boshConfig.TrustedCert = properties.RootCA[len(properties.RootCA)-1].CertPEM
	}
	boshConfig.Authentication.UAA.URL = fmt.Sprintf("%s://%s:%d", scheme, host, 8443)
	boshConfig.Authentication.UAA.ClientCredentials.ID = properties.ClientID
	boshConfig.Authentication.UAA.ClientCredentials.Secret = properties.ClientSecret
	boshConfig.Deployment = properties.Deployment
}

func getAllProxyURL(opsman config.OpsManager) string {
	if err := opsman.ValidateSSH(); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("ssh+socks5://%s@%s:22?private-key=%s", opsman.SshUser, opsman.Hostname, opsman.PrivateKey)
}

func (l ConfigLoader) setApiConfig(properties *config.CFProperties, toSource bool) {
	var controller = &l.cfg.TargetApi
	if toSource {
		controller = &l.cfg.SourceApi
	}
	controller.URL = properties.URL
	controller.Username = properties.Username
	controller.Password = properties.Password
}

func (l ConfigLoader) setMigrationConfig(properties config.CCDBProperties, toSource bool) {
	for _, m := range l.cfg.Migration.Migrators {
		l.setCloudControllerConfig(m.Name, properties, toSource)
	}
}

func (l ConfigLoader) setCloudControllerConfig(configType string, properties config.CCDBProperties, toSource bool) {
	m := configType

	if toSource {
		if c, ok := lookupValue(l.cfg.Migration, m).(cc.Config); ok {
			l.setCCDBConfig(&c.SourceCloudControllerDatabase, properties)
			setMigrator(l.cfg.Migration, m, c)
		} else {
			if conf, ok := l.getCloudControllerConfig(m); ok {
				l.setCCDBConfig(&conf.SourceCloudControllerDatabase, properties)
				setMigrator(l.cfg.Migration, m, conf)
			}
		}
	} else {
		if c, ok := lookupValue(l.cfg.Migration, m).(cc.Config); ok {
			l.setCCDBConfig(&c.TargetCloudControllerDatabase, properties)
			setMigrator(l.cfg.Migration, m, c)
		} else {
			if conf, ok := l.getCloudControllerConfig(m); ok {
				l.setCCDBConfig(&conf.TargetCloudControllerDatabase, properties)
				setMigrator(l.cfg.Migration, m, conf)
			}
		}
	}
}

func (l ConfigLoader) getCloudControllerConfig(configType string) (cc.Config, bool) {
	var conf cc.Config
	m, err := l.migrationReader.GetMigration()
	if err != nil {
		log.Fatalln(err)
	}

	if !cc.IsCCDBTypeMigrator(configType) {
		return cc.Config{}, false
	}

	c := config.NewMapDecoder(conf).Decode(*m, configType)
	if err != nil {
		log.Fatalln(err)
	}

	if _, ok := c.(cc.Config); !ok {
		log.Fatalln(fmt.Errorf("could not load ccdb config for %s", configType))
	}

	return c.(cc.Config), true
}

func (l ConfigLoader) setCCDBConfig(database *cc.DatabaseConfig, properties config.CCDBProperties) {
	database.Host = properties.Host
	database.Username = properties.Username
	database.Password = properties.Password
	database.EncryptionKey = properties.EncryptionKey
	database.SSHHost = properties.SSHHost
	database.SSHUsername = properties.SSHUsername
	database.SSHPrivateKey = properties.SSHPrivateKey
	database.TunnelRequired = true
}

func lookupValue(migration config.Migration, key string) interface{} {
	if m, ok := lookupMigrator(migration, key); ok {
		return m.Value[key]
	}
	return nil
}

func lookupMigrator(migration config.Migration, key string) (*config.Migrator, bool) {
	m := config.LookupMigrator(migration, key)
	if m == nil {
		return nil, false
	}
	return m, true
}

func setMigrator(migration config.Migration, key string, value cc.Config) {
	for i, v := range migration.Migrators {
		if v.Name == key {
			migration.Migrators[i].Value = map[string]interface{}{key: value}
		}
	}
}
