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

package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	ConfigDir         string
	ConfigFile        string
	Debug             bool
	DryRun            bool `mapstructure:"dry_run"`
	DomainsToReplace  map[string]string
	ExportDir         string   `mapstructure:"export_dir"`
	ExcludedOrgs      []string `mapstructure:"exclude_orgs"`
	IncludedOrgs      []string `mapstructure:"include_orgs"`
	IgnoreServiceKeys bool     `mapstructure:"ignore_service_keys"`
	Foundations       struct {
		Source OpsManager `yaml:"source"`
		Target OpsManager `yaml:"target"`
	} `yaml:"foundations"`
	Migration   Migration
	Name        string
	Services    []string        `mapstructure:"services"`
	Instances   []string        `mapstructure:"instances"`
	SourceApi   CloudController `yaml:"source_api" mapstructure:"source_api"`
	SourceBosh  Bosh            `yaml:"source_bosh" mapstructure:"source_bosh"`
	TargetApi   CloudController `yaml:"target_api" mapstructure:"target_api"`
	TargetBosh  Bosh            `yaml:"target_bosh" mapstructure:"target_bosh"`
	initialized bool
}

type CloudController struct {
	URL          string `yaml:"url" mapstructure:"url"`
	Username     string `yaml:"username" mapstructure:"username"`
	Password     string `yaml:"password" mapstructure:"password"`
	ClientID     string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
}

type Migration struct {
	UseDefaultMigrator bool       `mapstructure:"use_default_migrator"`
	Migrators          []Migrator `yaml:"migrators"`
}

type Migrator struct {
	Name  string                 `yaml:"name" mapstructure:"name"`
	Value map[string]interface{} `yaml:"migrator" mapstructure:"migrator"`
}

