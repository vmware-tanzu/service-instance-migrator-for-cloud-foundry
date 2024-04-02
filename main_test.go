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

package main_test

import (
	"fmt"
	"go/build"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	boshcli "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/cli"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	packagePath               = "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/cmd/si-migrator"
	exportOrgName             = "tas1-test-org"
	importOrgName             = "tas2-test-org"
	spaceName                 = "si-migrator-test-space"
	exportFailureErrMsgFormat = "error creating %s service instance %s"
)

var serviceInstanceMigrator string

func setup(t *testing.T) {
	serviceInstanceMigrator = buildSIMigrator(t)
}

func TestVersionFlag(t *testing.T) {
	setup(t)
	executeCommand(t, serviceInstanceMigrator, "--version")
}

func Test_ExportOrgSpaceCommand(t *testing.T) {
	setup(t)
	exportTestOrgSpace(t)
}

func Test_ImportOrgSpaceCommand(t *testing.T) {
	setup(t)
	importTestOrgSpace(t)
}

func exportTestOrgSpace(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working dir, %s", err)
	}

	cfg := config.NewDefaultConfig()
	mr, err := config.NewMigrationReader(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	uaaFactory := uaa.NewFactory()
	omFactory := om.NewFactory()
	dirFactory := boshcli.NewFactory()

	migrate.NewConfigLoader(cfg, mr, om.NewPropertiesProvider(cfg, cfg.Foundations.Source, om.NewClientFactory(omFactory, uaaFactory), bosh.NewClientFactory(dirFactory, uaaFactory), credhub.NewClientFactory())).SourceApiConfig()

	client := newCFClient(t, &cf.Config{
		URL:         cfg.SourceApi.URL,
		Username:    cfg.SourceApi.Username,
		Password:    cfg.SourceApi.Password,
		SSLDisabled: true,
	})
	deleteServices(t, client, exportOrgName, spaceName)
	deleteTestOrgSpace(t, client, exportOrgName)

	space := createTestOrgSpace(t, client, exportOrgName, spaceName)
	createUserProvidedServiceInstance(t, client, space)

	executeCommand(t, serviceInstanceMigrator, "export", "space", spaceName, "--org", exportOrgName, "--export-dir", path.Join(cwd, "export"), "--debug")
}

func importTestOrgSpace(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working dir, %s", err)
	}

	cfg := config.NewDefaultConfig()
	mr, err := config.NewMigrationReader(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	uaaFactory := uaa.NewFactory()
	omFactory := om.NewFactory()
	dirFactory := boshcli.NewFactory()

	migrate.NewConfigLoader(cfg, mr, om.NewPropertiesProvider(cfg, cfg.Foundations.Target, om.NewClientFactory(omFactory, uaaFactory), bosh.NewClientFactory(dirFactory, uaaFactory), credhub.NewClientFactory())).TargetApiConfig()

	client := newCFClient(t, &cf.Config{
		URL:         cfg.TargetApi.URL,
		Username:    cfg.TargetApi.Username,
		Password:    cfg.TargetApi.Password,
		SSLDisabled: true,
	})

	deleteServices(t, client, importOrgName, spaceName)
	deleteTestOrgSpace(t, client, importOrgName)
	createTestOrgSpace(t, client, importOrgName, spaceName)
	replacePath(path.Join(cwd, "export", exportOrgName), exportOrgName, importOrgName)
	executeCommand(t, serviceInstanceMigrator, "import", "space", spaceName, "--org", importOrgName, "--import-dir", path.Join(cwd, "export"), "--debug")
	verifyUserProvidedService(t, cfg)
}

func verifyUserProvidedService(t *testing.T, cfg *config.Config) {
	client := newCFClient(t, &cf.Config{
		URL:         cfg.TargetApi.URL,
		Username:    cfg.TargetApi.Username,
		Password:    cfg.TargetApi.Password,
		SSLDisabled: true,
	})
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working dir, %s", err)
	}
	org, err := client.GetOrgByName(importOrgName)
	require.NoError(t, err)
	space, err := client.GetSpaceByName(spaceName, org.Guid)
	require.NoErrorf(t, err, "failed to find space %q in org %q", spaceName, org.Guid)
	instance := &cf.ServiceInstance{}
	fd, err := io.NewFileDescriptor(path.Join(cwd, "export", importOrgName, spaceName, "si-migrator-ups.yml"))
	require.NoError(t, err)
	err = io.NewParser().Unmarshal(instance, fd)
	require.NoError(t, err)
	require.Equal(t, "si-migrator-ups", instance.Name)
	ups, err := client.ListUserProvidedServiceInstancesByQuery(url.Values{"q": []string{
		fmt.Sprintf("organization_guid:%s", org.Guid),
		fmt.Sprintf("space_guid:%s", space.Guid),
	}})
	require.NoErrorf(t, err, "error getting ups in %s/%s", org.Name, space.Name)
	require.True(t, len(ups) > 0, "no user provided services found in %s/%s", org.Name, space.Name)
	assert.Equal(t, instance.Name, ups[0].Name)
	assert.Equal(t, instance.Credentials, ups[0].Credentials)
	assert.Equal(t, strings.Split(instance.Tags, ","), ups[0].Tags)
}

func executeCommand(t *testing.T, serviceInstanceMigrator string, args ...string) {
	cmd := exec.Command(serviceInstanceMigrator, args...)
	stdOutErr, err := cmd.CombinedOutput()
	if exitErr, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, "0", exitErr.Error())
	} else if err != nil {
		require.NoError(t, err)
	}
	fmt.Println(string(stdOutErr))
}

