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

package test

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"

	"github.com/cloudfoundry-community/go-cfclient"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	packagePath = "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/cmd/si-migrator"
	OrgName     = "si-migrator-test-org"
	SpaceName   = "si-migrator-test-space"
)

var ServiceInstanceMigratorPath string

func Setup(t *testing.T) {
	InitLogger(log.InfoLevel.String())
	ServiceInstanceMigratorPath = buildSIMigrator(t)
}

func InitLogger(level string) {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	if level == "" {
		var ok bool
		level, ok = os.LookupEnv("LOG_LEVEL")
		if !ok {
			level = log.InfoLevel.String()
		}
	}

	l, err := log.ParseLevel(level)
	if err != nil {
		l = log.InfoLevel
	}
	log.SetLevel(l)
}

func SetupExportCommand(t *testing.T) cf.Client {
	client := NewCFClient(t, true)

	DeleteServices(t, client, OrgName, SpaceName)
	DeleteOrg(t, client, OrgName)
	org := CreateOrg(t, client, OrgName)
	space := CreateSpace(t, client, SpaceName, org.Guid)
	CreateCredhubService(t, client, space.Guid)
	CreateSQLServerService(t, client, space.Guid)
	CreateMySQLService(t, client, space.Guid)

	return client
}

func SetupImportCommand(t *testing.T) cf.Client {
	client := NewCFClient(t, false)

	DeleteServices(t, client, OrgName, SpaceName)
	DeleteOrg(t, client, OrgName)
	org := CreateOrg(t, client, OrgName)
	CreateSpace(t, client, SpaceName, org.Guid)

	return client
}

func NewCFClient(t *testing.T, toSource bool) cf.Client {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working dir, %s", err)
	}

	_ = os.Unsetenv("SI_MIGRATOR_CONFIG_FILE")
	_ = os.Unsetenv("SI_MIGRATOR_CONFIG_HOME")
	err = os.Setenv("SI_MIGRATOR_CONFIG_HOME", cwd)
	if err != nil {
		log.Fatalf("could not set SI_MIGRATOR_CONFIG_HOME to %s: %v", cwd, err)
	}

	cfg := config.New("", path.Join(cwd, "si-migrator.yml"))
	if err != nil {
		log.Fatalf("could not load config, %s", err)
	}

	mr, err := config.NewMigrationReader(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	if toSource {
		migrate.NewConfigLoader(cfg, mr, om.NewPropertiesProvider(cfg, cfg.Foundations.Source, om.NewClient(), bosh.NewClient(), credhub.NewClient())).SourceApiConfig()
		client, err := cf.NewClient(&cf.Config{
			URL:         cfg.SourceApi.URL,
			Username:    cfg.SourceApi.Username,
			Password:    cfg.SourceApi.Password,
			SSLDisabled: true,
		})
		require.NoError(t, err, "error creating cf client")

		return client
	}

	migrate.NewConfigLoader(cfg, mr, om.NewPropertiesProvider(cfg, cfg.Foundations.Target, om.NewClient(), bosh.NewClient(), credhub.NewClient())).TargetApiConfig()
	client, err := cf.NewClient(&cf.Config{
		URL:         cfg.TargetApi.URL,
		Username:    cfg.TargetApi.Username,
		Password:    cfg.TargetApi.Password,
		SSLDisabled: true,
	})
	require.NoError(t, err, "error creating cf client")

	return client
}

func DeleteOrg(t *testing.T, client cf.Client, orgName string) {
	org, err := client.GetOrgByName(orgName)
	if err != nil {
		if cfclient.IsOrganizationNotFoundError(err) {
			log.Infof("Not deleting org, org %s not found", orgName)
			return
		}
		require.NoError(t, err, "error getting org %s by name", orgName)
	}
	err = client.DeleteOrg(org.Guid, true, false)
	log.Infof("Deleted org %s", orgName)
	require.NoError(t, err, "error deleting org %s", orgName)
}

func CreateOrg(t *testing.T, client cf.Client, orgName string) cfclient.Org {
	log.Infof("Creating org %s", orgName)
	org, err := client.CreateOrg(cfclient.OrgRequest{Name: orgName})
	if err != nil {
		if !cfclient.IsOrganizationNameTakenError(err) {
			require.NoErrorf(t, err, "error creating org %s", orgName)
		}
		org, err = client.GetOrgByName(orgName)
		require.NoError(t, err, "error getting org %s by name", orgName)
		log.Infof("Not creating org, org %s already exists", orgName)
		return org
	}
	log.Infof("Created org %s, orgGuid: %s", orgName, org.Guid)
	return org
}