type OpsManager struct {
	URL          string `yaml:"url"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	ClientID     string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
	Hostname     string `yaml:"hostname,omitempty" mapstructure:"hostname,omitempty"`
	IP           string `yaml:"ip,omitempty" mapstructure:"ip,omitempty"`
	PrivateKey   string `yaml:"private_key,omitempty" mapstructure:"private_key,omitempty"`
	SshUser      string `yaml:"ssh_user,omitempty" mapstructure:"ssh_user,omitempty"`
}

//counterfeiter:generate -o fakes . Loader

type Loader interface {
	BoshConfig(toSource bool) *Bosh
	SourceBoshConfig() *Bosh
	TargetBoshConfig() *Bosh
	CFConfig(toSource bool) *CloudController
	SourceApiConfig() *CloudController
	TargetApiConfig() *CloudController
	CCDBConfig(m string, toSource bool) interface{}
	SourceCCDBConfig(m string) interface{}
	TargetCCDBConfig(m string) interface{}
}

//counterfeiter:generate -o fakes . BoshPropertiesBuilder

type BoshPropertiesBuilder interface {
	Build() *BoshProperties
}

//counterfeiter:generate -o fakes . CFPropertiesBuilder

type CFPropertiesBuilder interface {
	Build() *CFProperties
}

//counterfeiter:generate -o fakes . CCDBPropertiesBuilder

type CCDBPropertiesBuilder interface {
	Build() *CCDBProperties
}

//counterfeiter:generate -o fakes . PropertiesProvider

type PropertiesProvider interface {
	Environment(BoshPropertiesBuilder, CFPropertiesBuilder, CCDBPropertiesBuilder) EnvProperties
	SourceBoshPropertiesBuilder() BoshPropertiesBuilder
	TargetBoshPropertiesBuilder() BoshPropertiesBuilder
	SourceCFPropertiesBuilder() CFPropertiesBuilder
	TargetCFPropertiesBuilder() CFPropertiesBuilder
	SourceCCDBPropertiesBuilder(b BoshPropertiesBuilder) CCDBPropertiesBuilder
	TargetCCDBPropertiesBuilder(b BoshPropertiesBuilder) CCDBPropertiesBuilder
}

type EnvProperties struct {
	*BoshProperties
	*CFProperties
	*CCDBProperties
}

type BoshProperties struct {
	URL, AllProxy, ClientID, ClientSecret string
	RootCA                                []httpclient.CA
	Deployment                            string
}

type CFProperties struct {
	URL, Username, Password string
}

type CCDBProperties struct {
	Host, Username, Password, EncryptionKey string
	SSHHost, SSHUsername, SSHPrivateKey     string
}

func NewDefaultConfig() (*Config, error) {
	configDir := ""

	if cfgHome, ok := os.LookupEnv("SI_MIGRATOR_CONFIG_HOME"); ok {
		configDir = cfgHome
	}

	if configFile, ok := os.LookupEnv("SI_MIGRATOR_CONFIG_FILE"); ok {
		return New(configDir, configFile), nil
	}

	if _, ok := hasSuffix(configDir); ok {
		configFile := configDir
		configDir, _ = filepath.Split(configFile)
		return New(configDir, configFile), nil
	}

	return New(configDir, ""), nil
}

func New(configDir string, configFile string) *Config {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	c := &Config{
		Name:       "si-migrator",
		ConfigDir:  configDir,
		ConfigFile: configFile,
		ExportDir:  path.Join(cwd, "export"),
	}

	c.initViperConfig()
	cobra.OnInitialize(c.initLogger, c.initViperConfig)

	return c
}

func Parse(configFilePath string) (Config, error) {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(configFileBytes, &cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if err := c.SourceApi.Validate(); err != nil {
		return fmt.Errorf("source cf api configuration error: %s", err)
	}
	if err := c.TargetApi.Validate(); err != nil {
		return fmt.Errorf("target cf api configuration error: %s", err)
	}
	if err := c.SourceBosh.Validate(); err != nil {
		return fmt.Errorf("source bosh configuration error: %s", err)
	}
	if err := c.TargetBosh.Validate(); err != nil {
		return fmt.Errorf("target bosh configuration error: %s", err)
	}
	if err := c.Foundations.Source.Validate(); err != nil {
		return err
	}
	if err := c.Foundations.Target.Validate(); err != nil {
		return err
	}
	return nil
}

func (c CloudController) Validate() error {
	if c.URL == "" {
		return NewFieldError("cf url", errors.New("can't be empty"))
	}
	if c.Username == "" && c.ClientID == "" {
		return NewFieldsError([]string{"cf username", "client_id"}, errors.New("can't be empty"))
	}
	if c.Password == "" && c.ClientSecret == "" {
		return NewFieldsError([]string{"cf password", "client_secret"}, errors.New("can't be empty"))
	}
	return nil
}

func (c CloudController) IsSet() bool {
	return c != CloudController{}
}

func (c OpsManager) Validate() error {
	if c.URL == "" {
		return NewFieldError("ops manager url", errors.New("can't be empty"))
	}
	if c.Username == "" && c.ClientID == "" {
		return NewFieldsError([]string{"ops manager username", "client_id"}, errors.New("can't be empty"))
	}
	if c.Password == "" && c.ClientSecret == "" {
		return NewFieldsError([]string{"ops manager password", "client_secret"}, errors.New("can't be empty"))
	}
	return nil
}

func (c OpsManager) ValidateSSH() error {
	if c.SshUser == "" {
		return NewFieldError("ops manager ssh user", errors.New("can't be empty"))
	}
	if c.Hostname == "" {
		return NewFieldError("ops manager hostname", errors.New("can't be empty"))
	}
	if c.PrivateKey == "" {
		return NewFieldError("ops manager private-key", errors.New("can't be empty"))
	}
	return nil
}

/// initConfig reads in config file and ENV variables if set.
func (c *Config) initViperConfig() {
	v := viper.New()
	if c.ConfigFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(c.ConfigFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.WithError(err).Error("failed to find home directory")
			os.Exit(1)
		}

		v.SetConfigName(c.Name)
		v.SetConfigType("yaml")
		if c.ConfigDir != "" {
			// Add ConfigDir to the search path
			v.AddConfigPath(c.ConfigDir)
		}
		// Search config in home directory
		v.AddConfigPath(home)
		v.AddConfigPath(fmt.Sprintf("$HOME/.config/%s", c.Name)) // optionally look for config in the XDG_CONFIG_HOME
		v.AddConfigPath(".")                                     // finally, look in the working directory
	}
	v.SetEnvPrefix(c.Name)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Errorf("failed to load config file: %s, error: %s", v.ConfigFileUsed(), err)
			panic(err)
		} else {
			log.Errorf("failed to read config, error: %s", err)
			panic(err)
		}
	}

	if !c.initialized {
		c.applyViperOverrides(v)
	}
	c.initialized = true
}

func (c *Config) applyViperOverrides(v *viper.Viper) {
	err := v.Unmarshal(c)
	if err != nil {
		panic(err)
	}
	// have to explicitly convert map[string]string
	if len(v.GetStringMapString("domains_to_replace")) > 0 {
		c.DomainsToReplace = v.GetStringMapString("domains_to_replace")
	}
	if c.ConfigFile == "" {
		c.ConfigFile = v.ConfigFileUsed()
	}
}

func hasSuffix(configDir string) (string, bool) {
	if strings.HasSuffix(configDir, "yml") {
		return "yml", true
	}
	if strings.HasSuffix(configDir, "yaml") {
		return "yaml", true
	}
	return "", false
}
