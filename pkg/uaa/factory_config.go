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

package uaa

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	gourl "net/url"
	"strconv"
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"

	"github.com/cloudfoundry/bosh-utils/crypto"
)

type Config struct {
	Host string
	Port int
	Path string

	ClientID     string
	ClientSecret string

	Username string
	Password string

	CACert   string
	AllProxy string

	TokenFunc func(bool) (string, error)
}

func NewConfigFromURL(url string) (Config, error) {
	if len(url) == 0 {
		return Config{}, errors.New("expected non-empty URL")
	}

	parsedURL, err := gourl.Parse(url)
	if err != nil {
		return Config{}, fmt.Errorf(fmt.Sprintf("Parsing URL '%s'", url)+": %w", err)
	}

	host := parsedURL.Host
	port := 443
	path := parsedURL.Path

	if len(host) == 0 {
		host = url
		path = ""
	}

	if strings.Contains(host, ":") {
		var portStr string

		host, portStr, err = net.SplitHostPort(host)
		if err != nil {
			return Config{}, fmt.Errorf(fmt.Sprintf("Extracting host/port from URL '%s'", url)+": %w", err)
		}

		port, err = strconv.Atoi(portStr)
		if err != nil {
			return Config{}, fmt.Errorf(fmt.Sprintf("Extracting port from URL '%s'", url)+": %w", err)
		}
	}

	if len(host) == 0 {
		return Config{}, fmt.Errorf("expected to extract host from URL '%s'", url)
	}

	return Config{Host: host, Port: port, Path: path}, nil
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return errors.New("missing 'Host'")
	}

	if c.Port == 0 {
		return errors.New("missing 'Port'")
	}

	if c.Username == "" && c.ClientID == "" {
		return config.NewFieldsError([]string{"uaa username", "client_id"}, errors.New("can't both be empty"))
	}

	if c.Password == "" && c.ClientSecret == "" {
		return config.NewFieldsError([]string{"uaa password", "client_secret"}, errors.New("can't both be empty"))
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
