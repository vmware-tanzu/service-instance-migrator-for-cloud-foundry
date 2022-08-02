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
	"errors"
	"github.com/cloudfoundry/bosh-utils/crypto"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/net"
)

type Config struct {
	Scheme string
	Host   string
	Port   int
	Path   string

	CACert   string
	AllProxy string

	TokenFunc func(bool) (string, error)
}

func NewConfigFromURL(url string) (Config, error) {
	scheme, host, port, path, err := net.ParseURL(url)
	if err != nil {
		return Config{}, err
	}

	return Config{Scheme: scheme, Host: host, Port: port, Path: path}, nil
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return errors.New("missing 'Host'")
	}

	if c.Port == 0 {
		return errors.New("missing 'Port'")
	}

	if _, err := c.CACertPool(); err != nil {
		return err
	}

	return nil
}

func (c Config) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	return crypto.CertPoolFromPEM([]byte(c.CACert))
}
