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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"
	gourl "net/url"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
)

type ClientRequest struct {
	endpoint     string
	client       string
	clientSecret string
	httpClient   *httpclient.DefaultHTTPClient
}

func NewClientRequest(
	endpoint string,
	client string,
	clientSecret string,
	httpClient *httpclient.DefaultHTTPClient,
) ClientRequest {
	return ClientRequest{
		endpoint:     endpoint,
		client:       client,
		clientSecret: clientSecret,
		httpClient:   httpClient,
	}
}

func (r ClientRequest) Get(path string, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *gohttp.Request) {
		req.Header.Add("Accept", "application/json")
		req.SetBasicAuth(gourl.QueryEscape(r.client), gourl.QueryEscape(r.clientSecret))
	}

	resp, err := r.httpClient.GetCustomized(url, setHeaders)
	if err != nil {
		return fmt.Errorf("error performing request GET '%s': %w", url, err)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling UAA response: %w", err)
	}

	return nil
}

func (r ClientRequest) Post(path string, payload []byte, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *gohttp.Request) {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(gourl.QueryEscape(r.client), gourl.QueryEscape(r.clientSecret))
	}

	resp, err := r.httpClient.PostCustomized(url, payload, setHeaders)

	if err != nil {
		return fmt.Errorf("error performing request POST '%s': %w", url, err)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling UAA response: %w", err)
	}

	return nil
}

func (r ClientRequest) readResponse(resp *gohttp.Response) ([]byte, error) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading UAA response")
	}

	if resp.StatusCode != gohttp.StatusOK {
		msg := "UAA responded with non-successful status code '%d' response '%s'"
		return nil, fmt.Errorf(msg, resp.StatusCode, respBody)
	}

	return respBody, nil
}
