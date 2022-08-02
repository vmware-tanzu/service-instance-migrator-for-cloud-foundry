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

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

func TestConfig(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name    string
		want    Config
		wantErr bool
	}{
		{
			name: "creates a config used for connecting to mysql",
			want: Config{
				Type: "scp",
				Minio: struct {
					Alias      string `yaml:"alias" default:"minio"`
					URL        string `yaml:"url"`
					AccessKey  string `yaml:"access_key"`
					SecretKey  string `yaml:"secret_key"`
					Insecure   bool   `yaml:"insecure,omitempty"`
					BucketName string `yaml:"bucket_name"`
					BucketPath string `yaml:"bucket_path" default:"p.mysql"`
					Api        string `yaml:"api" default:"S3v4"`
					Path       string `yaml:"path" default:"auto"`
				}{
					Alias:      "tas1ecstestdrive",
					URL:        "https://object.ecstestdrive.com",
					AccessKey:  "access-key@ecstestdrive.emc.com",
					SecretKey:  "some-secret-key",
					Insecure:   false,
					BucketName: "mysql-tas1",
					BucketPath: "p.mysql",
				},
				S3: struct {
					Endpoint        string `yaml:"endpoint"`
					AccessKeyID     string `yaml:"access_key_id"`
					SecretAccessKey string `yaml:"secret_access_key"`
					Region          string `yaml:"region" default:"us-east-1"`
					BucketName      string `yaml:"bucket_name"`
					BucketPath      string `yaml:"bucket_path" default:"p.mysql"`
					Insecure        bool   `yaml:"insecure,omitempty"`
					ForcePathStyle  bool   `yaml:"force_path_style,omitempty"`
				}{
					Endpoint:        "https://s3.us-east-1.amazonaws.com",
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "us-west-1",
					BucketName:      "mysql-tas1",
					BucketPath:      "p.mysql",
					Insecure:        false,
					ForcePathStyle:  true,
				},
				SCP: struct {
					Username             string `yaml:"username"`
					Hostname             string `yaml:"hostname"`
					DestinationDirectory string `yaml:"destination_directory"`
					Port                 int    `yaml:"port"`
					PrivateKey           string `yaml:"private_key"`
				}{
					Username:             "me",
					Hostname:             "my.scp.host.com",
					DestinationDirectory: "/path/to/backup",
					Port:                 22,
					PrivateKey:           "/path/to/private-key",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := config.New(filepath.Join(pwd, "testdata"), filepath.Join(filepath.Join(pwd, "testdata"), "config.yml"))
			r, err := config.NewMigrationReader(c)
			require.NoError(t, err)
			cfg, err := r.GetMigration()
			require.NoError(t, err)
			var conf Config
			got := config.NewMapDecoder(conf).Decode(*cfg, "mysql")
			if (err != nil) != tt.wantErr {
				t.Errorf("Config() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config() got = %v, want %v", got, tt.want)
			}
		})
	}
}
