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
	"fmt"
)

type ClientTokenSession struct {
	uaa        UAA
	token      AccessToken
	parameters []TokenParameters
}

type TokenParameters struct {
	Key   string // e.g. "username"
	Value string
}

type Options func(*ClientTokenSession)

func NewClientTokenSession(uaa UAA, opts ...Options) *ClientTokenSession {
	session := &ClientTokenSession{uaa: uaa}

	for _, o := range opts {
		o(session)
	}

	return session
}

func WithPasswordCredentials(username, password string) Options {
	return func(session *ClientTokenSession) {
		parameters := []TokenParameters{
			{Key: "username", Value: username},
			{Key: "password", Value: password},
		}
		session.parameters = parameters
	}
}

func (c *ClientTokenSession) ClientCredentialsTokenFunc(retried bool) (string, error) {
	if c.token == nil || retried {
		token, err := c.uaa.ClientCredentialsGrant()
		if err != nil {
			return "", err
		}

		c.token = token
	}
	return c.token.Type() + " " + c.token.Value(), nil
}

func (c *ClientTokenSession) OwnerPasswordCredentialsTokenFunc(retried bool) (string, error) {
	if c.token == nil || retried {
		token, err := c.uaa.OwnerPasswordCredentialsGrant(c.parameters)
		if err != nil {
			return "", fmt.Errorf("failed to grant access to user: %w", err)
		}

		c.token = token
	}
	return c.token.Type() + " " + c.token.Value(), nil
}
