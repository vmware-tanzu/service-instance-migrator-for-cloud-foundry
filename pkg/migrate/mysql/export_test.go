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

package mysql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	scriptfakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	iofakes "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/s3/fakes"
)

func TestInstallADBRPlugin(t *testing.T) {
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		wantErr error
	}{
		{
			name: "calls install plugin",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			step: InstallADBRPlugin(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					require.Equal(t, "CF_HOME='.cf' cf install-plugin -r CF-Community \"ApplicationDataBackupRestore\" -f\n", dst.String())
					return exec.Result{
						Output: dst.String(),
					}, nil
				},
			}, ".cf"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateBackup(t *testing.T) {
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		wantErr error
	}{
		{
			name: "creates a backup",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			wantErr: nil,
			step: CreateBackup(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("CF_HOME='.cf' cf adbr backup %q", "some-instance"), dst.String())
					return exec.Result{
						Output: dst.String(),
					}, nil
				},
			}, ".cf", cf.ServiceInstance{
				Name: "some-instance",
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetBackupStatus(t *testing.T) {
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		wantErr error
		want    string
	}{
		{
			name: "gets backup status",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			wantErr: nil,
			step: GetBackupStatus(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("CF_HOME='.cf' cf adbr get-status %q", "some-instance"), dst.String())
					return exec.Result{
						Output: `Getting status of service instance mysqldb in org cloudfoundry / space test-app as admin...
[Wed Dec  1 22:54:08 UTC 2021] Status: Backup was successful. Uploaded 3.2M`,
					}, nil
				},
			}, ".cf", cf.ServiceInstance{
				Name: "some-instance",
			}, 1*time.Minute, 1*time.Millisecond),
			want: `Getting status of service instance mysqldb in org cloudfoundry / space test-app as admin...
[Wed Dec  1 22:54:08 UTC 2021] Status: Backup was successful. Uploaded 3.2M`,
		},
		{
			name: "fails to backup status",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			wantErr: errors.New("adbr failed to backup instance \"some-instance\""),
			step: GetBackupStatus(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("CF_HOME='.cf' cf adbr get-status %q", "some-instance"), dst.String())
					return exec.Result{
						Output: "Backup failed",
					}, nil
				},
			}, ".cf", cf.ServiceInstance{
				Name: "some-instance",
			}, 1*time.Minute, 1*time.Millisecond),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun)
			if err != nil && tt.wantErr == nil {
				require.NoError(t, err)
			} else if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, res.(exec.Result).Output)
		})
	}
}

