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

package cf

import (
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	type args struct {
		cfg          *Config
		retryTimeout time.Duration
		retryPause   time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *ClientImpl
		wantErr bool
	}{
		{
			name: "create a cf client with config",
			args: args{
				cfg: &Config{
					Username:     "some-user",
					Password:     "some-password",
					URL:          "https://api.example.com",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-password",
					SSLDisabled:  true,
				},
				retryPause:   3 * time.Second,
				retryTimeout: 1 * time.Second,
			},
			want: &ClientImpl{
				cfConfig:      nil,
				CachingClient: nil,
				Config: &Config{
					SSLDisabled:  true,
					URL:          "https://api.example.com",
					Username:     "some-user",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-password",
				},
				RetryPause:   3 * time.Second,
				RetryTimeout: 1 * time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.cfg,
				WithRetryPause(tt.args.retryPause),
				WithRetryTimeout(tt.args.retryTimeout),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_lazyLoadClientConfig(t *testing.T) {
	type fields struct {
		Config *Config
	}
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *cfclient.Config
	}{
		{
			name: "lazy load cf client config",
			fields: fields{
				Config: &Config{
					SSLDisabled:  true,
					URL:          "https://api.example.com",
					Username:     "some-user",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-password",
				},
			},
			args: args{
				cfg: &Config{
					SSLDisabled:  true,
					URL:          "https://api.example.com",
					Username:     "some-user",
					Password:     "some-password",
					ClientID:     "some-client-id",
					ClientSecret: "some-client-password",
				},
			},
			want: &cfclient.Config{
				ApiAddress:        "https://api.example.com",
				Username:          "some-user",
				Password:          "some-password",
				ClientID:          "some-client-id",
				ClientSecret:      "some-client-password",
				SkipSslValidation: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.fields.Config)
			assert.NoError(t, err)
			if got := c.lazyLoadClientConfig(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lazyLoadClientConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
