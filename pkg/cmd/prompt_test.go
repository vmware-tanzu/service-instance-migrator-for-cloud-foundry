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

package cmd_test

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"io"
	"strings"
	"testing"
)

func TestConfirmYesOrNo(t *testing.T) {
	type args struct {
		s string
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "confirms no when arg is n",
			args: args{
				s: "",
				r: strings.NewReader("n\n"),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "confirms yes when arg is y",
			args: args{
				s: "",
				r: strings.NewReader("y\n"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "confirms yes when arg is return",
			args: args{
				s: "",
				r: strings.NewReader("\n"),
			},
			want:    true,
			wantErr: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmd.ConfirmYesOrNo(tt.args.s, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfirmYesOrNo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConfirmYesOrNo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
