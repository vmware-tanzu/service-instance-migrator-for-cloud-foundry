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

package config_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"testing"
)

func TestMapDecoder_Decode(t *testing.T) {
	type fields struct {
		Config *config.Migration
	}
	type args struct {
		migration config.Migration
		key       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "decodes a configuration into a map structure",
			fields: fields{
				Config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"source_ccdb": map[string]interface{}{
									"db_host":           "192.168.11.24",
									"db_username":       "tas1_ccdb_username",
									"db_password":       "tas1_ccdb_password",
									"db_encryption_key": "tas1_ccdb_enc_key",
									"ssh_host":          "opsman1.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om1_rsa_key",
									"ssh_tunnel":        true,
								},
								"target_ccdb": map[string]interface{}{
									"db_host":           "192.168.12.24",
									"db_username":       "tas2_ccdb_username",
									"db_password":       "tas2_ccdb_password",
									"db_encryption_key": "tas2_ccdb_enc_key",
									"ssh_host":          "opsman2.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om2_rsa_key",
									"ssh_tunnel":        true,
								},
							},
						},
						{
							Name: "ecs",
							Value: map[string]interface{}{
								"source_ccdb": map[string]interface{}{
									"db_host":           "192.168.11.24",
									"db_username":       "tas1_ccdb_username",
									"db_password":       "tas1_ccdb_password",
									"db_encryption_key": "tas1_ccdb_enc_key",
									"ssh_host":          "opsman1.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om1_rsa_key",
									"ssh_tunnel":        true,
								},
								"target_ccdb": map[string]interface{}{
									"db_host":           "192.168.12.24",
									"db_username":       "tas2_ccdb_username",
									"db_password":       "tas2_ccdb_password",
									"db_encryption_key": "tas2_ccdb_enc_key",
									"ssh_host":          "opsman2.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om2_rsa_key",
									"ssh_tunnel":        true,
								},
							},
						},
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":      "minio",
								"backup_directory": "/tmp/mysql-backup",
								"minio": map[string]interface{}{
									"alias":       "ecs-blobstore",
									"url":         "https://object.example.com:9021",
									"access_key":  "blobstore_access_key",
									"secret_key":  "blobstore_secret_key",
									"bucket_name": "mysql-tas1",
									"bucket_path": "p.mysql",
									"insecure":    false,
								},
								"scp": map[string]interface{}{
									"username":              "mysql",
									"hostname":              "mysql-backup.example.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/tmp/backup_rsa_key",
								},
							},
						},
					},
				},
			},
			args: args{
				migration: config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "sqlserver",
							Value: map[string]interface{}{
								"source_ccdb": map[string]interface{}{
									"db_host":           "192.168.11.24",
									"db_username":       "tas1_ccdb_username",
									"db_password":       "tas1_ccdb_password",
									"db_encryption_key": "tas1_ccdb_enc_key",
									"ssh_host":          "opsman1.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om1_rsa_key",
									"ssh_tunnel":        true,
								},
								"target_ccdb": map[string]interface{}{
									"db_host":           "192.168.12.24",
									"db_username":       "tas2_ccdb_username",
									"db_password":       "tas2_ccdb_password",
									"db_encryption_key": "tas2_ccdb_enc_key",
									"ssh_host":          "opsman2.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om2_rsa_key",
									"ssh_tunnel":        true,
								},
							},
						},
						{
							Name: "ecs",
							Value: map[string]interface{}{
								"source_ccdb": map[string]interface{}{
									"db_host":           "192.168.11.24",
									"db_username":       "tas1_ccdb_username",
									"db_password":       "tas1_ccdb_password",
									"db_encryption_key": "tas1_ccdb_enc_key",
									"ssh_host":          "opsman1.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om1_rsa_key",
									"ssh_tunnel":        true,
								},
								"target_ccdb": map[string]interface{}{
									"db_host":           "192.168.12.24",
									"db_username":       "tas2_ccdb_username",
									"db_password":       "tas2_ccdb_password",
									"db_encryption_key": "tas2_ccdb_enc_key",
									"ssh_host":          "opsman2.example.com",
									"ssh_username":      "ubuntu",
									"ssh_private_key":   "/tmp/om2_rsa_key",
									"ssh_tunnel":        true,
								},
							},
						},
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":      "minio",
								"backup_directory": "/tmp/mysql-backup",
								"minio": map[string]interface{}{
									"alias":       "ecs-blobstore",
									"url":         "https://object.example.com:9021",
									"access_key":  "blobstore_access_key",
									"secret_key":  "blobstore_secret_key",
									"bucket_name": "mysql-tas1",
									"bucket_path": "p.mysql",
									"insecure":    false,
								},
								"scp": map[string]interface{}{
									"username":              "mysql",
									"hostname":              "mysql-backup.example.com",
									"port":                  22,
									"destination_directory": "/var/vcap/data/mysql/backups",
									"private_key":           "/tmp/backup_rsa_key",
								},
							},
						},
					},
				},
				key: "sqlserver",
			},
			want: &config.Migration{
				Migrators: []config.Migrator{
					{
						Name: "sqlserver",
						Value: map[string]interface{}{
							"source_ccdb": map[string]interface{}{
								"db_host":           "192.168.11.24",
								"db_username":       "tas1_ccdb_username",
								"db_password":       "tas1_ccdb_password",
								"db_encryption_key": "tas1_ccdb_enc_key",
								"ssh_host":          "opsman1.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om1_rsa_key",
								"ssh_tunnel":        true,
							},
							"target_ccdb": map[string]interface{}{
								"db_host":           "192.168.12.24",
								"db_username":       "tas2_ccdb_username",
								"db_password":       "tas2_ccdb_password",
								"db_encryption_key": "tas2_ccdb_enc_key",
								"ssh_host":          "opsman2.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om2_rsa_key",
								"ssh_tunnel":        true,
							},
						},
					},
					{
						Name: "ecs",
						Value: map[string]interface{}{
							"source_ccdb": map[string]interface{}{
								"db_host":           "192.168.11.24",
								"db_username":       "tas1_ccdb_username",
								"db_password":       "tas1_ccdb_password",
								"db_encryption_key": "tas1_ccdb_enc_key",
								"ssh_host":          "opsman1.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om1_rsa_key",
								"ssh_tunnel":        true,
							},
							"target_ccdb": map[string]interface{}{
								"db_host":           "192.168.12.24",
								"db_username":       "tas2_ccdb_username",
								"db_password":       "tas2_ccdb_password",
								"db_encryption_key": "tas2_ccdb_enc_key",
								"ssh_host":          "opsman2.example.com",
								"ssh_username":      "ubuntu",
								"ssh_private_key":   "/tmp/om2_rsa_key",
								"ssh_tunnel":        true,
							},
						},
					},
					{
						Name: "mysql",
						Value: map[string]interface{}{
							"backup_type":      "minio",
							"backup_directory": "/tmp/mysql-backup",
							"minio": map[string]interface{}{
								"alias":       "ecs-blobstore",
								"url":         "https://object.example.com:9021",
								"access_key":  "blobstore_access_key",
								"secret_key":  "blobstore_secret_key",
								"bucket_name": "mysql-tas1",
								"bucket_path": "p.mysql",
								"insecure":    false,
							},
							"scp": map[string]interface{}{
								"username":              "mysql",
								"hostname":              "mysql-backup.example.com",
								"port":                  22,
								"destination_directory": "/var/vcap/data/mysql/backups",
								"private_key":           "/tmp/backup_rsa_key",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := config.NewMapDecoder(tt.fields.Config)
			got := d.Decode(tt.args.migration, tt.args.key)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(config.Config{})); diff != "" {
				t.Errorf("GetMigration() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
