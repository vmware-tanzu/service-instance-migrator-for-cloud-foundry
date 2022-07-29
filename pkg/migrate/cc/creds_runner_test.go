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

package cc

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec/fakes"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
)

func Test_getCredentials(t *testing.T) {
	type args struct {
		ctx        context.Context
		e          *fakes.FakeExecutor
		om         config.OpsManager
		deployment string
	}
	tests := []struct {
		name     string
		args     args
		username string
		password string
		wantErr  bool
	}{
		{
			name: "get credhub credentials returns username and password",
			args: args{
				ctx: context.TODO(),
				e: &fakes.FakeExecutor{
					ExecuteStub: func(ctx context.Context, reader io.Reader) (exec.Result, error) {
						return exec.Result{
							Status: &exec.Status{
								Output: `password: some-password
password_hash: some-password-hash
username: some-username
`,
							},
						}, nil
					},
				},
				om:         config.OpsManager{},
				deployment: "cf",
			},
			username: "some-username",
			password: "some-password",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password, err := getCredentials(tt.args.ctx, tt.args.e, tt.args.om, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if username != tt.username {
				t.Errorf("getCredentials() got = %v, want %v", username, tt.username)
			}
			if password != tt.password {
				t.Errorf("getCredentials() got1 = %v, want %v", password, tt.password)
			}
		})
	}
}

