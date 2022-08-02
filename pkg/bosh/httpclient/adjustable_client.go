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
	gohttp "net/http"
)

//counterfeiter:generate -o fakes . Adjustment

type Adjustment interface {
	Adjust(req *gohttp.Request, retried bool) error
	NeedsReadjustment(*gohttp.Response) bool
}

//counterfeiter:generate -o fakes . AdjustedClient

type AdjustedClient interface {
	Do(*gohttp.Request) (*gohttp.Response, error)
}

type AdjustableClient struct {
	client     AdjustedClient
	adjustment Adjustment
}

func NewAdjustableClient(client AdjustedClient, adjustment Adjustment) AdjustableClient {
	return AdjustableClient{client: client, adjustment: adjustment}
}

func (c AdjustableClient) Do(req *gohttp.Request) (*gohttp.Response, error) {
	retried := req.Body != nil

	err := c.adjustment.Adjust(req, retried)
	if err != nil {
		return nil, err
	}

	originalBody, err := MakeReplayable(req)
	if originalBody != nil {
		defer originalBody.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("error making the request retryable: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if c.adjustment.NeedsReadjustment(resp) {
		resp.Body.Close()

		if req.GetBody != nil {
			req.Body, err = req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("error updating request body for retry: %w", err)
			}
		}

		err := c.adjustment.Adjust(req, true)
		if err != nil {
			return nil, err
		}

		// Try one more time again after an adjustment
		return c.client.Do(req)
	}

	return resp, nil
}
