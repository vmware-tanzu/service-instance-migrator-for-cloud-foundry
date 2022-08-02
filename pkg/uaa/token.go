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

type AccessTokenImpl struct {
	type_       string
	accessValue string
}

var _ AccessToken = &AccessTokenImpl{}

func (t AccessTokenImpl) Type() string  { return t.type_ }
func (t AccessTokenImpl) Value() string { return t.accessValue }
func (t AccessTokenImpl) IsValid() bool { return t.type_ != "" && t.accessValue != "" }

type RefreshableAccessTokenImpl struct {
	accessToken  AccessToken
	refreshValue string
}

var _ RefreshableAccessToken = &RefreshableAccessTokenImpl{}

func (t RefreshableAccessTokenImpl) Type() string  { return t.accessToken.Type() }
func (t RefreshableAccessTokenImpl) Value() string { return t.accessToken.Value() }
func (t RefreshableAccessTokenImpl) IsValid() bool { return t.accessToken.IsValid() }

func (t RefreshableAccessTokenImpl) RefreshValue() string {
	return t.refreshValue
}

type TokenInfo struct {
	Username  string   `json:"user_name"`
	Scopes    []string `json:"scope"`
	ExpiredAt int      `json:"exp"`
}

func NewAccessToken(accessValueType, accessValue string) AccessToken {
	return AccessTokenImpl{
		type_:       accessValueType,
		accessValue: accessValue,
	}
}

func NewRefreshableAccessToken(accessValueType, accessValue, refreshValue string) RefreshableAccessToken {
	if len(refreshValue) == 0 {
		panic("Expected non-empty refresh token value")
	}

	return &RefreshableAccessTokenImpl{
		accessToken: AccessTokenImpl{
			type_:       accessValueType,
			accessValue: accessValue,
		},
		refreshValue: refreshValue,
	}
}
