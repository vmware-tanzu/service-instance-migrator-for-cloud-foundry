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

package io_test

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		name string
		want *io.Parser
	}{
		{
			name: "creates a valid Parser",
			want: &io.Parser{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := io.NewParser(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Unmarshal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "si")
	require.NoError(t, err)
	p := path.Join(tmpDir, "org", "spacey")
	d := io.NewFileSystemHelper()
	err = d.Mkdir(p)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(p, "si-name.yml"), []byte(`name: si-name
guid: maybe-a-guid
type: user-provided
`), 0755)
	require.NoError(t, err)

	type serviceInstance struct {
		Name string `yaml:"name,omitempty"`
		GUID string `yaml:"guid,omitempty"`
		Type string `yaml:"type,omitempty"`
	}

	type args struct {
		b  interface{}
		fd io.FileDescriptor
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    interface{}
	}{
		{
			name: "unmarshal yaml file into struct",
			args: args{
				b: &serviceInstance{},
				fd: io.FileDescriptor{
					Name:      "si-name",
					Org:       "org",
					Space:     "spacey",
					BaseDir:   tmpDir,
					Extension: "yml",
				},
			},
			wantErr: false,
			want: &serviceInstance{
				Name: "si-name",
				GUID: "maybe-a-guid",
				Type: "user-provided",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := io.NewParser()
			if err := p.Unmarshal(tt.args.b, tt.args.fd); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := tt.args.b; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Marshal(t *testing.T) {
	dir, err := os.MkdirTemp("", "si")
	require.NoError(t, err)

	type args struct {
		v  interface{}
		fd io.FileDescriptor
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "writes yaml to file",
			args: args{
				v: struct {
					Name string
					GUID string
					Type string
				}{
					Name: "si-name",
					GUID: "maybe-a-guid",
					Type: "user-provided",
				},
				fd: io.FileDescriptor{
					Name:      "si-name",
					Org:       "org",
					Space:     "spacey",
					BaseDir:   dir,
					Extension: "yml",
				},
			},
			wantErr: false,
			want: `name: si-name
guid: maybe-a-guid
type: user-provided
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &io.Parser{}
			if err := w.Marshal(tt.args.v, tt.args.fd); (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := os.ReadFile(path.Join(dir, "org", "spacey", "si-name.yml"))
			require.NoError(t, err, "could not read file")
			require.Equal(t, tt.want, string(got))
		})
	}
}
