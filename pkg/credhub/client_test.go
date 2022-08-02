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

package credhub_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetCreds(t *testing.T) {
	type fields struct {
		allProxy     string
		caCert       []byte
		clientID     string
		clientSecret string
	}
	type args struct {
		ref string
	}
	tests := []struct {
		name    string
		handler func() http.Handler
		fields  fields
		args    args
		want    map[string][]map[string]interface{}
		wantErr bool
	}{
		{
			name: "get creds returns credhub credentials",
			handler: func() http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/oauth/token", FakeHandler(t, fakeAccessToken))
				mux.HandleFunc("/api/v1/data", FakeHandler(t, fakeData))
				return mux
			},
			fields: fields{
				allProxy:     "",
				caCert:       nil,
				clientID:     "",
				clientSecret: "",
			},
			args: args{
				ref: "/path/to/credential",
			},
			want: map[string][]map[string]interface{}{
				"data": {map[string]interface{}{
					"type":               "user",
					"id":                 "some-uuid",
					"name":               "/path/to/credential",
					"version_created_at": "2022-01-27T01:10:06Z",
					"metadata":           nil,
					"value": map[string]interface{}{
						"username":      "some-user",
						"password":      "some-password",
						"password_hash": "some-password-hash",
					},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(tt.handler())
			defer s.Close()
			serverURL, err := url.Parse(s.URL)
			require.NoError(t, err)
			c := credhub.New(
				fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Hostname()),
				serverURL.Port(),
				serverURL.Port(),
				tt.fields.allProxy,
				tt.fields.caCert,
				tt.fields.clientID,
				tt.fields.clientSecret,
			)
			got, err := c.GetCreds(tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCreds() got = %v, want %v", got, tt.want)
			}

		})
	}
}

func FakeHandler(t *testing.T, s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, s)
		assert.NoError(t, err)
	}
}

var fakeAccessToken = `{
    "access_token": "fake-access-token",
    "token_type": "bearer",
    "expires_in": 599,
    "scope": "uaa.resource credhub.write credhub.read clients.admin bosh.admin",
    "jti": "some-random-string"
}`

var fakeData = `{
   "data":[
      {
         "type":"user",
         "id":"some-uuid",
         "name":"/path/to/credential",
         "version_created_at":"2022-01-27T01:10:06Z",
         "metadata":null,
         "value":{
            "username":"some-user",
            "password":"some-password",
            "password_hash":"some-password-hash"
         }
      }
   ]
}`
