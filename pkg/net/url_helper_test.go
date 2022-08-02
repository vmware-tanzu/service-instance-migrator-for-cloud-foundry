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

package net_test

import (
	"strings"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/net"

	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		args       args
		wantScheme string
		wantHost   string
		wantPort   int
		wantPath   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "with scheme and path",
			args: args{
				url: "https://opsman.url.com/uaa",
			},
			wantScheme: "https",
			wantHost:   "opsman.url.com",
			wantPort:   443,
			wantPath:   "/uaa",
			wantErr:    false,
		},
		{
			name: "with path, without scheme",
			args: args{
				url: "opsman.url.com/uaa",
			},
			wantScheme: "https",
			wantHost:   "opsman.url.com",
			wantPort:   443,
			wantPath:   "/uaa",
			wantErr:    false,
		},
		{
			name: "with scheme, without path",
			args: args{
				url: "https://bosh.url.com",
			},
			wantScheme: "https",
			wantHost:   "bosh.url.com",
			wantPort:   443,
			wantPath:   "",
			wantErr:    false,
		},
		{
			name: "without scheme, without path",
			args: args{
				url: "bosh.url.com",
			},
			wantScheme: "https",
			wantHost:   "bosh.url.com",
			wantPort:   443,
			wantPath:   "",
			wantErr:    false,
		},
		{
			name: "with scheme and path and port",
			args: args{
				url: "https://opsman.url.com:8234/uaa",
			},
			wantScheme: "https",
			wantHost:   "opsman.url.com",
			wantPort:   8234,
			wantPath:   "/uaa",
			wantErr:    false,
		},
		{
			name: "with path and port, without scheme",
			args: args{
				url: "opsman.url.com:9999/uaa",
			},
			wantScheme: "https",
			wantHost:   "opsman.url.com",
			wantPort:   9999,
			wantPath:   "/uaa",
			wantErr:    false,
		},
		{
			name: "with scheme and port, without path",
			args: args{
				url: "https://bosh.url.com:8834",
			},
			wantScheme: "https",
			wantHost:   "bosh.url.com",
			wantPort:   8834,
			wantPath:   "",
			wantErr:    false,
		},
		{
			name: "with port, without scheme, without path",
			args: args{
				url: "bosh.url.com:9123",
			},
			wantScheme: "https",
			wantHost:   "bosh.url.com",
			wantPort:   9123,
			wantPath:   "",
			wantErr:    false,
		},
		{
			name: "with empty url",
			args: args{
				url: "",
			},
			wantScheme: "",
			wantHost:   "",
			wantPort:   0,
			wantPath:   "",
			wantErr:    true,
			wantErrMsg: "expected non-empty URL",
		},
		{
			name: "with port, without host",
			args: args{
				url: ":9123",
			},
			wantScheme: "",
			wantHost:   "",
			wantPort:   0,
			wantPath:   "",
			wantErr:    true,
			wantErrMsg: "missing protocol scheme",
		},
		{
			name: "with path, without host",
			args: args{
				url: "/uaa",
			},
			wantScheme: "",
			wantHost:   "",
			wantPort:   0,
			wantPath:   "",
			wantErr:    true,
			wantErrMsg: "expected to extract host from URL",
		},
		{
			name: "with invalid syntax",
			args: args{
				url: "https:/uaa",
			},
			wantScheme: "",
			wantHost:   "",
			wantPort:   0,
			wantPath:   "",
			wantErr:    true,
			wantErrMsg: "error extracting port from URL",
		},
		{
			name: "with too many colons in address",
			args: args{
				url: "uaa.com::3333",
			},
			wantScheme: "",
			wantHost:   "",
			wantPort:   0,
			wantPath:   "",
			wantErr:    true,
			wantErrMsg: "error extracting host/port from URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScheme, gotHost, gotPort, gotPath, err := net.ParseURL(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tt.wantErrMsg))
				return
			}
			if gotScheme != tt.wantScheme {
				t.Errorf("ParseURL() got = %v, want %v", gotScheme, tt.wantScheme)
			}
			if gotHost != tt.wantHost {
				t.Errorf("ParseURL() got1 = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("ParseURL() got2 = %v, want %v", gotPort, tt.wantPort)
			}
			if gotPath != tt.wantPath {
				t.Errorf("ParseURL() got3 = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}
