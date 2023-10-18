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

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type ClientRequest struct {
	endpoint     string
	contextId    string
	httpClient   *httpclient.DefaultHTTPClient
	fileReporter director.FileReporter
}

func NewClientRequest(
	endpoint string,
	httpClient *httpclient.DefaultHTTPClient,
	fileReporter director.FileReporter,
) ClientRequest {
	return ClientRequest{
		endpoint:     endpoint,
		httpClient:   httpClient,
		fileReporter: fileReporter,
	}
}

func (r ClientRequest) WithContext(contextId string) ClientRequest {
	// returns a copy of the ClientRequest
	r.contextId = contextId
	return r
}

func (r ClientRequest) Get(path string, response interface{}) error {
	respBody, _, err := r.RawGet(path, nil, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling Director response: %w", err)
	}

	return nil
}

func (r ClientRequest) Post(path string, payload []byte, f func(*http.Request), response interface{}) error {
	respBody, _, err := r.RawPost(path, payload, f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling Director response: %w", err)
	}

	return nil
}

func (r ClientRequest) Put(path string, payload []byte, f func(*http.Request), response interface{}) error {
	respBody, _, err := r.RawPut(path, payload, f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling Director response: %w", err)
	}

	return nil
}

func (r ClientRequest) Delete(path string, response interface{}) error {
	respBody, _, err := r.RawDelete(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling Director response: %w", err)
	}

	return nil
}

func (r ClientRequest) RawGet(path string, out io.Writer, f func(*http.Request)) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := r.setContextIDHeader(f)

	resp, err := r.httpClient.GetCustomized(url, wrapperFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("error performing request GET '%s': %w", url, err)
	}

	return r.readResponse(resp, out)
}

// RawPost follows redirects via GET unlike generic HTTP clients
func (r ClientRequest) RawPost(path string, payload []byte, f func(*http.Request)) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := func(req *http.Request) {
		if f != nil {
			f(req)
		}

		isArchive := req.Header.Get("content-type") == "application/x-compressed"

		if isArchive && req.ContentLength > 0 && req.Body != nil {
			req.Body = r.fileReporter.TrackUpload(req.ContentLength, req.Body)
		}
	}

	wrapperFunc = r.setContextIDHeader(wrapperFunc)

	resp, err := r.httpClient.PostCustomized(url, payload, wrapperFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("error performing request POST '%s': %w", url, err)
	}

	return r.optionallyFollowResponse(url, resp)
}

// RawPut follows redirects via GET unlike generic HTTP clients
func (r ClientRequest) RawPut(path string, payload []byte, f func(*http.Request)) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := r.setContextIDHeader(f)

	resp, err := r.httpClient.PutCustomized(url, payload, wrapperFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("error performing request PUT '%s': %w", url, err)
	}

	return r.optionallyFollowResponse(url, resp)
}

// RawDelete follows redirects via GET unlike generic HTTP clients
func (r ClientRequest) RawDelete(path string) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := r.setContextIDHeader(nil)

	resp, err := r.httpClient.DeleteCustomized(url, wrapperFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("error performing request DELETE '%s': %w", url, err)
	}

	return r.optionallyFollowResponse(url, resp)
}

func (r ClientRequest) setContextIDHeader(f func(*http.Request)) func(*http.Request) {
	return func(req *http.Request) {
		if f != nil {
			f(req)
		}
		if r.contextId != "" {
			req.Header.Set("X-Bosh-Context-Id", r.contextId)
		}
	}
}

func (r ClientRequest) optionallyFollowResponse(url string, resp *http.Response) ([]byte, *http.Response, error) {
	body, resp, err := r.readResponse(resp, nil)
	if err != nil {
		return body, resp, err
	}

	// Follow redirect via GET
	if resp != nil && resp.StatusCode == http.StatusFound {
		redirectURL, err := resp.Location()
		if err != nil || redirectURL == nil {
			return body, resp, fmt.Errorf("error getting Location header from POST '%s': %w", url, err)
		}

		return r.RawGet(redirectURL.Path, nil, nil)
	}

	return body, resp, nil
}

type ShouldTrackDownload interface {
	ShouldTrackDownload() bool
}

func (r ClientRequest) readResponse(resp *http.Response, out io.Writer) ([]byte, *http.Response, error) {
	defer resp.Body.Close()

	var respBody []byte

	if out == nil {
		if resp.Request != nil {
			sanitizer := director.RequestSanitizer{Request: (*resp.Request)}
			sanitizedRequest, _ := sanitizer.SanitizeRequest()
			b, err := httputil.DumpRequest(&sanitizedRequest, true)
			if err == nil {
				log.Debugln("Dumping Director client request:")
				log.Debugf("%s", string(b))
			}
		}

		b, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Debugf("Dumping Director client response:\n%s", string(b))
		}

		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading Director response: %w", err)
		}
	}

	not200 := resp.StatusCode != http.StatusOK
	not201 := resp.StatusCode != http.StatusCreated
	not204 := resp.StatusCode != http.StatusNoContent
	not206 := resp.StatusCode != http.StatusPartialContent
	not302 := resp.StatusCode != http.StatusFound

	if not200 && not201 && not204 && not206 && not302 {
		msg := "Director responded with non-successful status code '%d' response '%s'"
		return respBody, resp, fmt.Errorf(msg, resp.StatusCode, respBody)
	}

	if out != nil {
		showProgress := true

		if typedOut, ok := out.(ShouldTrackDownload); ok {
			showProgress = typedOut.ShouldTrackDownload()
		}

		if showProgress {
			out = r.fileReporter.TrackDownload(resp.ContentLength, out)
		}

		_, err := io.Copy(out, resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("error copying Director response: %w", err)
		}
	}

	return respBody, resp, nil
}
