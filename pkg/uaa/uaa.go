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

type uaaImpl struct {
	client Client
}

func (u uaaImpl) ClientCredentialsGrant() (AccessToken, error) {
	resp, err := u.client.ClientCredentialsGrant()
	if err != nil {
		return nil, err
	}

	return NewAccessToken(
		resp.Type,
		resp.AccessToken,
	), nil
}

func (u uaaImpl) OwnerPasswordCredentialsGrant(parameters []TokenParameters) (AccessToken, error) {
	resp, err := u.client.OwnerPasswordCredentialsGrant(parameters)
	if err != nil {
		return nil, err
	}

	return NewRefreshableAccessToken(
		resp.Type,
		resp.AccessToken,
		resp.RefreshToken,
	), nil
}

func (u uaaImpl) RefreshTokenGrant(refreshValue string) (AccessToken, error) {
	resp, err := u.client.RefreshTokenGrant(refreshValue)
	if err != nil {
		return nil, err
	}

	return NewRefreshableAccessToken(
		resp.Type,
		resp.AccessToken,
		resp.RefreshToken,
	), nil
}
