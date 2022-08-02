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
	gourl "net/url"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
)

type Client struct {
	clientRequest ClientRequest
}

type TokenResp struct {
	Type         string `json:"token_type"`    // e.g. "bearer"
	AccessToken  string `json:"access_token"`  // e.g. "eyJhbGciOiJSUzI1NiJ9.eyJq<snip>fQ.Mr<snip>RawG"
	RefreshToken string `json:"refresh_token"` // e.g. "eyJhbGciOiJSUzI1NiJ9.eyJq<snip>fQ.Mr<snip>RawG"
}

func NewClient(endpoint string, clientID string, clientSecret string, httpClient *httpclient.DefaultHTTPClient) Client {
	return Client{NewClientRequest(endpoint, clientID, clientSecret, httpClient)}
}

func (c Client) ClientCredentialsGrant() (TokenResp, error) {
	query := gourl.Values{}

	query.Add("grant_type", "client_credentials")

	var resp TokenResp
	err := c.clientRequest.Post("/oauth/token", []byte(query.Encode()), &resp)
	if err != nil {
		return resp, fmt.Errorf("requesting token via Client credentials grant: %w", err)
	}

	return resp, nil
}

func (c Client) OwnerPasswordCredentialsGrant(params []TokenParameters) (TokenResp, error) {
	query := gourl.Values{}

	query.Add("grant_type", "password")

	for _, param := range params {
		query.Add(param.Key, param.Value)
	}

	var resp TokenResp

	err := c.clientRequest.Post("/oauth/token", []byte(query.Encode()), &resp)
	if err != nil {
		return resp, fmt.Errorf("requesting token via password credentials grant: %w", err)
	}

	return resp, nil
}

func (c Client) RefreshTokenGrant(refreshValue string) (TokenResp, error) {
	query := gourl.Values{}

	query.Add("grant_type", "refresh_token")
	query.Add("refresh_token", refreshValue)

	var resp TokenResp

	err := c.clientRequest.Post("/oauth/token", []byte(query.Encode()), &resp)
	if err != nil {
		return resp, fmt.Errorf("error refreshing token: %w", err)
	}

	return resp, nil
}
