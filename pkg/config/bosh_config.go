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
	"github.com/pkg/errors"
)

type Bosh struct {
	URL            string `yaml:"url" mapstructure:"url"`
	AllProxy       string `yaml:"all_proxy" mapstructure:"all_proxy"`
	TrustedCert    string `yaml:"root_ca_cert" mapstructure:"root_ca_cert"`
	Authentication Authentication
	Deployment     string
}

type Authentication struct {
	Basic UserCredentials
	UAA   UAAAuthentication
}

func (a Authentication) Validate(URLRequired bool) error {
	uaaSet := a.UAA.IsSet()
	basicSet := a.Basic.IsSet()

	switch {
	case !uaaSet && !basicSet:
		return fmt.Errorf("must specify an authentication type")
	case uaaSet && basicSet:
		return fmt.Errorf("cannot specify both basic and UAA authentication")
	case uaaSet:
		if err := a.UAA.Validate(URLRequired); err != nil {
			return NewFieldError("authentication.uaa", err)
		}
	case basicSet:
		if err := a.Basic.Validate(); err != nil {
			return NewFieldError("authentication.basic", err)
		}
	}
	return nil
}

func (a Authentication) IsSet() bool {
	uaaIsSet := a.UAA.IsSet()
	basicSet := a.Basic.IsSet()

	return uaaIsSet || basicSet
}

type UAAAuthentication struct {
	URL               string
	ClientCredentials ClientCredentials `yaml:"client_credentials" mapstructure:"client_credentials"`
	UserCredentials   UserCredentials   `yaml:"user_credentials" mapstructure:"user_credentials"`
}

func (a UAAAuthentication) IsSet() bool {
	clientCredentialsSet := a.ClientCredentials.IsSet()
	userCredentialsSet := a.UserCredentials.IsSet()

	return clientCredentialsSet || userCredentialsSet
}

type UserCredentials struct {
	Username string
	Password string
}

func (c UserCredentials) IsSet() bool {
	return c != UserCredentials{}
}

func (c UserCredentials) Validate() error {
	if c.Username == "" {
		return NewFieldError("username", errors.New("can't be empty"))
	}
	if c.Password == "" {
		return NewFieldError("password", errors.New("can't be empty"))
	}
	return nil
}

type ClientCredentials struct {
	ID     string `yaml:"client_id" mapstructure:"client_id"`
	Secret string `yaml:"client_secret" mapstructure:"client_secret"`
}

func (cc ClientCredentials) IsSet() bool {
	return cc != ClientCredentials{}
}

func (cc ClientCredentials) Validate() error {
	if cc.ID == "" {
		return NewFieldError("client_id", errors.New("can't be empty"))
	}
	if cc.Secret == "" {
		return NewFieldError("client_secret", errors.New("can't be empty"))
	}
	return nil
}

func (b Bosh) Validate() error {
	if b.URL == "" {
		return fmt.Errorf("must specify bosh url")
	}
	if b.AllProxy == "" {
		return fmt.Errorf("must specify bosh all_proxy")
	}

	return b.Authentication.Validate(false)
}

func (b Bosh) IsSet() bool {
	return b != Bosh{}
}

func (a UAAAuthentication) Validate(URLRequired bool) error {
	return validateAuthenticationFields(URLRequired, a.URL, a.ClientCredentials, a.UserCredentials)
}

func validateAuthenticationFields(URLRequired bool, url string, clientCredentials ClientCredentials, userCredentials UserCredentials) error {
	urlIsSet := url != ""
	clientCredentialsSet := clientCredentials.IsSet()
	userCredentialsSet := userCredentials.IsSet()

	switch {
	case !urlIsSet && !clientCredentialsSet && !userCredentialsSet:
		return fmt.Errorf("must specify UAA authentication")
	case !urlIsSet && URLRequired:
		return NewFieldError("uaa url", errors.New("can't be empty"))
	case !clientCredentialsSet && !userCredentialsSet:
		return fmt.Errorf("authentication should contain either user_credentials or client_credentials")
	case clientCredentialsSet && userCredentialsSet:
		return fmt.Errorf("contains both client and user credentials")
	case clientCredentialsSet:
		if err := clientCredentials.Validate(); err != nil {
			return NewFieldError("client_credentials", err)
		}
	case userCredentialsSet:
		if err := userCredentials.Validate(); err != nil {
			return NewFieldError("user_credentials", err)
		}
	}
	return nil
}
