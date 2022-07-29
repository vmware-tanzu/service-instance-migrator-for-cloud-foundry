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

package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	proxy "github.com/cloudfoundry/socks5-proxy"

	"code.cloudfoundry.org/tlsconfig"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func DefaultClient(allProxy string, certPool *x509.CertPool) *http.Client {
	return New(allProxy, false, false, true, certPool)
}

func InsecureClient() *http.Client {
	return New("", true, false, true, nil)
}

func New(
	allProxy string,
	insecureSkipVerify,
	externalClient bool,
	disableKeepAlives bool,
	certPool *x509.CertPool,
) *http.Client {
	socks5Proxy := proxy.NewSocks5Proxy(proxy.NewHostKey(), log.New(ioutil.Discard, "", log.LstdFlags), 1*time.Minute)
	dialContextFunc := boshhttp.SOCKS5DialContextFuncFromEnvironment(&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}, socks5Proxy)

	if len(allProxy) > 0 {
		dialContextFunc = SOCKS5DialContextFuncFromAllProxy(allProxy, socks5Proxy)
	}

	serviceDefaults := tlsconfig.WithInternalServiceDefaults()
	if externalClient {
		serviceDefaults = tlsconfig.WithExternalServiceDefaults()
	}

	tlsConfig, err := tlsconfig.Build(
		serviceDefaults,
		WithInsecureSkipVerify(insecureSkipVerify),
		WithClientSessionCache(0),
	).Client(tlsconfig.WithAuthority(certPool))
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         dialContextFunc,
			TLSHandshakeTimeout: 30 * time.Second,
			DisableKeepAlives:   disableKeepAlives,
		},
	}

	return client
}

func WithInsecureSkipVerify(insecureSkipVerify bool) tlsconfig.TLSOption {
	return func(config *tls.Config) error {
		config.InsecureSkipVerify = insecureSkipVerify
		return nil
	}
}

func WithClientSessionCache(capacity int) tlsconfig.TLSOption {
	return func(config *tls.Config) error {
		config.ClientSessionCache = tls.NewLRUClientSessionCache(capacity)
		return nil
	}
}
