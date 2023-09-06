/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This code was taken from https://github.com/rancher/local-path-provisioner
and modified to work with the configuration options used by OpenEBS
*/

package app

import (
	"testing"
)

func TestConvertToK(t *testing.T) {
	type args struct {
		limit      string
		pvcStorage int64
	}
	tests := map[string]struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		"Missing limit grace": {
			args: args{
				limit:      "",
				pvcStorage: 5000000000,
			},
			want:    "0k",
			wantErr: false,
		},
		"Present limit with grace": {
			args: args{
				limit:      "0%",
				pvcStorage: 5000,
			},
			want:    "5k",
			wantErr: false,
		},
		"Present limit grace exceeding 100%": {
			args: args{
				limit:      "200%",
				pvcStorage: 5000000,
			},
			want:    "9766k",
			wantErr: false,
		},
		"Present limit grace with decimal%": {
			args: args{
				limit:      ".5%",
				pvcStorage: 1000,
			},
			want:    "1k", // the final result of limit can't be a float
			wantErr: false,
		},
		"Present limit grace with invalid pattern": {
			args: args{
				limit:      "10",
				pvcStorage: 10000,
			},
			want:    "",
			wantErr: true,
		},
		"Present limit grace with only %": {
			args: args{
				limit:      "%",
				pvcStorage: 10000,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToK(tt.args.limit, tt.args.pvcStorage)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertToK() = %v, want %v", got, tt.want)
			}
		})
	}
}
