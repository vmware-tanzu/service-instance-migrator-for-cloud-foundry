//go:build integration
// +build integration

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

package e2e

import (
	"os"
	"path"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/test"

	log "github.com/sirupsen/logrus"
)

func Test_ExportOrgSpaceCommand(t *testing.T) {
	test.Setup(t)
	test.SetupExportCommand(t)

	exportOrgSpace(t)
}

func exportOrgSpace(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working dir, %s", err)
	}

	test.RunMigratorCommand(t, "export", "space", test.SpaceName, "--org", test.ExportOrgName, "--export-dir", path.Join(cwd, "export-space-tests"), "--debug", "-n")
}