func Test_getEncryptionKey(t *testing.T) {
	type args struct {
		ctx        context.Context
		e          *fakes.FakeExecutor
		om         config.OpsManager
		deployment string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get credhub credentials returns encryption key",
			args: args{
				ctx: context.TODO(),
				e: &fakes.FakeExecutor{
					ExecuteStub: func(ctx context.Context, reader io.Reader) (exec.Result, error) {
						return exec.Result{
							Status: &exec.Status{
								Output: "some-encryption-key",
							},
						}, nil
					},
				},
				om:         config.OpsManager{},
				deployment: "cf",
			},
			want:    "some-encryption-key",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEncryptionKey(tt.args.ctx, tt.args.e, tt.args.om, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEncryptionKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getEncryptionKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findDeploymentName(t *testing.T) {
	type args struct {
		ctx     context.Context
		e       *fakes.FakeExecutor
		om      config.OpsManager
		pattern string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get cf deployment name from bosh",
			args: args{
				ctx: context.TODO(),
				e: &fakes.FakeExecutor{
					ExecuteStub: func(ctx context.Context, reader io.Reader) (exec.Result, error) {
						return exec.Result{
							Status: &exec.Status{
								Output: "cf-abc12345678de1234567",
							},
						}, nil
					},
				},
				om:      config.OpsManager{},
				pattern: "^cf",
			},
			want:    "cf-abc12345678de1234567",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findDeploymentName(tt.args.ctx, tt.args.e, tt.args.om, tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("findDeploymentName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findDeploymentName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findInstanceName(t *testing.T) {
	type args struct {
		ctx        context.Context
		e          exec.Executor
		om         config.OpsManager
		deployment string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get mysql database instance name from bosh",
			args: args{
				ctx: context.TODO(),
				e: &fakes.FakeExecutor{
					ExecuteStub: func(ctx context.Context, reader io.Reader) (exec.Result, error) {
						return exec.Result{
							Status: &exec.Status{
								Output: "192.168.1.21",
							},
						}, nil
					},
				},
				om:         config.OpsManager{},
				deployment: "cf-abc12345678de1234567",
			},
			want:    "192.168.1.21",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findInstanceName(tt.args.ctx, tt.args.e, tt.args.om, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("findInstanceName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findInstanceName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findInstanceIPAddress(t *testing.T) {
	type args struct {
		ctx        context.Context
		e          exec.Executor
		om         config.OpsManager
		deployment string
		instance   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get ccdb ip address from bosh",
			args: args{
				ctx: context.TODO(),
				e: &fakes.FakeExecutor{
					ExecuteStub: func(ctx context.Context, reader io.Reader) (exec.Result, error) {
						return exec.Result{
							Status: &exec.Status{
								Output: "192.168.1.21",
							},
						}, nil
					},
				},
				om:         config.OpsManager{},
				deployment: "cf-abc12345678de1234567",
				instance:   "mysql_proxy/guid",
			},
			want:    "192.168.1.21",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findInstanceIPAddress(tt.args.ctx, tt.args.e, tt.args.om, tt.args.deployment, tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("findInstanceIPAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findInstanceIPAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetrieveCloudControllerDatabaseCredentials(t *testing.T) {
	type args struct {
		e        *fakes.FakeExecutor
		cfg      *DatabaseConfig
		isExport bool
		config   *config.Migration
		dryRun   bool
		om       config.OpsManager
	}
	tests := []struct {
		name       string
		args       args
		wantErr    error
		beforeFunc func(args args)
		afterFunc  func(args args)
	}{
		{
			name: "retrieves all ccdb credentials",
			args: args{
				e:      new(fakes.FakeExecutor),
				cfg:    &DatabaseConfig{},
				config: &config.Migration{},
				om: config.OpsManager{
					Hostname: "om.target.com",
				},
			},
			beforeFunc: func(args args) {
				args.e.ExecuteReturnsOnCall(0, exec.Result{
					Status: &exec.Status{
						Output: "cf-abc12345678de1234567",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(1, exec.Result{
					Status: &exec.Status{
						Output: "mysql_proxy/guid",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(2, exec.Result{
					Status: &exec.Status{
						Output: "192.168.1.21",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(3, exec.Result{
					Status: &exec.Status{
						Output: `password: some-password
password_hash: some-password-hash
username: some-username
`,
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(4, exec.Result{
					Status: &exec.Status{
						Output: "some-encryption-key",
					},
				}, nil)
			},
			afterFunc: func(args args) {
				require.Equal(t, 5, args.e.ExecuteCallCount())
			},
		},
		{
			name: "does not retrieve creds if already set",
			args: args{
				e: new(fakes.FakeExecutor),
				cfg: &DatabaseConfig{
					Username: "ccdb-username",
					Password: "ccdb-password",
				},
				isExport: false,
				om: config.OpsManager{
					Hostname: "om.target.com",
				},
				config: &config.Migration{},
			},
			beforeFunc: func(args args) {
				args.e.ExecuteReturnsOnCall(0, exec.Result{
					Status: &exec.Status{
						Output: "cf-abc12345678de1234567",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(1, exec.Result{
					Status: &exec.Status{
						Output: "mysql_proxy/guid",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(2, exec.Result{
					Status: &exec.Status{
						Output: "192.168.1.21",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(3, exec.Result{
					Status: &exec.Status{
						Output: "some-encryption-key",
					},
				}, nil)
			},
			afterFunc: func(args args) {
				require.Equal(t, 4, args.e.ExecuteCallCount())
			},
		},
		{
			name: "does not retrieve encryption key if already set",
			args: args{
				e: new(fakes.FakeExecutor),
				cfg: &DatabaseConfig{
					EncryptionKey: "some-key",
				},
				isExport: false,
				om: config.OpsManager{
					Hostname: "om.target.com",
				},
				config: &config.Migration{},
			},
			beforeFunc: func(args args) {
				args.e.ExecuteReturnsOnCall(0, exec.Result{
					Status: &exec.Status{
						Output: "cf-abc12345678de1234567",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(1, exec.Result{
					Status: &exec.Status{
						Output: "mysql_proxy/guid",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(2, exec.Result{
					Status: &exec.Status{
						Output: "192.168.1.21",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(3, exec.Result{
					Status: &exec.Status{
						Output: `password: some-password
password_hash: some-password-hash
username: some-username
`,
					},
				}, nil)
			},
			afterFunc: func(args args) {
				require.Equal(t, 4, args.e.ExecuteCallCount())
			},
		},
		{
			name: "does not retrieve ip address if already set",
			args: args{
				e: new(fakes.FakeExecutor),
				cfg: &DatabaseConfig{
					Host: "192.168.1.20",
				},
				isExport: false,
				om: config.OpsManager{
					Hostname: "om.target.com",
				},
				config: &config.Migration{},
			},
			beforeFunc: func(args args) {
				args.e.ExecuteReturnsOnCall(0, exec.Result{
					Status: &exec.Status{
						Output: "cf-abc12345678de1234567",
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(1, exec.Result{
					Status: &exec.Status{
						Output: `password: some-password
password_hash: some-password-hash
username: some-username
`,
					},
				}, nil)
				args.e.ExecuteReturnsOnCall(2, exec.Result{
					Status: &exec.Status{
						Output: "some-encryption-key",
					},
				}, nil)
			},
			afterFunc: func(args args) {
				require.Equal(t, 3, args.e.ExecuteCallCount())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc(tt.args)
			_, err := flow.RunWith(SetCloudControllerDatabaseCredentials(tt.args.e, tt.args.cfg, tt.args.om), context.TODO(), tt.args.config, tt.args.dryRun)
			if err != nil && tt.wantErr == nil {
				require.NoError(t, err)
			} else if tt.wantErr != nil {
				require.Error(t, err, tt.wantErr)
				require.EqualError(t, err, tt.wantErr.Error())
			}
			tt.afterFunc(tt.args)
		})
	}
}
