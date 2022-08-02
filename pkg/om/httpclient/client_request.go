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
	"encoding/json"
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type ClientRequest struct {
	endpoint   string
	httpClient *httpclient.DefaultHTTPClient
}

func NewClientRequest(
	endpoint string,
	httpClient *httpclient.DefaultHTTPClient,
) ClientRequest {
	return ClientRequest{
		endpoint:   endpoint,
		httpClient: httpClient,
	}
}

func (r ClientRequest) Get(path string, response interface{}) error {
	respBody, _, err := r.RawGet(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return nil
}

func (r ClientRequest) RawGet(path string) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("performing request GET '%s': %w", url, err)
	}

	return r.readResponse(resp)
}

func (r ClientRequest) readResponse(resp *http.Response) ([]byte, *http.Response, error) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var respBody []byte

	if resp.Request != nil {
		b, err := httputil.DumpRequest(resp.Request, true)
		if err == nil {
			log.Debugln("Dumping client request:")
			log.Debugf("%s", string(b))
		}
	}

	b, err := httputil.DumpResponse(resp, true)
	if err == nil {
		log.Debugln("Dumping client response:")
		log.Debugf("%s", string(b))
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response: %w", err)
	}

	not200 := resp.StatusCode != http.StatusOK
	not201 := resp.StatusCode != http.StatusCreated
	not204 := resp.StatusCode != http.StatusNoContent
	not206 := resp.StatusCode != http.StatusPartialContent
	not302 := resp.StatusCode != http.StatusFound

	if not200 && not201 && not204 && not206 && not302 {
		msg := "server responded with non-successful status code '%d' response '%s'"
		return respBody, resp, fmt.Errorf(msg, resp.StatusCode, respBody)
	}

	return respBody, resp, nil
}
