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
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"

	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	socksproxy "github.com/cloudfoundry/socks5-proxy"
)

func SOCKS5DialContextFuncFromAllProxy(allProxy string, socks5Proxy boshhttp.ProxyDialer) boshhttp.DialContextFunc {
	allProxy = strings.TrimPrefix(allProxy, "ssh+")

	proxyURL, err := url.Parse(allProxy)
	if err != nil {
		return errorDialFunc(err, "Parsing bosh all_proxy url")
	}
	queryMap, err := url.ParseQuery(proxyURL.RawQuery)
	if err != nil {
		return errorDialFunc(err, "Parsing bosh all_proxy query params")
	}

	username := ""
	if proxyURL.User != nil {
		username = proxyURL.User.Username()
	}

	proxySSHKeyPath := queryMap.Get("private-key")
	if proxySSHKeyPath == "" {
		return errorDialFunc(
			fmt.Errorf("required query param 'private-key' not found"),
			"error parsing bosh all_proxy query params",
		)
	}

	proxySSHKey, err := os.ReadFile(proxySSHKeyPath)
	if err != nil {
		return errorDialFunc(err, "Reading private key file for SOCKS5 Proxy")
	}

	var (
		dialer socksproxy.DialFunc
		mut    sync.RWMutex
	)
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		mut.RLock()
		haveDialer := dialer != nil
		mut.RUnlock()

		if haveDialer {
			return dialer(network, address)
		}

		mut.Lock()
		defer mut.Unlock()
		if dialer == nil {
			proxyDialer, err := socks5Proxy.Dialer(username, string(proxySSHKey), proxyURL.Host)
			if err != nil {
				return nil, fmt.Errorf("error creating SOCKS5 dialer: %w", err)
			}
			dialer = proxyDialer
		}
		return dialer(network, address)
	}
}

func errorDialFunc(err error, cause string) boshhttp.DialContextFunc {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return nil, fmt.Errorf(cause+": %w", err)
	}
}