func createTestOrgSpace(t *testing.T, client cf.Client, orgName, spaceName string) cfclient.Space {
	// create a containing org
	org, err := client.CreateOrg(cfclient.OrgRequest{Name: orgName})
	require.NoErrorf(t, err, "error creating org %s", orgName)

	// create a containing space
	space, err := client.CreateSpace(cfclient.SpaceRequest{
		Name:             spaceName,
		OrganizationGuid: org.Guid,
	})
	require.NoErrorf(t, err, "error creating space %s", spaceName)
	return space
}

func createUserProvidedServiceInstance(t *testing.T, client cf.Client, space cfclient.Space) {
	// create a user provided service
	_, err := client.CreateUserProvidedServiceInstance(cfclient.UserProvidedServiceInstanceRequest{
		Name: "si-migrator-ups",
		Credentials: map[string]interface{}{
			"username": "admin",
			"password": "secret",
		},
		SpaceGuid: space.Guid,
		Tags:      []string{"tag1", "tag2"},
	})
	require.NoError(t, err, "error creating user provided service instance")
}

func deleteTestOrgSpace(t *testing.T, client cf.Client, orgName string) {
	org, err := client.GetOrgByName(orgName)
	if err != nil {
		if cfclient.IsOrganizationNotFoundError(err) {
			return
		}
		require.NoError(t, err, "error getting org %s by name", orgName)
	}
	err = client.DeleteOrg(org.Guid, true, false)
	require.NoError(t, err, "error deleting org %s", orgName)
}

func deleteServices(t *testing.T, client cf.Client, orgName, spaceName string) {
	org, err := client.GetOrgByName(orgName)
	if err != nil {
		if cfclient.IsOrganizationNotFoundError(err) {
			return
		}
		require.NoError(t, err, "error getting org %s by name", orgName)
	}

	space, err := client.GetSpaceByName(spaceName, org.Guid)
	if err != nil {
		if cfclient.IsSpaceNotFoundError(err) {
			return
		}
		require.NoError(t, err, "error getting space %s by name", spaceName)
	}
	require.NoErrorf(t, err, "failed to find space %q in org %q", spaceName, org.Guid)
	sis, err := client.ListSpaceServiceInstances(space.Guid)
	require.NoError(t, err, "error getting service instances for '%s/%s'", orgName, spaceName)
	log.Infof("Getting service instances from org: %s, space: %s, orgGuid: %s, spaceGuid: %s", org.Name, space.Name, org.Guid, space.Guid)
	for _, si := range sis {
		bindings, err := client.ListServiceBindingsByQuery(url.Values{"q": []string{fmt.Sprintf("service_instance_guid:%s", si.Guid)}})
		for _, binding := range bindings {
			err = client.DeleteServiceBinding(binding.Guid)
			require.NoError(t, err, "error deleting service binding %s", binding.Name)
			err = client.DeleteApp(binding.AppGuid)
			require.NoError(t, err, "error deleting app %s", binding.AppGuid)
		}
		log.Infof("Deleting service instance %s, guid: %s", si.Name, si.Guid)
		err = client.DeleteServiceInstance(si.Guid, false, true)
		require.NoError(t, err, "error deleting service instance %s", si.Name)
		log.Infoln("Waiting for service instance to delete")
		err = waitForReady(10*time.Minute, client, si.Guid, "delete")
		require.NoError(t, err, "error deleting service instance %s", si.Name)
		log.Infof("Service instance %s is deleted", si.Name)
	}
}

func newCFClient(t *testing.T, config *cf.Config) cf.Client {
	client, err := cf.NewClient(config)
	require.NoError(t, err, "error creating cf client")
	return client
}

func buildSIMigrator(t *testing.T) string {
	tmpDir := os.TempDir()

	executable := filepath.Join(tmpDir, path.Base(packagePath))
	if runtime.GOOS == "windows" {
		executable = executable + ".exe"
	}

	cmdArgs := []string{"build"}
	cmdArgs = append(cmdArgs, "-o", executable, packagePath)

	goBuild := exec.Command("go", cmdArgs...)
	goBuild.Env = replaceGoPath(os.Environ(), build.Default.GOPATH)

	output, err := goBuild.CombinedOutput()
	require.NoErrorf(t, err, "Failed to build %s:\n\nError:\n%s\n\nOutput:\n%s", packagePath, err, string(output))
	return executable
}

func replaceGoPath(environ []string, newGoPath string) []string {
	var newEnviron []string
	for _, v := range environ {
		if !strings.HasPrefix(v, "GOPATH=") {
			newEnviron = append(newEnviron, v)
		}
	}
	return append(newEnviron, "GOPATH="+newGoPath)
}

func replacePath(path string, old, new string) {
	newPath := strings.Replace(path, old, new, 1)
	os.Rename(path, newPath)
}

func waitForReady(timeout time.Duration, client cf.Client, serviceInstanceGUID string, operation string) error {
	done := make(chan bool)
	var err error
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			var si cfclient.ServiceInstance
			si, err = client.GetServiceInstanceByGuid(serviceInstanceGUID)
			if err != nil {
				log.Errorf("failed to find service instance %s", serviceInstanceGUID)
				close(done)
				return
			}

			state := si.LastOperation.State

			switch state {
			case "failed":
				log.Errorf("service instance %s failed to %s", serviceInstanceGUID, operation)
				close(done)
				return
			case "succeeded":
				log.Infof("%s service instance succeeded", operation)
				close(done)
				return
			case "in progress":
				log.Warnf("%s service instance in progress...", operation)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	select {
	case <-done:
		if err != nil {
			if operation == "delete" && cfclient.IsServiceInstanceNotFoundError(err) {
				return nil
			}
		}
		return err
	case <-time.After(timeout):
		close(done)
		return fmt.Errorf("timed out waiting for service instance to %s", operation)
	}
}
