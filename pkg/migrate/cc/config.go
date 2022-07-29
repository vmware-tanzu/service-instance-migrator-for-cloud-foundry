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

package cc

import (
	"errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

var ccdbTypeMigrators = []string{"ecs", "sqlserver"}

type Config struct {
	SourceCloudControllerDatabase DatabaseConfig `yaml:"source_ccdb"`
	TargetCloudControllerDatabase DatabaseConfig `yaml:"target_ccdb"`
}

type DatabaseConfig struct {
	Host           string `yaml:"db_host"`
	Username       string `yaml:"db_username"`
	Password       string `yaml:"db_password"`
	EncryptionKey  string `yaml:"db_encryption_key"`
	SSHHost        string `yaml:"ssh_host"`
	SSHUsername    string `yaml:"ssh_username"`
	SSHPassword    string `yaml:"ssh_password"`
	SSHPrivateKey  string `yaml:"ssh_private_key"`
	TunnelRequired bool   `yaml:"ssh_tunnel"`
}

func IsCCDBTypeMigrator(configType string) bool {
	for _, m := range ccdbTypeMigrators {
		if m == configType {
			return true
		}
	}
	return false
}

func (c Config) Validate(isSource bool) error {
	if isSource {
		if err := c.SourceCloudControllerDatabase.Validate(); err != nil {
			return err
		}
		return nil
	}

	if err := c.TargetCloudControllerDatabase.Validate(); err != nil {
		return err
	}
	return nil
}

func (c Config) IsSet() bool {
	if c.SourceCloudControllerDatabase.IsSet() {
		return true
	}
	if c.TargetCloudControllerDatabase.IsSet() {
		return true
	}
	return false
}

func (c DatabaseConfig) Validate() error {
	if c.Host == "" {
		return config.NewFieldError("ccdb host", errors.New("can't be empty"))
	}
	if c.Username == "" {
		return config.NewFieldError("ccdb username", errors.New("can't be empty"))
	}
	if c.Password == "" {
		return config.NewFieldError("ccdb password", errors.New("can't be empty"))
	}
	if c.EncryptionKey == "" {
		return config.NewFieldError("ccdb encryption key", errors.New("can't be empty"))
	}
	if c.TunnelRequired {
		if c.SSHHost == "" {
			return config.NewFieldError("ssh host", errors.New("can't be empty"))
		}
		if c.SSHUsername == "" {
			return config.NewFieldError("ssh username", errors.New("can't be empty"))
		}
		if c.SSHPassword == "" && c.SSHPrivateKey == "" {
			return config.NewFieldsError([]string{"ssh password", "ssh private key"}, errors.New("can't be empty"))
		}
	}
	return nil
}

func (c DatabaseConfig) IsSet() bool {
	return c != DatabaseConfig{}
}
