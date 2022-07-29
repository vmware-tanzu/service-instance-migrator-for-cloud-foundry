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
	"fmt"
	"net/http"
	"time"
)

type retryClient struct {
	delegate              Client
	maxAttempts           uint
	retryDelay            time.Duration
	isResponseAttemptable func(*http.Response, error) (bool, error)
}

func NewRetryClient(
	delegate Client,
	maxAttempts uint,
	retryDelay time.Duration,
) Client {
	return &retryClient{
		delegate:              delegate,
		maxAttempts:           maxAttempts,
		retryDelay:            retryDelay,
		isResponseAttemptable: nil,
	}
}

func NewNetworkSafeRetryClient(
	delegate Client,
	maxAttempts uint,
	retryDelay time.Duration,
) Client {
	return &retryClient{
		delegate:    delegate,
		maxAttempts: maxAttempts,
		retryDelay:  retryDelay,
		isResponseAttemptable: func(resp *http.Response, err error) (bool, error) {
			if err != nil || ((resp.Request.Method == "GET" || resp.Request.Method == "HEAD") && (resp.StatusCode == http.StatusGatewayTimeout || resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusBadGateway)) {
				return true, fmt.Errorf("retry error: %w", err)
			}

			return false, nil
		},
	}
}

func (r *retryClient) Do(req *http.Request) (*http.Response, error) {
	requestRetryable := NewRequestRetryable(req, r.delegate, r.isResponseAttemptable)
	retryStrategy := NewAttemptRetryStrategy(int(r.maxAttempts), r.retryDelay, requestRetryable)
	err := retryStrategy.Try()

	return requestRetryable.Response(), err
}
