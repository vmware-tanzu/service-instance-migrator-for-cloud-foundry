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
	"crypto/x509"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"strings"
	"testing"
)

var validCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDDTCCAfWgAwIBAgIJAOYPl1HNpMPsMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwIBcNMTYwMTE2MDY0NTA0WhgPMjI4OTEwMzAwNjQ1MDRa
MDAxCzAJBgNVBAYTAlVTMQ0wCwYDVQQKDARCT1NIMRIwEAYDVQQDDAkxMjcuMC4w
LjEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDlptk3/IXbiBgJO7DO
dSc9MASV7FSBATumxQcXvKzUuaBJECD/S/QdevoBtIXQhtyNdSNu8GN6cD550xs2
3DYibgPD+At1IxRHfGu0Hxn2ZbU4yP9SqUchJHOa7Rix6T2cnauYhh+FhilO0Elm
kOyOtAshnv70ZWUDez8ybExgSK2kCiq3tmFotNHpxN6gNJ9IQfYz1U3thX/kyjag
MrOTTzluGGgpyS7o+4eD5rL/pWTylkgufhqUm4CJkRbXlJ8Dd/bwuBtRTumO6C4q
sYU6/OGQT/HM+sYDzrUd2pe36dQ41oeWZhKn2DyixnLLqlcH3QxnHTeg139sIQfy
rIMPAgMBAAGjEzARMA8GA1UdEQQIMAaHBH8AAAEwDQYJKoZIhvcNAQEFBQADggEB
AKj2aCf1tLQFZLq+TYa/THoVP7Pmwnt49ViQO8nMnfCM3459d52vCdIodGocVg9T
x8N4/pIG3S0VCzQt+4+UJej6SyecrYpcMCtWhZ73zxTJ7lQUmknsqZCvC5BcjYgF
McML3CeFsHuHvwb7uH5h8VO6UWyFTj7xNsH4E3XZT3I92fdS11pfrBSJDGfkiAQ/
j3N1QevrxTlEuKLQFfFSbnA3XZGpkDzg/sqYiOHnVgbn84IIZ3lGXs+qzC5kTFfM
SC0K79vs7peS+FdzPUAuG7uyy0W0s5hFTRIlcvBO5w9QrwEnBEv7WrZ6oSZ5F3Ku
/M/AnjGop4LUFIbJQR0ns7U=
-----END CERTIFICATE-----`)
var invalidCert = []byte(`a totally trustworthy cert`)

func TestConfig_CACertPool(t *testing.T) {
	tests := []struct {
		name      string
		cert      string
		wantErr   bool
		afterFunc func(*x509.CertPool)
	}{
		{
			name:    "returns a cert when cert is valid",
			cert:    string(validCert),
			wantErr: false,
			afterFunc: func(pool *x509.CertPool) {
				require.NotNil(t, pool)
			},
		},
		{
			name:    "returns a nil cert",
			cert:    "",
			wantErr: false,
			afterFunc: func(pool *x509.CertPool) {
				require.Nil(t, pool)
			},
		},
		{
			name:    "returns a error when cert is not valid",
			cert:    string(invalidCert),
			wantErr: true,
			afterFunc: func(pool *x509.CertPool) {
				require.Nil(t, pool)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := om.Config{
				CACert: tt.cert,
			}
			got, err := c.CACertPool()
			if err != nil && !tt.wantErr {
				t.Errorf("CACertPool() error = %v", err)
				return
			}
			if tt.wantErr {
				require.Error(t, err)
			}
			tt.afterFunc(got)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		Host      string
		Port      int
		Path      string
		CACert    string
		AllProxy  string
		TokenFunc func(bool) (string, error)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "does not error when required fields are present",
			fields: fields{
				Host:   "valid.host.com",
				Port:   9999,
				CACert: string(validCert),
			},
			wantErr: false,
		},
		{
			name: "errors when host is missing",
			fields: fields{
				Port: 9999,
			},
			wantErr: true,
		},
		{
			name: "errors when port is missing",
			fields: fields{
				Host: "valid.host.com",
			},
			wantErr: true,
		},
		{
			name: "errors when cert is bad",
			fields: fields{
				Host:   "valid.host.com",
				Port:   9999,
				CACert: string(invalidCert),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := om.Config{
				Host:      tt.fields.Host,
				Port:      tt.fields.Port,
				Path:      tt.fields.Path,
				CACert:    tt.fields.CACert,
				AllProxy:  tt.fields.AllProxy,
				TokenFunc: tt.fields.TokenFunc,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConfigFromURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		args       args
		wantConfig om.Config
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "with scheme and path",
			args: args{
				url: "https://opsman.url.com/uaa",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "opsman.url.com",
				Port:   443,
				Path:   "/uaa",
			},

			wantErr: false,
		},
		{
			name: "with path, without scheme",
			args: args{
				url: "opsman.url.com/uaa",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "opsman.url.com",
				Port:   443,
				Path:   "/uaa",
			},
			wantErr: false,
		},
		{
			name: "with scheme, without path",
			args: args{
				url: "https://bosh.url.com",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "bosh.url.com",
				Port:   443,
				Path:   "",
			},
			wantErr: false,
		},
		{
			name: "without scheme, without path",
			args: args{
				url: "bosh.url.com",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "bosh.url.com",
				Port:   443,
				Path:   "",
			},
			wantErr: false,
		},
		{
			name: "with scheme and path and port",
			args: args{
				url: "https://opsman.url.com:8234/uaa",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "opsman.url.com",
				Port:   8234,
				Path:   "/uaa",
			},
			wantErr: false,
		},
		{
			name: "with path and port, without scheme",
			args: args{
				url: "opsman.url.com:9999/uaa",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "opsman.url.com",
				Port:   9999,
				Path:   "/uaa",
			},
			wantErr: false,
		},
		{
			name: "with scheme and port, without path",
			args: args{
				url: "https://bosh.url.com:8834",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "bosh.url.com",
				Port:   8834,
				Path:   "",
			},
			wantErr: false,
		},
		{
			name: "with port, without scheme, without path",
			args: args{
				url: "bosh.url.com:9123",
			},
			wantConfig: om.Config{
				Scheme: "https",
				Host:   "bosh.url.com",
				Port:   9123,
				Path:   "",
			},
			wantErr: false,
		},
		{
			name: "with empty url",
			args: args{
				url: "",
			},
			wantConfig: om.Config{
				Scheme: "",
				Host:   "",
				Port:   0,
				Path:   "",
			},
			wantErr:    true,
			wantErrMsg: "expected non-empty URL",
		},
		{
			name: "with port, without host",
			args: args{
				url: ":9123",
			},
			wantConfig: om.Config{
				Scheme: "",
				Host:   "",
				Port:   0,
				Path:   "",
			},
			wantErr:    true,
			wantErrMsg: "missing protocol scheme",
		},
		{
			name: "with path, without host",
			args: args{
				url: "/uaa",
			},
			wantConfig: om.Config{
				Scheme: "",
				Host:   "",
				Port:   0,
				Path:   "",
			},
			wantErr:    true,
			wantErrMsg: "expected to extract host from URL",
		},
		{
			name: "with invalid syntax",
			args: args{
				url: "https:/uaa",
			},
			wantConfig: om.Config{
				Scheme: "",
				Host:   "",
				Port:   0,
				Path:   "",
			},
			wantErr:    true,
			wantErrMsg: "error extracting port from URL",
		},
		{
			name: "with too many colons in address",
			args: args{
				url: "uaa.com::3333",
			},
			wantConfig: om.Config{
				Scheme: "",
				Host:   "",
				Port:   0,
				Path:   "",
			},
			wantErr:    true,
			wantErrMsg: "error extracting host/port from URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := om.NewConfigFromURL(tt.args.url)
			gotHost, gotPort, gotPath := cfg.Host, cfg.Port, cfg.Path
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tt.wantErrMsg))
				return
			}
			if gotHost != tt.wantConfig.Host {
				t.Errorf("NewConfigFromURL() gotHost = %v, wantHost %v", gotHost, tt.wantConfig.Host)
			}
			if gotPort != tt.wantConfig.Port {
				t.Errorf("NewConfigFromURL() gotPort = %v, wantPort %v", gotPort, tt.wantConfig.Port)
			}
			if gotPath != tt.wantConfig.Path {
				t.Errorf("NewConfigFromURL() gotPath = %v, wantPath %v", gotPath, tt.wantConfig.Path)
			}
		})
	}
}
