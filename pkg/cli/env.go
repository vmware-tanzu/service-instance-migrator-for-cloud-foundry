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

package cli

var (
	cliName     = "service-instance-migrator"
	cliVersion  = "unknown"
	cliGitSHA   = "unknown sha"
	cliGitDirty = ""
)

type CompiledEnv struct {
	Name     string
	Version  string
	GitSha   string
	GitDirty bool
}

var Env CompiledEnv

func init() {
	// must be created inside the init function to pickup build specific params
	Env = CompiledEnv{
		Name:     cliName,
		Version:  cliVersion,
		GitSha:   cliGitSHA,
		GitDirty: cliGitDirty != "",
	}
}
