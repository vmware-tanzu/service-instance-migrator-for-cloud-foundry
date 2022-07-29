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

package bosh

import (
	"fmt"
	"regexp"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/cli"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"

	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/pkg/errors"
)

const (
	uaaTypeString = "uaa"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Client

type Client interface {
	VerifyAuth() error
	FindDeployment(name string) (director.DeploymentResp, bool, error)
	FindVM(deploymentName, processName string) (director.VMInfo, bool, error)
}

//counterfeiter:generate -o fakes . ClientFactory

type ClientFactory interface {
	New(url string,
		allProxy string,
		trustedCertPEM []byte,
		certAppender CertAppender,
		directorFactory DirectorFactory,
		uaaFactory UAAFactory,
		boshAuth config.Authentication) (Client, error)
}

type ClientFactoryFunc func(url string, allProxy string, trustedCertPEM []byte, certAppender CertAppender, directorFactory DirectorFactory, uaaFactory UAAFactory, boshAuth config.Authentication) (Client, error)

func (f ClientFactoryFunc) New(
	url string,
	allProxy string,
	trustedCertPEM []byte,
	certAppender CertAppender,
	directorFactory DirectorFactory,
	uaaFactory UAAFactory,
	boshAuth config.Authentication,
) (Client, error) {
	return f(url, allProxy, trustedCertPEM, certAppender, directorFactory, uaaFactory, boshAuth)
}

type ClientImpl struct {
	url      string
	allProxy string

	PollingInterval time.Duration
	BoshInfo        Info

	trustedCertPEM []byte
	boshAuth       config.Authentication

	uaaFactory      UAAFactory
	directorFactory DirectorFactory
}

//counterfeiter:generate -o fakes . UAA

type UAA interface {
	uaa.UAA
}

//counterfeiter:generate -o fakes . DirectorFactory

type DirectorFactory interface {
	New(allProxy string, config director.FactoryConfig, taskReporter director.TaskReporter, fileReporter director.FileReporter) (cli.Director, error)
}

//counterfeiter:generate -o fakes . UAAFactory

type UAAFactory interface {
	New(config uaa.Config) (uaa.UAA, error)
}

//counterfeiter:generate -o fakes . CertAppender

type CertAppender interface {
	AppendCertsFromPEM(pemCerts []byte) (ok bool)
}

type Info struct {
	Version            string
	UserAuthentication UserAuthentication `json:"user_authentication"`
}

type UserAuthentication struct {
	Options AuthenticationOptions
}

type AuthenticationOptions struct {
	URL string
}

func NewClient() ClientFactoryFunc {
	return New
}

func New(
	url string,
	allProxy string,
	trustedCertPEM []byte,
	certAppender CertAppender,
	directorFactory DirectorFactory,
	uaaFactory UAAFactory,
	boshAuth config.Authentication,
) (Client, error) {

	certAppender.AppendCertsFromPEM(trustedCertPEM)

	noAuthClient := &ClientImpl{url: url, allProxy: allProxy, trustedCertPEM: trustedCertPEM, directorFactory: directorFactory}

	boshInfo, err := noAuthClient.GetInfo()
	if err != nil {
		return nil, errors.Wrap(err, "error fetching BOSH director information")
	}

	return &ClientImpl{
		url:             url,
		allProxy:        allProxy,
		trustedCertPEM:  trustedCertPEM,
		boshAuth:        boshAuth,
		uaaFactory:      uaaFactory,
		directorFactory: directorFactory,
		PollingInterval: 5,
		BoshInfo:        boshInfo,
	}, nil
}

func (c *ClientImpl) GetInfo() (Info, error) {
	var boshInfo Info
	d, err := c.Director()
	if err != nil {
		return Info{}, errors.Wrap(err, "failed to build director")
	}

	directorInfo, err := d.Info()
	if err != nil {
		return Info{}, err
	}

	boshInfo.Version = directorInfo.Version

	if directorInfo.Auth.Type != uaaTypeString {
		return boshInfo, nil
	}

	uaaURL, ok := directorInfo.Auth.Options["url"].(string)
	if ok {
		boshInfo.UserAuthentication = UserAuthentication{
			Options: AuthenticationOptions{
				URL: uaaURL,
			},
		}
	} else {
		return Info{}, errors.New("cannot retrieve UAA URL from info endpoint")
	}

	return boshInfo, nil
}

func (c *ClientImpl) FindVM(deploymentName, processName string) (director.VMInfo, bool, error) {
	d, err := c.Director()
	if err != nil {
		return director.VMInfo{}, false, fmt.Errorf("failed to build director: %w", err)
	}

	dep, err := d.FindDeployment(deploymentName)
	if err != nil {
		return director.VMInfo{}, false, fmt.Errorf("cannot find deployment %s: %w", deploymentName, err)
	}

	infos, err := dep.VMInfos()
	if err != nil {
		return director.VMInfo{}, false, fmt.Errorf("cannot get the list of vms: %w", err)
	}

	for _, info := range infos {
		for _, p := range info.Processes {
			if processName == p.Name {
				return info, true, nil
			}
		}
	}

	return director.VMInfo{}, false, nil
}

func (c *ClientImpl) FindDeployment(pattern string) (director.DeploymentResp, bool, error) {
	d, err := c.Director()
	if err != nil {
		return director.DeploymentResp{}, false, fmt.Errorf("failed to build director: %w", err)
	}

	deployments, err := d.ListDeployments()
	if err != nil {
		return director.DeploymentResp{}, false, fmt.Errorf("cannot get the list of deployments: %w", err)
	}

	for _, d := range deployments {
		if found, _ := regexp.MatchString(pattern, d.Name); found {
			return d, true, nil
		}
	}

	return director.DeploymentResp{}, false, nil
}

func (c *ClientImpl) Director() (cli.Director, error) {
	directorConfig, err := c.directorConfig()
	if err != nil {
		return nil, err
	}
	return c.directorFactory.New(c.allProxy, directorConfig, director.NewNoopTaskReporter(), director.NewNoopFileReporter())
}

func (c *ClientImpl) directorConfig() (director.FactoryConfig, error) {
	directorConfig, err := director.NewConfigFromURL(c.url)
	if err != nil {
		return director.FactoryConfig{}, errors.Wrap(err, "failed to build director config from url")
	}
	directorConfig.CACert = string(c.trustedCertPEM)

	if c.boshAuth.UAA.IsSet() {
		var uaaClient uaa.UAA
		var err error
		if c.boshAuth.UAA.URL != "" {
			uaaClient, err = buildUAA(c.allProxy, c.boshAuth.UAA.URL, c.boshAuth, directorConfig.CACert, c.uaaFactory)
		} else {
			uaaClient, err = buildUAA(c.allProxy, c.BoshInfo.UserAuthentication.Options.URL, c.boshAuth, directorConfig.CACert, c.uaaFactory)
		}
		if err != nil {
			return director.FactoryConfig{}, errors.Wrap(err, "failed to build UAA client")
		}

		directorConfig.TokenFunc = uaa.NewClientTokenSession(uaaClient).ClientCredentialsTokenFunc
	} else {
		directorConfig.Client = c.boshAuth.Basic.Username
		directorConfig.ClientSecret = c.boshAuth.Basic.Password
	}

	return directorConfig, nil
}

func (c *ClientImpl) VerifyAuth() error {
	d, err := c.Director()
	if err != nil {
		return errors.Wrap(err, " to verify credentials")
	}
	isAuthenticated, err := d.IsAuthenticated()
	if err != nil {
		return errors.Wrap(err, "failed to verify credentials")
	}
	if isAuthenticated {
		return nil
	}
	return errors.New("not authenticated")
}

func buildUAA(allProxy, uaaURL string, boshAuth config.Authentication, CACert string, factory UAAFactory) (uaa.UAA, error) {
	uaaConfig, err := uaa.NewConfigFromURL(uaaURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build UAA config from url")
	}
	uaaConfig.ClientID = boshAuth.UAA.ClientCredentials.ID
	uaaConfig.ClientSecret = boshAuth.UAA.ClientCredentials.Secret
	uaaConfig.CACert = CACert
	uaaConfig.AllProxy = allProxy
	return factory.New(uaaConfig)
}