func CreateSpace(t *testing.T, client cf.Client, spaceName string, orgGuid string) cfclient.Space {
	log.Infof("Creating space %s, orgGuid: %s", spaceName, orgGuid)
	space, err := client.CreateSpace(cfclient.SpaceRequest{
		Name:             spaceName,
		OrganizationGuid: orgGuid,
	})
	if err != nil {
		if !cfclient.IsSpaceNameTakenError(err) {
			require.NoErrorf(t, err, "error creating space %s", spaceName)
		}
		space, err = client.GetSpaceByName(spaceName, orgGuid)
		require.NoError(t, err, "error getting space %s by name", spaceName)
		log.Infof("Not creating space, space %s already exists", spaceName)
		return space
	}
	log.Infof("Created space %s", spaceName)
	return space
}

func DeleteServices(t *testing.T, client cf.Client, orgName string, spaceName string) {
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
		err = WaitForReady(10 * time.Minute, client, si.Guid, "delete")
		require.NoError(t, err, "error deleting service instance %s", si.Name)
		log.Infof("Service instance %s is deleted", si.Name)
	}
}


func RunMigratorCommand(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command(ServiceInstanceMigratorPath, args...)
	stdOutErr, err := cmd.CombinedOutput()
	if exitErr, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, "0", exitErr.Error())
	}
	t.Log(string(stdOutErr))
}

func CreateCredhubService(t *testing.T, client cf.Client, spaceGuid string) {
	space, err := client.GetSpaceByGuid(spaceGuid)
	require.NoError(t, err, "error getting space by guid %s", spaceGuid)
	serviceCredentials := map[string]interface{}{"username": "admin", "password": "password1234"}
	si := CreateManagedServiceInstance(t, client, space, "database1", "credhub-broker", "default", serviceCredentials)
	err = WaitForReady(10 * time.Minute, client, si.Guid, "create")
	require.NoError(t, err, "error creating %s service instance %s", "credhub", si.Name)
	CreateApp(t, client, space, "secure-credentials-demo", si)
}

func CreateMySQLService(t *testing.T, client cf.Client, spaceGuid string)  {
	space, err := client.GetSpaceByGuid(spaceGuid)
	require.NoError(t, err, "error getting space by guid %s", spaceGuid)
	si := CreateManagedServiceInstance(t, client, space, "mysqldb", "dedicated-mysql-broker", "db-small", nil)
	err = WaitForReady(10 * time.Minute, client, si.Guid, "create")
	require.NoError(t, err, "error creating %s service instance %s", "p.mysql", si.Name)
	CreateApp(t, client, space, "spring-music", si)
}

func CreateSQLServerService(t *testing.T, client cf.Client, spaceGuid string) {
	space, err := client.GetSpaceByGuid(spaceGuid)
	require.NoError(t, err, "error getting space by guid %s", spaceGuid)
	si := CreateManagedServiceInstance(t, client, space, "sql-test", "SQLServer", "sharedVM", nil)
	err = WaitForReady(10 * time.Minute, client, si.Guid, "create")
	require.NoError(t, err, "error creating %s service instance %s", "SQLServer", si.Name)
	CreateApp(t, client, space, "client-example", si)
}

func CreateManagedServiceInstance(t *testing.T, client cf.Client, space cfclient.Space, name, brokerName, planName string, params map[string]interface{}) cfclient.ServiceInstance {
	var broker cfclient.ServiceBroker
	brokers, err := client.ListServiceBrokers()
	require.NoError(t, err, "error listing service brokers")
	for _, b := range brokers {
		if b.Name == brokerName {
			broker = b
			break
		}
	}
	if broker.Name != brokerName {
		t.Fatalf("failed to find broker by name: %s", brokerName)
	}
	var plans []cfclient.ServicePlan
	plans, err = client.ListServicePlansByQuery(url.Values{
		"q": []string{"service_broker_guid:" + broker.Guid},
	})
	require.NoError(t, err, "error listing plans by broker guid: %s", broker.Guid)

	servicePlan := plans[0] // default to the first plan
	for i, p := range plans {
		if p.Name == planName {
			servicePlan = plans[i]
			break
		}
	}

	var s cfclient.ServiceInstance
	s, err = client.CreateServiceInstance(cfclient.ServiceInstanceRequest{
		Name:            name,
		SpaceGuid:       space.Guid,
		ServicePlanGuid: servicePlan.Guid,
		Parameters:      params,
		Tags:            nil,
	})
	require.NoError(t, err, "error creating service instance: %s", name)
	log.Infof("creating service instance: %s", name)

	return s
}

func CreateApp(t *testing.T, client cf.Client, space cfclient.Space, appName string, si cfclient.ServiceInstance) *cfclient.ServiceBinding {
	app, err := client.CreateApp(cfclient.AppCreateRequest{
		Name:      appName,
		SpaceGuid: space.Guid,
		State:     cfclient.APP_STOPPED,
	})
	require.NoErrorf(t, err, "error creating app %s", appName)

	binding, err := client.CreateServiceBinding(app.Guid, si.Guid)
	require.NoErrorf(t, err, "error creating app %s", appName)

	return binding
}

func WaitForReady(timeout time.Duration, client cf.Client, serviceInstanceGUID string, operation string) error {
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

func buildSIMigrator(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "si_artifacts")
	require.NoError(t, err, "Error generating a temp artifact dir")

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
