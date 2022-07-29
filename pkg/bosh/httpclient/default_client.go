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

package httpclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type DefaultHTTPClient struct {
	client Client
	logTag string
	opts   Opts
}

type Opts struct {
	NoRedactUrlQuery bool
}

func NewHTTPClient(client Client) *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: client,
		logTag: "httpClient",
	}
}

func NewHTTPClientOpts(client Client, opts Opts) *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: client,
		logTag: "httpClient",
		opts:   opts,
	}
}

func (c *DefaultHTTPClient) Post(endpoint string, payload []byte) (*http.Response, error) {
	return c.PostCustomized(endpoint, payload, nil)
}

func (c *DefaultHTTPClient) PostCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error) {
	postPayload := strings.NewReader(string(payload))

	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	log.Debugf("Sending POST request to endpoint '%s'", redactedEndpoint)

	request, err := http.NewRequest("POST", endpoint, postPayload)
	if err != nil {
		return nil, fmt.Errorf("error creating POST request: %w", err)
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error performing POST request: %w", scrubErrorOutput(err))
	}

	return response, nil
}

func (c *DefaultHTTPClient) Put(endpoint string, payload []byte) (*http.Response, error) {
	return c.PutCustomized(endpoint, payload, nil)
}

func (c *DefaultHTTPClient) PutCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error) {
	putPayload := strings.NewReader(string(payload))

	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	log.Debugf("Sending PUT request to endpoint '%s'", redactedEndpoint)

	request, err := http.NewRequest("PUT", endpoint, putPayload)
	if err != nil {
		return nil, fmt.Errorf("error creating PUT request: %w", err)
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error performing PUT request: %w", scrubErrorOutput(err))
	}

	return response, nil
}

func (c *DefaultHTTPClient) Get(endpoint string) (*http.Response, error) {
	return c.GetCustomized(endpoint, nil)
}

func (c *DefaultHTTPClient) GetCustomized(endpoint string, f func(*http.Request)) (*http.Response, error) {
	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	log.Debugf("Sending GET request to endpoint '%s'", redactedEndpoint)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request: %w", scrubErrorOutput(err))
	}

	return response, nil
}

func (c *DefaultHTTPClient) Delete(endpoint string) (*http.Response, error) {
	return c.DeleteCustomized(endpoint, nil)
}

func (c *DefaultHTTPClient) DeleteCustomized(endpoint string, f func(*http.Request)) (*http.Response, error) {
	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	log.Debugf("Sending DELETE request with endpoint %s", redactedEndpoint)

	request, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating DELETE request: %w", err)
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error performing DELETE request: %w", err)
	}
	return response, nil
}

var scrubUserinfoRegex = regexp.MustCompile("(https?://.*:).*@")

func scrubEndpointQuery(endpoint string) string {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "error occurred parsing endpoint"
	}

	query := parsedURL.Query()
	for key := range query {
		query[key] = []string{"<redacted>"}
	}

	parsedURL.RawQuery = query.Encode()

	unescapedEndpoint, _ := url.QueryUnescape(parsedURL.String())
	return unescapedEndpoint
}

func scrubErrorOutput(err error) error {
	errorMsg := err.Error()
	errorMsg = scrubUserinfoRegex.ReplaceAllString(errorMsg, "$1<redacted>@")

	return errors.New(errorMsg)
}
