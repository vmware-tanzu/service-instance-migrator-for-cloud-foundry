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

package om

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"

	boshhttp "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
)

type Factory struct{}

func NewFactory() Factory {
	return Factory{}
}

func (f Factory) New(config Config) (OpsManager, error) {
	err := config.Validate()
	if err != nil {
		return opsManagerImpl{}, fmt.Errorf("failed to validate opsman client config: %w", err)
	}

	client, err := f.httpClient(config)
	if err != nil {
		return opsManagerImpl{}, err
	}

	return opsManagerImpl{client: client}, nil
}

func (f Factory) httpClient(config Config) (httpclient.OpsManHTTPClient, error) {
	certPool, err := config.CACertPool()
	if err != nil {
		return httpclient.OpsManHTTPClient{}, err
	}

	if certPool == nil {
		log.Debugln("Using default root CAs")
	} else {
		log.Debugln("Using custom root CAs")
	}

	var rawClient *http.Client
	if certPool == nil || len(config.AllProxy) == 0 {
		rawClient = boshhttp.InsecureClient()
	} else {
		rawClient = boshhttp.DefaultClient(config.AllProxy, certPool)
	}
	retryClient := boshhttp.NewNetworkSafeRetryClient(rawClient, 5, 500*time.Millisecond)
	authedClient := boshhttp.NewAdjustableClient(retryClient, boshhttp.NewAuthRequestAdjustment(
		config.TokenFunc,
		"",
		"",
	))
	httpClient := boshhttp.NewHTTPClientOpts(authedClient, boshhttp.Opts{NoRedactUrlQuery: true})

	endpoint := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(config.Host, fmt.Sprintf("%d", config.Port)),
	}

	return httpclient.NewHTTPClient(
		endpoint.String(),
		httpClient,
	), nil
}
