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

package exec

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	type args struct {
		debug bool
	}
	tests := []struct {
		name string
		args args
		want ShellScriptExecutor
	}{
		{
			name: "creates a shell script executor with debug turned on",
			args: args{
				debug: true,
			},
			want: ShellScriptExecutor{
				debug: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *ShellScriptExecutor
			if got = NewExecutor(WithDebug(tt.args.debug)); got.debug != tt.want.debug {
				t.Errorf("NewExecutor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellScriptExecutor_Execute(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteString("ls")
	reader := strings.NewReader(buffer.String())
	type fields struct {
		Debug bool
	}
	type args struct {
		src io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "executes commands in shell",
			fields: fields{
				Debug: false,
			},
			args: args{
				src: reader,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShellScriptExecutor{
				debug: tt.fields.Debug,
			}
			var result Result
			var err error
			if result, err = s.Execute(context.Background(), tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != "" {
				require.Equal(t, tt.want, result.Output)
			}
		})
	}
}
