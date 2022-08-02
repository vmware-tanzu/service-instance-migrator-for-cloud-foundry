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

package migrate_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	configfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	mysql "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/config"
	"testing"
)

func TestMigratorHelper_GetMigratorConfig(t *testing.T) {
	m := &config.Migration{
		UseDefaultMigrator: false,
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
				},
			},
		},
	}

	type fields struct {
		migrationReader *configfakes.FakeMigrationReader
	}
	type args struct {
		data interface{}
		si   *cf.ServiceInstance
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "returns the config for a mysql migrator",
			fields: fields{
				migrationReader: &configfakes.FakeMigrationReader{
					GetMigrationStub: func() (*config.Migration, error) {
						return m, nil
					},
				},
			},
			args: args{
				data: mysql.Config{},
				si: &cf.ServiceInstance{
					Name:    "mysqldb",
					Type:    "managed_service",
					Service: migrate.MySQLService.String(),
				},
			},
			want: mysql.Config{
				Type:            "minio",
				BackupDirectory: "/tmp/mysql-backup",
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
					Alias:      "ecs-blobstore",
					URL:        "https://object.example.com:9021",
					AccessKey:  "blobstore_access_key",
					SecretKey:  "blobstore_secret_key",
					Insecure:   false,
					BucketName: "mysql-tas1",
					BucketPath: "p.mysql",
					Api:        "",
					Path:       "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := migrate.NewMigratorHelper(tt.fields.migrationReader)
			got, err := r.GetMigratorConfig(tt.args.data, tt.args.si)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMigratorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetMigratorConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
