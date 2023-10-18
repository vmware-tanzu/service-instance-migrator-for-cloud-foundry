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
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type RequestRetryable struct {
	request   *http.Request
	requestID string
	delegate  Client
	attempt   int

	originalBody io.ReadCloser // buffer request body to memory for retries
	response     *http.Response

	uuidGenerator         boshuuid.Generator
	isResponseAttemptable func(*http.Response, error) (bool, error)
}

func NewRequestRetryable(
	request *http.Request,
	delegate Client,
	isResponseAttemptable func(*http.Response, error) (bool, error),
) *RequestRetryable {
	if isResponseAttemptable == nil {
		isResponseAttemptable = defaultIsAttemptable
	}

	return &RequestRetryable{
		request:               request,
		delegate:              delegate,
		attempt:               0,
		uuidGenerator:         boshuuid.NewGenerator(),
		isResponseAttemptable: isResponseAttemptable,
	}
}

func (r *RequestRetryable) Attempt() (bool, error) {
	var err error

	if r.requestID == "" {
		r.requestID, err = r.uuidGenerator.Generate()
		if err != nil {
			return false, fmt.Errorf("error generating request uuid: %w", err)
		}
	}

	if r.attempt == 0 {
		r.originalBody, err = MakeReplayable(r.request)
		if err != nil {
			return false, fmt.Errorf("error ensuring request can be retried: %w", err)
		}
	} else if r.attempt > 0 && r.request.GetBody != nil {
		r.request.Body, err = r.request.GetBody()
		if err != nil {
			if r.originalBody != nil {
				r.originalBody.Close()
			}

			return false, fmt.Errorf("error updating request body for retry: %w", err)
		}
	}

	// close previous attempt's response body to prevent HTTP client resource leaks
	if r.response != nil {
		// net/http response body early closing does not block until the body is
		// properly cleaned up, which would lead to a 'request canceled' error.
		// Yielding the CPU should allow the scheduler to run the cleanup tasks
		// before continuing. But we found that that behavior is not deterministic,
		// we instead avoid the problem altogether by reading the entire body and
		// forcing an EOF.
		// This should not be necessary when the following CL gets accepted:
		// https://go-review.googlesource.com/c/go/+/62891
		_, _ = io.Copy(io.Discard, r.response.Body)

		r.response.Body.Close()
	}

	r.attempt++

	log.Debugf("[requestID=%s] Requesting (attempt=%d): %s", r.requestID, r.attempt, formatRequest(r.request))
	r.response, err = r.delegate.Do(r.request)

	attemptable, err := r.isResponseAttemptable(r.response, err)
	if !attemptable && r.originalBody != nil {
		r.originalBody.Close()
	}

	return attemptable, err
}

func (r *RequestRetryable) Response() *http.Response {
	return r.response
}

func defaultIsAttemptable(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	return !wasSuccessful(resp), nil
}

func wasSuccessful(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < http.StatusMultipleChoices
}

func formatRequest(req *http.Request) string {
	if req == nil {
		return "Request(nil)"
	}

	return fmt.Sprintf("Request{ Method: '%s', URL: '%s' }", req.Method, req.URL)
}

func MakeReplayable(r *http.Request) (io.ReadCloser, error) {
	var err error

	if r.Body == nil {
		return nil, nil
	} else if r.GetBody != nil {
		return nil, nil
	}

	var originalBody = r.Body

	if seekableBody, ok := r.Body.(io.ReadSeeker); ok {
		r.GetBody = func() (io.ReadCloser, error) {
			_, err := seekableBody.Seek(0, 0)
			if err != nil {
				return nil, fmt.Errorf("Seeking to beginning of seekable request body: %w", err)
			}

			return io.NopCloser(seekableBody), nil
		}
	} else {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return originalBody, fmt.Errorf("Buffering request body: %w", err)
		}

		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	r.Body, err = r.GetBody()
	if err != nil {
		return originalBody, fmt.Errorf("Buffering request body: %w", err)
	}

	return originalBody, nil
}