func TestGetLatestBackup(t *testing.T) {
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		wantErr error
	}{
		{
			name: "gets latest backup",
			args: args{
				dryRun: false,
				config: &config.Migration{},
			},
			wantErr: nil,
			step: GetLatestBackup(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("CF_HOME='.cf' cf adbr list-backups %q -l 1", "some-instance"), dst.String())
					return exec.Result{
						Output: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
19d5e006-4e5d-4e3d-a7f2-53a8448da432_1638563813   Fri Dec  3 20:36:53 UTC 2021`,
					}, nil
				},
			}, ".cf", cf.ServiceInstance{
				Name:       "some-instance",
				GUID:       "19d5e006-4e5d-4e3d-a7f2-53a8448da432",
				BackupID:   "1638563813",
				BackupDate: "2021/12/03",
				BackupTime: "20:36:53",
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadBackup(t *testing.T) {
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		wantErr error
	}{
		{
			name: "downloads latest scp backup",
			args: args{
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type": "scp",
								"scp": map[string]interface{}{
									"username":              "me",
									"hostname":              "my.scp.host.com",
									"port":                  22,
									"destination_directory": "/path/to/backup",
									"private_key":           "/path/to/private-key",
								},
							},
						},
					},
				},
				dryRun: false,
			},
			step: DownloadBackup(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					var dst bytes.Buffer
					_, err := io.Copy(&dst, r)
					require.NoError(t, err)
					require.True(t, strings.Contains(dst.String(), `me@my.scp.host.com:/path/to/backup/p.mysql/service-instance_some-guid/2021/11/24/some-guid_1637790300.tar .`))
					return exec.Result{Output: dst.String()}, nil
				},
				LastResultStub: func() exec.Result {
					return exec.Result{
						Output: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637787892   Wed Nov 24 21:04:52 UTC 2021`,
					}
				},
			}, &cf.ServiceInstance{
				GUID:       "some-guid",
				BackupDate: "2021/11/24",
				BackupTime: "21:04:52",
				BackupID:   "1637790300",
				BackupFile: "/path/to/mysql-backup-1638568756-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg",
			}, &fakes.FakeObjectDownloader{}, func(s string) (string, string, error) {
				return "2021/11/24", "21:04:52", nil
			}, func(s string) (string, error) {
				return "1637790300", nil
			}, &iofakes.FakeFileSystemOperations{}, "."),
		},
		{
			name: "downloads latest minio backup",
			args: args{
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":      "minio",
								"backup_directory": "/path/to/mysql-backups",
								"minio": map[string]interface{}{
									"alias":       "minio",
									"url":         "https://object.store.com",
									"access_key":  "some-access-key",
									"secret_key":  "some-secret-key",
									"bucket_name": "some-bucket",
									"bucket_path": "some-path",
									"insecure":    false,
								},
							},
						},
					},
				},
				dryRun: false,
			},
			step: DownloadBackup(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					var dst bytes.Buffer
					_, err := io.Copy(&dst, r)
					require.NoError(t, err)
					require.True(t, strings.Contains(dst.String(), `mc alias set minio https://object.store.com some-access-key some-secret-key
mkdir -p /path/to/mysql-backups/some-guid/1637790300
mc cp -q minio/some-bucket/some-path/service-instance_some-guid/2021/11/24/some-guid_1637790300.tar /path/to/mysql-backups
tar xvf /path/to/mysql-backups/some-guid_1637790300.tar -C /path/to/mysql-backups/some-guid/1637790300
mv /path/to/mysql-backups/some-guid/1637790300/*.tar.gpg /path/to/mysql-backups/some-guid/1637790300/mysql-backup.tar.gpg
rm -f /path/to/mysql-backups/some-guid_1637790300.tar`))
					return exec.Result{Output: dst.String()}, nil
				},
				LastResultStub: func() exec.Result {
					return exec.Result{
						Output: "Added `minio` successfully.\n...guid_1637790300.tar:  3.21 MiB / 3.21 MiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  3.64 MiB/s 0sx mysql-backup-1638566423-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg\nx mysql-backup-1638566423-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.txt",
					}
				},
			}, &cf.ServiceInstance{
				GUID:       "some-guid",
				BackupDate: "2021/11/24",
				BackupTime: "21:04:52",
				BackupID:   "1637790300",
				BackupFile: "/path/to/mysql-backups/service-instance_1638568756-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg",
			}, &fakes.FakeObjectDownloader{}, func(s string) (string, string, error) {
				return "2021/11/24", "21:04:52", nil
			}, func(s string) (string, error) {
				return "1637790300", nil
			}, &iofakes.FakeFileSystemOperations{}, "."),
		},
		{
			name: "downloads latest s3 backup",
			args: args{
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":      "s3",
								"backup_directory": "/path/to/mysql-backups",
								"s3": map[string]interface{}{
									"endpoint":          "https://object.store.com",
									"access_key_id":     "some-access-key",
									"secret_access_key": "some-secret-key",
									"region":            "some-region",
									"bucket_name":       "some-bucket",
									"bucket_path":       "some-path",
									"insecure":          false,
									"force_path_style":  false,
								},
							},
						},
					},
				},
				dryRun: false,
			},
			step: DownloadBackup(&scriptfakes.FakeExecutor{
				LastResultStub: func() exec.Result {
					return exec.Result{
						Output: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
19d5e006-4e5d-4e3d-a7f2-53a8448da432_1638563813   Fri Dec  3 20:36:53 UTC 2021`,
					}
				},
			}, &cf.ServiceInstance{
				GUID:       "19d5e006-4e5d-4e3d-a7f2-53a8448da432",
				BackupDate: "2021/12/03",
				BackupTime: "20:36:53",
				BackupID:   "1638563813",
				BackupFile: "/path/to/mysql-backups/service-instance_1638563813-19d5e006-4e5d-4e3d-a7f2-53a8448da432.tar.gpg",
			}, &fakes.FakeObjectDownloader{}, func(s string) (string, string, error) {
				return "2021/11/24", "21:04:52", nil
			}, func(s string) (string, error) {
				return "1637790300", nil
			}, &iofakes.FakeFileSystemOperations{}, "."),
		},
		{
			name: "downloads latest minio backup with insecure flag",
			args: args{
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type":      "minio",
								"backup_directory": "/path/to/mysql-backups",
								"minio": map[string]interface{}{
									"alias":       "minio",
									"url":         "https://object.store.com",
									"access_key":  "some-access-key",
									"secret_key":  "some-secret-key",
									"bucket_name": "some-bucket",
									"bucket_path": "some-path",
									"insecure":    true,
								},
							},
						},
					},
				},
				dryRun: false,
			},
			step: DownloadBackup(&scriptfakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					var dst bytes.Buffer
					_, err := io.Copy(&dst, r)
					require.NoError(t, err)
					require.True(t, strings.Contains(dst.String(), `mc alias --insecure set minio https://object.store.com some-access-key some-secret-key
mkdir -p /path/to/mysql-backups/some-guid/1637790300
mc cp -q --insecure minio/some-bucket/some-path/service-instance_some-guid/2021/11/24/some-guid_1637790300.tar /path/to/mysql-backups
tar xvf /path/to/mysql-backups/some-guid_1637790300.tar -C /path/to/mysql-backups/some-guid/1637790300
mv /path/to/mysql-backups/some-guid/1637790300/*.tar.gpg /path/to/mysql-backups/some-guid/1637790300/mysql-backup.tar.gpg
rm -f /path/to/mysql-backups/some-guid_1637790300.tar`))
					return exec.Result{Output: dst.String()}, nil
				},
				LastResultStub: func() exec.Result {
					return exec.Result{
						Output: "Added `minio` successfully.\n...guid_1637790300.tar:  3.21 MiB / 3.21 MiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  3.64 MiB/s 0sx mysql-backup-1638566423-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg\nx mysql-backup-1638566423-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.txt",
					}
				},
			}, &cf.ServiceInstance{
				GUID:       "some-guid",
				BackupDate: "2021/11/24",
				BackupTime: "21:04:52",
				BackupID:   "1637790300",
			}, &fakes.FakeObjectDownloader{}, func(s string) (string, string, error) {
				return "2021/11/24", "21:04:52", nil
			}, func(s string) (string, error) {
				return "1637790300", nil
			}, &iofakes.FakeFileSystemOperations{}, "."),
		},
		{
			name: "missing required parameters in config",
			args: args{
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"backup_type": "scp",
								"scp": map[string]interface{}{
									"username":              "",
									"hostname":              "",
									"destination_directory": "",
									"port":                  0,
									"private_key":           "",
								},
							},
						},
					},
				},
				dryRun: false,
			},
			step: DownloadBackup(&scriptfakes.FakeExecutor{}, &cf.ServiceInstance{
				GUID:       "some-guid",
				BackupDate: "2021/11/24",
				BackupTime: "21:04:52",
				BackupID:   "1637790300",
				BackupFile: "/path/to/mysql-backup-1638568756-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg",
			}, &fakes.FakeObjectDownloader{}, func(s string) (string, string, error) {
				return "2021/11/24", "21:04:52", nil
			}, func(s string) (string, error) {
				return "1637790300", nil
			}, &iofakes.FakeFileSystemOperations{}, "."),
			wantErr: fmt.Errorf(`required param "hostname" is not set in si-migrator.yml for scp backup: required param "username" is not set in si-migrator.yml for scp backup: required param "destination directory" is not set in si-migrator.yml for scp backup`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun)
			if err != nil && tt.wantErr == nil {
				require.NoError(t, err)
			} else if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
			}
		})
	}
}

func TestRetrieveEncryptionKey(t *testing.T) {
	fakeScriptExecutor := new(scriptfakes.FakeExecutor)
	type args struct {
		config   *config.Migration
		instance *cf.ServiceInstance
		dryRun   bool
		want1    string
		want2    string
		want3    string
		want4    string
		om       config.OpsManager
	}
	tests := []struct {
		name               string
		args               args
		wantErr            error
		fakeScriptExecutor *scriptfakes.FakeExecutor
	}{
		{
			name: "retrieves encryption key",
			args: args{
				dryRun: false,
				instance: &cf.ServiceInstance{
					GUID:       "some-guid",
					BackupID:   "some-backup-id",
					BackupFile: "/path/to/mysql-backup-1638566224-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg",
				},
				om: config.OpsManager{
					URL:          "opsman.tas2.example.com",
					Username:     "admin",
					Password:     "admin-password",
					ClientID:     "",
					ClientSecret: "",
					PrivateKey:   "opsman-private-key",
					IP:           "10.1.1.1",
					SshUser:      "ubuntu",
				},
				config: &config.Migration{},
				want1:  "ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh deps --column=name | grep '^cf-' | tr -d '\\t\\n'",
				want2:  "OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t 'opsman.tas2.example.com' -k curl -s -p /api/v0/deployed/products/cf-7de431470b92530a463b/credentials/.uaa.credhub_admin_client_client_credentials | jq -r .credential.value.password",
				want3:  "ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh deps --column=name | grep pivotal-mysql",
				want4:  "ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh -d pivotal-mysql-205aa795cf42568cc984 ssh dedicated-mysql-broker/0 -c \"/var/vcap/packages/credhub-cli/bin/credhub api https://credhub.service.cf.internal:8844 --ca-cert /var/vcap/jobs/adbr-api/config/credhub_ca.pem && \\\n\n/var/vcap/packages/credhub-cli/bin/credhub login --client-name credhub_admin_client --client-secret credhub-secret && \\\n/var/vcap/packages/credhub-cli/bin/credhub get -n /tanzu-mysql/backups/some-guid_some-backup-id -q\"",
			},
			fakeScriptExecutor: fakeScriptExecutor,
		},
	}
	for _, tt := range tests {
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(0, exec.Result{
			Status: &exec.Status{
				Output: `cf-7de431470b92530a463b`,
			},
		}, nil)
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(1, exec.Result{
			Status: &exec.Status{
				Output: `credhub-secret`,
			},
		}, nil)
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(2, exec.Result{
			Status: &exec.Status{
				Output: `pivotal-mysql-205aa795cf42568cc984`,
			},
		}, nil)
		tt.fakeScriptExecutor.ExecuteReturnsOnCall(3, exec.Result{
			Status: &exec.Status{
				Output: `some-enc-key`,
			},
		}, nil)
		t.Run(tt.name, func(t *testing.T) {
			if _, err := flow.RunWith(RetrieveEncryptionKey(tt.fakeScriptExecutor, tt.args.om, tt.args.instance, func(s string) (string, error) {
				return "some-enc-key", nil
			}), context.TODO(), tt.args.config, tt.args.dryRun); err != tt.wantErr {
				t.Errorf("Run() = %v, want %v", err, tt.wantErr)
			}
		})
		require.Equal(t, 4, tt.fakeScriptExecutor.ExecuteCallCount())
		_, got1 := tt.fakeScriptExecutor.ExecuteArgsForCall(0)
		require.Equal(t, tt.args.want1, copyFrom(t, got1).String())
		_, got2 := tt.fakeScriptExecutor.ExecuteArgsForCall(1)
		require.Equal(t, tt.args.want2, copyFrom(t, got2).String())
		_, got3 := tt.fakeScriptExecutor.ExecuteArgsForCall(2)
		require.Equal(t, tt.args.want3, copyFrom(t, got3).String())
		_, got4 := tt.fakeScriptExecutor.ExecuteArgsForCall(3)
		println(got4)
		require.Equal(t, tt.args.want4, copyFrom(t, got4).String())
	}
}

func Test_backupDateTimeExtractor(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name     string
		args     args
		wantDate string
		wantTime string
		wantErr  string
	}{
		{
			name: "at least 3 lines of output returns date and time",
			args: args{
				s: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637787892   Wed Nov 24 21:04:52 UTC 2021
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637596800   Mon Nov 22 16:00:00 UTC 2021
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637568000   Mon Nov 22 08:00:00 UTC 2021
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637539200   Mon Nov 22 00:00:00 UTC 2021`,
			},
			wantDate: "2021/11/24",
			wantTime: "21:04:52",
		},
		{
			name: "day contains single digit",
			args: args{
				s: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
19d5e006-4e5d-4e3d-a7f2-53a8448da432_1638394201   Wed Dec  1 21:30:01 UTC 2021`,
			},
			wantDate: "2021/12/01",
			wantTime: "21:30:01",
		},
		{
			name: "less than 3 lines returns empty strings and error",
			args: args{
				s: "line1\nline1",
			},
			wantDate: "",
			wantTime: "",
			wantErr:  "couldn't extract datetime, not enough lines in output",
		},
		{
			name: "empty string returns empty strings and error",
			args: args{
				s: "",
			},
			wantDate: "",
			wantTime: "",
			wantErr:  "couldn't extract datetime, output is empty",
		},
		{
			name: "empty string returns empty strings and error",
			args: args{
				s: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637787892   Wed`,
			},
			wantDate: "",
			wantTime: "",
			wantErr:  "couldn't extract datetime, not enough fields to parse in output",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, gotTime, gotErr := backupDateTimeExtractor(tt.args.s)
			if tt.wantErr != "" {
				require.Error(t, gotErr, tt.wantErr)
			}
			if gotDate != tt.wantDate {
				t.Errorf("backupDateTimeExtractor() gotDate = %s, wantDate %s", gotDate, tt.wantDate)
			}
			if gotTime != tt.wantTime {
				t.Errorf("backupDateTimeExtractor() gotTime = %s, wantTime %s", gotTime, tt.wantTime)
			}
		})
	}
}

func Test_backupIDExtractor(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		error   error
	}{
		{
			name: "at least 3 lines with backups returns backup id",
			args: args{
				s: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637787892   Wed Nov 24 21:04:52 UTC 2021
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637596800   Mon Nov 22 16:00:00 UTC 2021
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637568000   Mon Nov 22 08:00:00 UTC 2021
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637539200   Mon Nov 22 00:00:00 UTC 2021`,
			},
			want: "1637787892",
		},
		{
			name: "less than 3 lines returns empty strings and error",
			args: args{
				s: "line1\nline1",
			},
			want:    "",
			wantErr: true,
			error:   fmt.Errorf("couldn't extract backup id, not enough lines in output"),
		},
		{
			name: "empty string returns empty strings and error",
			args: args{
				s: "",
			},
			want:    "",
			wantErr: true,
			error:   fmt.Errorf("couldn't extract backup id, output is empty"),
		},
		{
			name: "empty string returns empty strings and error",
			args: args{
				s: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
006b68c9-e7a4-467c-b90d-72d44e3f3039_1637787892   Wed`,
			},
			want:    "",
			wantErr: true,
			error:   fmt.Errorf("couldn't extract backup id, not enough fields to parse in output"),
		},
		{
			name: "no underscore in id returns error",
			args: args{
				s: `Getting backups of service instance mysqldb in org cloudfoundry / space test-app as admin...
Backup ID                                         Time of Backup
006b68c9-e7a4-467c-b90d-72d44e3f30391637787892   Wed Nov 24 21:04:52 UTC 2021`,
			},
			want:    "",
			wantErr: true,
			error:   fmt.Errorf("couldn't extract backup id, no underscore in fields"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := backupIDExtractor(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("backupIDExtractor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				require.EqualError(t, err, tt.error.Error())
			}
			if got != tt.want {
				t.Errorf("backupIDExtractor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encryptionKeyExtractor(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "extract encryption key from output",
			args: args{
				s: `dedicated-mysql-broker/9f971072-d71a-4535-bba6-fac6b0e68daa: stderr | Unauthorized use is strictly prohibited. All access and activity
dedicated-mysql-broker/9f971072-d71a-4535-bba6-fac6b0e68daa: stderr | is subject to logging and monitoring.
dedicated-mysql-broker/9f971072-d71a-4535-bba6-fac6b0e68daa: stdout | Setting the target url: https://credhub.service.cf.internal:8844
dedicated-mysql-broker/9f971072-d71a-4535-bba6-fac6b0e68daa: stdout | Login Successful
dedicated-mysql-broker/9f971072-d71a-4535-bba6-fac6b0e68daa: stdout | 0bvUHqAFqoCK
dedicated-mysql-broker/9f971072-d71a-4535-bba6-fac6b0e68daa: stderr | Connection to 192.168.2.31 closed.
`,
			},
			want:    "0bvUHqAFqoCK",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encryptionKeyExtractor(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("encryptionKeyExtractor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("encryptionKeyExtractor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkRequiredParams(t *testing.T) {
	type args struct {
		msg    string
		params map[string]string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		wantWrappedErr bool
		errorMessage   string
	}{
		{
			name: "empty params should not return error",
			args: args{
				msg:    "",
				params: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "empty value should return error",
			args: args{
				msg:    "param1 is required",
				params: map[string]string{"param1": ""},
			},
			wantErr: true,
		},
		{
			name: "more than one missing required param should return wrapped error",
			args: args{
				msg:    "required param %q is not set",
				params: map[string]string{"param1": "", "param2": ""},
			},
			errorMessage:   "required param 'param1 is not set; required param 'param2' is not set",
			wantErr:        true,
			wantWrappedErr: true,
		},
		{
			name: "wrapped error messages are recursively added to message",
			args: args{
				msg:    "required param %q is not set",
				params: map[string]string{"param1": "", "param2": "", "param3": ""},
			},
			errorMessage:   "required param 'param1 is not set; required param 'param2' is not set; required param 'param3' is not set",
			wantErr:        true,
			wantWrappedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkRequiredParams(tt.args.msg, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkRequiredParams() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if tt.wantWrappedErr {
					var wrappedErr interface{ Unwrap() error }
					if errors.As(err, &wrappedErr) {
						require.Error(t, errors.Unwrap(err), tt.errorMessage)
					} else {
						t.Errorf("checkRequiredParams() error = %v, wantWrappedErr %v", err, tt.wantWrappedErr)
					}
				} else {
					require.Error(t, err, tt.errorMessage)
				}
			}
		})
	}
}

func copyFrom(t *testing.T, r io.Reader) *bytes.Buffer {
	dst := &bytes.Buffer{}
	_, err := io.Copy(dst, r)
	require.NoError(t, err)
	return dst
}
