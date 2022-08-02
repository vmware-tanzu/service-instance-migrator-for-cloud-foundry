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

import (
	"fmt"
	"net"
	gohttp "net/http"
	"net/url"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	"github.com/cloudfoundry/bosh-cli/director"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Director

type Director interface {
	IsAuthenticated() (bool, error)
	Info() (director.Info, error)
	ListDeployments() ([]director.DeploymentResp, error)
	FindDeployment(name string) (Deployment, error)
}

type Factory struct{}

func NewFactory() Factory {
	return Factory{}
}

func (f Factory) New(allProxy string, factoryConfig director.FactoryConfig, taskReporter director.TaskReporter, fileReporter director.FileReporter) (Director, error) {
	err := factoryConfig.Validate()
	if err != nil {
		return DirectorImpl{}, fmt.Errorf("error validating Director connection config: %w", err)
	}

	client, err := f.httpClient(allProxy, factoryConfig, taskReporter, fileReporter)
	if err != nil {
		return DirectorImpl{}, err
	}

	return DirectorImpl{client: client}, nil
}

func (f Factory) httpClient(allProxy string, factoryConfig director.FactoryConfig, taskReporter director.TaskReporter, fileReporter director.FileReporter) (DirectorClient, error) {
	certPool, err := factoryConfig.CACertPool()
	if err != nil {
		return DirectorClient{}, err
	}

	if certPool == nil {
		log.Debugf("Using default root CAs")
	} else {
		log.Debugf("Using custom root CAs")
	}

	rawClient := httpclient.DefaultClient(allProxy, certPool)
	authAdjustment := director.NewAuthRequestAdjustment(
		factoryConfig.TokenFunc,
		factoryConfig.Client,
		factoryConfig.ClientSecret,
	)
	rawClient.CheckRedirect = func(req *gohttp.Request, via []*gohttp.Request) error {
		if len(via) > 10 {
			return fmt.Errorf("too many redirects")
		}

		// Since redirected requests are not retried,
		// forcefully adjust auth token as this is the last chance.
		err := authAdjustment.Adjust(req, true)
		if err != nil {
			return err
		}

		req.URL.Host = net.JoinHostPort(factoryConfig.Host, fmt.Sprintf("%d", factoryConfig.Port))

		clearHeaders(req)
		clearBody(req)

		return nil
	}

	retryClient := httpclient.NewNetworkSafeRetryClient(rawClient, 5, 500*time.Millisecond)

	authedClient := httpclient.NewAdjustableClient(retryClient, authAdjustment)

	httpOpts := httpclient.Opts{NoRedactUrlQuery: true}
	httpClient := httpclient.NewHTTPClientOpts(authedClient, httpOpts)

	endpoint := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(factoryConfig.Host, fmt.Sprintf("%d", factoryConfig.Port)),
	}

	return NewClient(endpoint.String(), httpClient, taskReporter, fileReporter), nil
}

func clearBody(req *gohttp.Request) {
	req.Body = nil
}

func clearHeaders(req *gohttp.Request) {
	authValue := req.Header.Get("Authorization")
	req.Header = make(map[string][]string)
	if authValue != "" {
		req.Header.Add("Authorization", authValue)
	}
}
