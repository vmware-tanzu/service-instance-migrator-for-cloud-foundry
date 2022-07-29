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

package uaa

import (
	"fmt"
	"net"
	gohttp "net/http"
	"net/url"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type FactoryImpl struct{}

func NewFactory() FactoryImpl {
	return FactoryImpl{}
}

func (f FactoryImpl) New(config Config) (UAA, error) {
	err := config.Validate()
	if err != nil {
		return uaaImpl{}, fmt.Errorf("error validating UAA connection config: %w", err)
	}

	client, err := f.httpClient(config)
	if err != nil {
		return uaaImpl{}, err
	}

	return uaaImpl{client: client}, nil
}

func (f FactoryImpl) httpClient(config Config) (Client, error) {
	certPool, err := config.CACertPool()
	if err != nil {
		return Client{}, err
	}

	if certPool == nil {
		log.Debugf("Using default root CAs")
	} else {
		log.Debugf("Using custom root CAs")
	}

	var rawClient *gohttp.Client
	if len(config.AllProxy) > 0 {
		rawClient = httpclient.DefaultClient(config.AllProxy, certPool)
	} else {
		rawClient = httpclient.InsecureClient()
	}
	retryClient := httpclient.NewNetworkSafeRetryClient(rawClient, 5, 500*time.Millisecond)

	httpClient := httpclient.NewHTTPClient(retryClient)

	endpoint := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(config.Host, fmt.Sprintf("%d", config.Port)),
		Path:   config.Path,
	}

	return NewClient(endpoint.String(), config.ClientID, config.ClientSecret, httpClient), err
}
