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

package om_test

import (
	"github.com/onsi/gomega/ghttp"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
	"testing"
)

func TestFactory_New(t *testing.T) {
	type args struct {
		AllProxy  string
		CACert    string
		TokenFunc func(bool) (string, error)
	}
	tests := []struct {
		name    string
		args    args
		want    om.OpsManager
		wantErr bool
	}{
		{
			name: "builds an ops manager with insecure http client",
			args: args{
				CACert: "",
			},
			want:    om.NewOpsManager(httpclient.OpsManHTTPClient{}),
			wantErr: false,
		},
		{
			name: "builds an ops manager with secure http client",
			args: args{
				AllProxy: "some-proxy-url",
				CACert:   string(validCert),
			},
			want:    om.NewOpsManager(httpclient.OpsManHTTPClient{}),
			wantErr: false,
		},
		{
			name: "errors with an invalid cert",
			args: args{
				AllProxy: "some-proxy-url",
				CACert:   string(invalidCert),
			},
			want:    om.NewOpsManager(httpclient.OpsManHTTPClient{}),
			wantErr: true,
		},
	}
	server := ghttp.NewTLSServer()
	defer server.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := om.NewConfigFromURL(server.URL())
			require.NoError(t, err)
			config.CACert = tt.args.CACert
			config.AllProxy = tt.args.AllProxy

			f := om.NewFactory()
			got, err := f.New(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Error(t, err)
			}
			require.NotNil(t, got)
		})
	}
}
