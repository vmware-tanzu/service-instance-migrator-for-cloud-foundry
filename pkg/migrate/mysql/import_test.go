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
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	mysql "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/config"
)

func TestTransferBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	backupFile := path.Join(tmpDir, "mysql-backup-1638566224-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg")
	err = os.WriteFile(backupFile, []byte(`some-data`), 0755)
	require.NoError(t, err)
	require.FileExists(t, backupFile)
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		want    string
		wantErr error
	}{
		{
			name: "transfers a backup",
			args: args{
				dryRun: true,
				config: &config.Migration{},
			},
			step: TransferBackup(&fakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh -d service-instance_some-guid scp %s mysql/0:/tmp", backupFile), dst.String())
					return exec.Result{
						Output: dst.String(),
					}, nil
				},
			}, config.OpsManager{
				URL:          "opsman.tas2.example.com",
				Username:     "admin",
				Password:     "admin-password",
				ClientID:     "",
				ClientSecret: "",
				PrivateKey:   "opsman-private-key",
				IP:           "10.1.1.1",
				SshUser:      "ubuntu",
			}, &cf.ServiceInstance{
				GUID:       "some-guid",
				BackupFile: backupFile,
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

func TestRestoreBackup(t *testing.T) {
	fakeScriptExecutorReturnsErrorOnFailedRestore := new(fakes.FakeExecutor)
	fakeScriptExecutorReturnsErrorOnFailedRestore.ExecuteReturnsOnCall(0, exec.Result{
		Status: &exec.Status{
			ScriptName: "",
			ScriptBody: "",
			Output: `mysql/bd424c1b-95ec-475e-ab2d-e114b75e67c0: stderr | Unauthorized use is strictly prohibited. All access and activity
mysql/bd424c1b-95ec-475e-ab2d-e114b75e67c0: stderr | is subject to logging and monitoring.
mysql/bd424c1b-95ec-475e-ab2d-e114b75e67c0: stdout | 2021/12/07 00:14:08 Starting restore\r
mysql/bd424c1b-95ec-475e-ab2d-e114b75e67c0: stdout | 2021/12/07 00:14:08 Applying backup artifact. This may take some time...\r
mysql/bd424c1b-95ec-475e-ab2d-e114b75e67c0: stdout | 2021/12/07 00:14:08 Error during restore process (500 Internal Server Error): failed to restore: Restore is permitted only in a non-empty service instance. Try again with a new service instance to restore.\r
mysql/bd424c1b-95ec-475e-ab2d-e114b75e67c0: stderr | Connection to 192.168.4.31 closed.\r

Running SSH:
  1 error occurred:
  * Running command: 'ssh -tt -o ServerAliveInterval=30 -o ForwardAgent=no -o PasswordAuthentication=no -o IdentitiesOnly=yes -o IdentityFile=/Users/malston/.bosh/tmp/ssh-priv-key224485643 -o UserKnownHostsFile=/Users/malston/.bosh/tmp/ssh-known-hosts658567406 -o ProxyCommand=nc -x 127.0.0.1:55136 %!h(MISSING) %!p(MISSING) -o StrictHostKeyChecking=yes 192.168.4.31 -l bosh_a7af7c56a499403 sudo mysql-restore --encryption-key fake-enc-key --restore-file /tmp/path/to/mysql-backup-1638835668-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg', stdout: '', stderr: '': exit status 1



Exit code 1
`,
			PID:      0,
			Done:     false,
			CPUTime:  0,
			ExitCode: 1,
			Error:    nil,
		}}, errors.New(`failed to execute bosh command "-d service-instance_some-guid ssh mysql/0 -c \"sudo mysql-restore --encryption-key fake-enc-key\r --restore-file /tmp/path/to/mysql-backup-1638835668-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg\"": error: exit status 1`))
	type args struct {
		config *config.Migration
		dryRun bool
	}
	tests := []struct {
		name    string
		args    args
		step    flow.StepFunc
		want    string
		wantErr error
	}{
		{
			name: "restores a backup",
			args: args{
				dryRun: true,
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"mysql": mysql.Config{},
							},
						},
					},
				},
			},
			step: RestoreBackup(&fakes.FakeExecutor{
				ExecuteStub: func(c context.Context, r io.Reader) (exec.Result, error) {
					dst := &bytes.Buffer{}
					_, err := io.Copy(dst, r)
					require.NoError(t, err)
					return exec.Result{
						Status: &exec.Status{Output: dst.String()},
					}, nil
				},
			}, config.OpsManager{
				URL:          "opsman.tas2.example.com",
				Username:     "admin",
				Password:     "admin-password",
				ClientID:     "",
				ClientSecret: "",
				PrivateKey:   "opsman-private-key",
				IP:           "10.1.1.1",
				SshUser:      "ubuntu",
			}, &cf.ServiceInstance{
				GUID:                "some-guid",
				BackupEncryptionKey: "fake-enc-key",
				BackupFile:          "/path/to/mysql-backup-1638566224-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg",
			}),
			want: "ssh_key_path=$(mktemp)\ncat \"opsman-private-key\" >\"$ssh_key_path\"\nchmod 0600 \"${ssh_key_path}\"\nbosh_ca_path=$(mktemp)\nbosh_ca_cert=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k certificate-authorities -f json | jq -r '.[] | select(.active==true) | .cert_pem')\"\necho \"$bosh_ca_cert\" >\"$bosh_ca_path\"\nchmod 0600 \"${bosh_ca_path}\"\ncreds=\"$(OM_CLIENT_ID='' OM_CLIENT_SECRET='' OM_USERNAME='admin' OM_PASSWORD='admin-password' om -t opsman.tas2.example.com -k curl -s -p /api/v0/deployed/director/credentials/bosh_commandline_credentials)\"\nbosh_all=\"$(echo \"$creds\" | jq -r .credential | tr ' ' '\\n' | grep '=')\"\nbosh_client=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT=')\"\nbosh_env=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_ENVIRONMENT=')\"\nbosh_secret=\"$(echo \"$bosh_all\" | tr ' ' '\\n' | grep 'BOSH_CLIENT_SECRET=')\"\nbosh_ca_cert=\"BOSH_CA_CERT=$bosh_ca_path\"\nbosh_proxy=\"BOSH_ALL_PROXY=ssh+socks5://ubuntu@10.1.1.1:22?private-key=${ssh_key_path}\"\nbosh_gw_host=\"BOSH_GW_HOST=10.1.1.1\"\nbosh_gw_user=\"BOSH_GW_USER=ubuntu\"\nbosh_gw_private_key=\"BOSH_GW_PRIVATE_KEY=${ssh_key_path}\"\ntrap 'rm -f ${ssh_key_path} ${bosh_ca_path}' EXIT\n/usr/bin/env \"$bosh_client\" \"$bosh_env\" \"$bosh_secret\" \"$bosh_ca_cert\" \"$bosh_proxy\" \"$bosh_gw_host\" \"$bosh_gw_user\" \"$bosh_gw_private_key\" bosh -d service-instance_some-guid ssh mysql/0 -c \"sudo mysql-restore --encryption-key fake-enc-key --restore-file /tmp/mysql-backup-1638566224-e41880ee-c5a5-4de1-a570-e7fe117bdfa8.tar.gpg\"",
		},
		{
			name: "fails to restore a non-empty service instance",
			args: args{
				dryRun: true,
				config: &config.Migration{
					Migrators: []config.Migrator{
						{
							Name: "mysql",
							Value: map[string]interface{}{
								"mysql": mysql.Config{
									Type: "minio",
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
										Alias:      "minio",
										URL:        "https://object.store.com",
										AccessKey:  "some-access-key",
										SecretKey:  "some-secret-key",
										Insecure:   false,
										BucketName: "some-bucket",
										BucketPath: "some-path",
										Api:        "",
										Path:       "",
									},
								},
							},
						},
					},
				},
			},
			step: RestoreBackup(fakeScriptExecutorReturnsErrorOnFailedRestore, config.OpsManager{
				URL:          "opsman.tas2.example.com",
				Username:     "admin",
				Password:     "admin-password",
				ClientID:     "",
				ClientSecret: "",
				PrivateKey:   "opsman-private-key",
				IP:           "10.1.1.1",
				SshUser:      "ubuntu",
			}, &cf.ServiceInstance{
				GUID: "some-guid",
			}),
			want: "Restore is permitted only in a non-empty service instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flow.Sequence(tt.step).Run(context.TODO(), tt.args.config, tt.args.dryRun)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			}
			require.Contains(t, res.(exec.Result).Status.Output, tt.want)
		})
	}
}
