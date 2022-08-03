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
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	sio "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type YAMLMigrationReader struct {
	migrationBytes []byte
	migration      *Migration
	filename       string
	mu             sync.Mutex
}

func NewMigrationReader(cfg *Config) (*YAMLMigrationReader, error) {
	if len(cfg.Migration.Migrators) == 0 {
		return ReadFromDir(cfg.ConfigDir)
	}
	return &YAMLMigrationReader{
		migration: &cfg.Migration,
		filename:  cfg.ConfigFile,
	}, nil
}

func ReadFrom(r io.Reader) (*YAMLMigrationReader, error) {
	buf := &strings.Builder{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	return ReadFromDir(buf.String())
}

func ReadFromDir(path string) (*YAMLMigrationReader, error) {
	b, filename, err := readLegacyConfig(path)
	if err != nil {
		return nil, err
	}

	var r *YAMLMigrationReader

	if len(b) == 0 {
		cfg := NewDefaultConfig()
		r = &YAMLMigrationReader{
			migration:      &cfg.Migration,
			migrationBytes: b,
			filename:       cfg.ConfigFile,
		}
		_, err = r.write()
		if err != nil {
			return nil, err
		}

		return r, nil
	}

	r = &YAMLMigrationReader{
		migrationBytes: b,
		filename:       filename,
	}

	_, err = r.read()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *YAMLMigrationReader) GetMigration() (*Migration, error) {
	if r.migration != nil {
		return r.migration, nil
	}

	_, err := r.read()
	if err != nil {
		return nil, err
	}

	return r.migration, nil
}

func (r *YAMLMigrationReader) GetString() (string, error) {
	if len(r.migrationBytes) == 0 {
		if r.migration != nil {
			_, err := r.write()
			if err != nil {
				return "", err
			}
		}
		return string(r.migrationBytes), nil
	}

	return string(r.migrationBytes), nil
}

func (r *YAMLMigrationReader) read() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := &Migration{}
	err := yaml.Unmarshal(r.migrationBytes, m)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal yaml from file: %s, %w", r.filename, err)
	}

	r.migration = m

	return len(r.migrationBytes), err
}

func (r *YAMLMigrationReader) write() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	bytes, err := yaml.Marshal(r.migration)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal yaml to file: %s, %w", r.filename, err)
	}

	r.migrationBytes = bytes

	return len(r.migrationBytes), err
}

func readLegacyConfig(configDir string) ([]byte, string, error) {
	if configDir == "" {
		if cfgHome, ok := os.LookupEnv("MIGRATOR_CONFIG_HOME"); ok {
			if isYamlFile(cfgHome) {
				return readConfigFile(cfgHome)
			}

			return readConfigDir(cfgHome)
		}

		if configFile, ok := os.LookupEnv("MIGRATOR_CONFIG_FILE"); ok {
			return readConfigFile(configFile)
		}
	}

	return readConfigDir(configDir)
}

func readConfigDir(configDir string) ([]byte, string, error) {
	if configDir == "" {
		dir, err := homedir.Dir()
		if err != nil {
			return nil, dir, errors.Wrap(err, "unable to find user home directory")
		}
		configDir = filepath.Join(dir, ".config", "si-migrator")
	}

	if isYamlFile(configDir) {
		return readConfigFile(configDir)
	}

	return readConfigFile(filepath.Join(configDir, "config.yml"))
}

func readConfigFile(configFile string) ([]byte, string, error) {
	_, err := sio.CreateFileIfNotExist(configFile)
	if err != nil {
		return nil, configFile, err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, configFile, errors.Wrap(err, fmt.Sprintf("cannot open file: %s", configFile))
	}

	return data, configFile, nil
}

func isYamlFile(configDir string) bool {
	if strings.HasSuffix(configDir, "yml") {
		return true
	}
	if strings.HasSuffix(configDir, "yaml") {
		return true
	}
	return false
}
