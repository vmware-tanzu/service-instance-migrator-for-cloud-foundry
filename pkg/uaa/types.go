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

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Token

// Token is a plain token with a value.
type Token interface {
	Type() string
	Value() string
	IsValid() bool
}

//counterfeiter:generate -o fakes . AccessToken

// AccessToken is purely an access token. It does not contain a refresh token and
// cannot be refreshed for another token.
type AccessToken interface {
	Token
}

//counterfeiter:generate -o fakes . RefreshableAccessToken

// RefreshableAccessToken is an access token with a refresh token that can be used
// to get another access token.
type RefreshableAccessToken interface {
	AccessToken
	RefreshValue() string
}

//counterfeiter:generate -o fakes . UAA

type UAA interface {
	RefreshTokenGrant(string) (AccessToken, error)
	ClientCredentialsGrant() (AccessToken, error)
	OwnerPasswordCredentialsGrant([]TokenParameters) (AccessToken, error)
}
