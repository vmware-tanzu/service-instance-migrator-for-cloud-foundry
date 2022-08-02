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
	"encoding/base64"
	"fmt"
	"net/http"
)

type RedirectFunc func(*http.Request, []*http.Request) error

type AuthRequestAdjustment struct {
	authFunc func(bool) (string, error)
	username string
	password string
}

func NewAuthRequestAdjustment(
	authFunc func(bool) (string, error),
	client,
	clientSecret string,
) AuthRequestAdjustment {
	return AuthRequestAdjustment{
		authFunc: authFunc,
		username: client,
		password: clientSecret,
	}
}

func (a AuthRequestAdjustment) NeedsReadjustment(resp *http.Response) bool {
	return resp.StatusCode == 401
}

func (a AuthRequestAdjustment) Adjust(req *http.Request, retried bool) error {
	if len(a.username) > 0 {
		data := []byte(fmt.Sprintf("%s:%s", a.username, a.password))
		encodedBasicAuth := base64.StdEncoding.EncodeToString(data)

		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encodedBasicAuth))
	} else if a.authFunc != nil {
		authHeader, err := a.authFunc(retried)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", authHeader)
	}

	return nil
}
