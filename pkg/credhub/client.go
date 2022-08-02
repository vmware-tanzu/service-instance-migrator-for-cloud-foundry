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

package credhub

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	golog "log"
	"net"
	"net/http"
	"strings"
	"time"

	boshhttp "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh/httpclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	"code.cloudfoundry.org/tlsconfig"
	"github.com/cloudfoundry/bosh-utils/crypto"
	proxy "github.com/cloudfoundry/socks5-proxy"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Client

type Client interface {
	GetCreds(ref string) (map[string][]map[string]interface{}, error)
}

//counterfeiter:generate -o fakes . ClientFactory

type ClientFactory interface {
	New(url string, credhubPort string, uaaPort string, allProxy string, caCert []byte, clientID string, clientSecret string) Client
}

type ClientFactoryFunc func(url string, credhubPort string, uaaPort string, allProxy string, caCert []byte, clientID string, clientSecret string) Client

func (f ClientFactoryFunc) New(url string, credhubPort string, uaaPort string, allProxy string, caCert []byte, clientID string, clientSecret string) Client {
	return f(url, credhubPort, uaaPort, allProxy, caCert, clientID, clientSecret)
}

type Credential struct {
}

type ClientImpl struct {
	url, credhubPort, uaaPort string
	allProxy                  string
	caCert                    []byte
	clientID, clientSecret    string
	dialContextFunc           DialContextFunc
}

func NewClientFactory() ClientFactoryFunc {
	return New
}

func New(
	url string,
	credhubPort string,
	uaaPort string,
	allProxy string,
	caCert []byte,
	clientID string,
	clientSecret string,
) Client {
	var dialContextFunc = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	if len(allProxy) > 0 {
		socks := proxy.NewSocks5Proxy(proxy.NewHostKey(), golog.New(ioutil.Discard, "", golog.LstdFlags), 1*time.Minute)
		dialContextFunc = boshhttp.SOCKS5DialContextFuncFromAllProxy(allProxy, socks)
	}

	return ClientImpl{
		allProxy:        allProxy,
		url:             url,
		credhubPort:     credhubPort,
		uaaPort:         uaaPort,
		caCert:          caCert,
		clientID:        clientID,
		clientSecret:    clientSecret,
		dialContextFunc: dialContextFunc,
	}
}

func (c ClientImpl) GetCreds(ref string) (map[string][]map[string]interface{}, error) {
	url := fmt.Sprintf("%s:%s", c.url, c.credhubPort)

	client, err := c.httpClient(c.dialContextFunc)
	if err != nil {
		return nil, err
	}

	token, err := c.getAccessToken(client)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/data?name=%s&current=true", url, ref), nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token["access_token"]))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get credentials, name '%s', status '%s'", ref, http.StatusText(res.StatusCode))
	}

	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	creds := make(map[string][]map[string]interface{})
	err = json.Unmarshal(content, &creds)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (c ClientImpl) getAccessToken(client HTTPClient) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s:%s", c.url, c.uaaPort)
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/oauth/token", url),
		strings.NewReader(
			fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials&token_format=jwt", c.clientID, c.clientSecret),
		),
	)
	if err != nil {
		log.Errorf("Error reading credhub request. %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return nil, err
	}

	token := make(map[string]interface{})
	err = json.Unmarshal(content, &token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

func (f DialContextFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c ClientImpl) httpClient(dialer DialContextFunc) (HTTPClient, error) {
	tlsConfig, err := configureTLS(c.caCert)
	if err != nil {
		log.Errorf("Error creating tls config. %v", err)
		return nil, err
	}
	return NewHTTPClient(dialer, tlsConfig), nil
}

func configureTLS(caCert []byte) (*tls.Config, error) {
	if len(caCert) == 0 {
		return nil, nil
	}
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("error getting a certificate pool to append our trusted cert to: %w", err)
	}

	certPool.AppendCertsFromPEM(caCert)
	caCertPool, err := crypto.CertPoolFromPEM(caCert)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
	).Client(tlsconfig.WithAuthority(caCertPool))
	if err != nil {
		return nil, err
	}

	return tlsConfig, nil
}
