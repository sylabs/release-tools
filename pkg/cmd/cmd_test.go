// Copyright 2021 Sylabs Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"reflect"
	"testing"
)

//nolint:dupl
func TestNewBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		opts     []BuildOption
		wantErr  error
		wantEnv  map[string]string
		wantArgs []string
	}{
		{
			name:     "Defaults",
			wantEnv:  map[string]string{"CGO_ENABLED": "0"},
			wantArgs: []string{"build", "-trimpath", "-ldflags", "-s -w", "./..."},
		},
		{
			name: "OptBuildPackages",
			opts: []BuildOption{
				OptBuildPackages("./pkg/one", "./pkg/two"),
			},
			wantEnv:  map[string]string{"CGO_ENABLED": "0"},
			wantArgs: []string{"build", "-trimpath", "-ldflags", "-s -w", "./pkg/one", "./pkg/two"},
		},
		{
			name: "OptBuildWithBuiltBy",
			opts: []BuildOption{
				OptBuildWithBuiltBy("bob"),
			},
			wantEnv:  map[string]string{"CGO_ENABLED": "0"},
			wantArgs: []string{"build", "-trimpath", "-ldflags", "-s -w -X main.builtBy=bob", "./..."},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			c, err := NewBuildCommand(tt.opts...)

			if got, want := err, tt.wantErr; !errors.Is(got, want) {
				t.Fatalf("got error %v, want %v", got, want)
			}

			if got, want := c.Env(), tt.wantEnv; !reflect.DeepEqual(got, want) {
				t.Errorf("got env %v, want %v", got, want)
			}

			if got, want := c.Args(), tt.wantArgs; !reflect.DeepEqual(got, want) {
				t.Errorf("got args %v, want %v", got, want)
			}
		})
	}
}

//nolint:dupl
func TestNewInstallCommand(t *testing.T) {
	tests := []struct {
		name     string
		opts     []BuildOption
		wantErr  error
		wantEnv  map[string]string
		wantArgs []string
	}{
		{
			name: "Defaults",
			wantEnv: map[string]string{
				"CGO_ENABLED": "0",
			},
			wantArgs: []string{
				"install",
				"-trimpath",
				"-ldflags",
				"-s -w",
				"./cmd/...",
			},
		},
		{
			name: "OptBuildPackages",
			opts: []BuildOption{
				OptBuildPackages("./pkg/one", "./pkg/two"),
			},
			wantEnv:  map[string]string{"CGO_ENABLED": "0"},
			wantArgs: []string{"install", "-trimpath", "-ldflags", "-s -w", "./pkg/one", "./pkg/two"},
		},
		{
			name: "OptBuildWithBuiltBy",
			opts: []BuildOption{
				OptBuildWithBuiltBy("bob"),
			},
			wantEnv:  map[string]string{"CGO_ENABLED": "0"},
			wantArgs: []string{"install", "-trimpath", "-ldflags", "-s -w -X main.builtBy=bob", "./cmd/..."},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			c, err := NewInstallCommand(tt.opts...)

			if got, want := err, tt.wantErr; !errors.Is(got, want) {
				t.Fatalf("got error %v, want %v", got, want)
			}

			if got, want := c.Env(), tt.wantEnv; !reflect.DeepEqual(got, want) {
				t.Errorf("got env %v, want %v", got, want)
			}

			if got, want := c.Args(), tt.wantArgs; !reflect.DeepEqual(got, want) {
				t.Errorf("got args %v, want %v", got, want)
			}
		})
	}
}

func TestNewTestCommand(t *testing.T) {
	tests := []struct {
		name     string
		opts     []TestOption
		wantErr  error
		wantEnv  map[string]string
		wantArgs []string
	}{
		{
			name:     "Defaults",
			wantEnv:  map[string]string{},
			wantArgs: []string{"test", "-race", "-cover", "./..."},
		},
		{
			name: "OptTestPackages",
			opts: []TestOption{
				OptTestPackages("./pkg/one", "./pkg/two"),
			},
			wantEnv:  map[string]string{},
			wantArgs: []string{"test", "-race", "-cover", "./pkg/one", "./pkg/two"},
		},
		{
			name: "OptBuildWithBuiltBy",
			opts: []TestOption{
				OptTestWithCoverPath("path"),
			},
			wantEnv:  map[string]string{},
			wantArgs: []string{"test", "-race", "-coverprofile", "path", "./..."},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			c, err := NewTestCommand(tt.opts...)

			if got, want := err, tt.wantErr; !errors.Is(got, want) {
				t.Fatalf("got error %v, want %v", got, want)
			}

			if got, want := c.Env(), tt.wantEnv; !reflect.DeepEqual(got, want) {
				t.Errorf("got env %v, want %v", got, want)
			}

			if got, want := c.Args(), tt.wantArgs; !reflect.DeepEqual(got, want) {
				t.Errorf("got args %v, want %v", got, want)
			}
		})
	}
}
