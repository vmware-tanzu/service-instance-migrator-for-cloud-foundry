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

package net

import (
	"errors"
	"fmt"
	"net"
	gourl "net/url"
	"strconv"
	"strings"
)

func ParseURL(url string) (string, string, int, string, error) {
	if len(url) == 0 {
		return "", "", 0, "", errors.New("expected non-empty URL")
	}

	parsedURL, err := gourl.Parse(url)
	if err != nil {
		return "", "", 0, "", fmt.Errorf(fmt.Sprintf("error parsing URL '%s'", url)+": %w", err)
	}

	defaultScheme := "https"
	scheme := parsedURL.Scheme
	host := parsedURL.Host
	port := 443
	path := parsedURL.Path

	if len(scheme) == 0 {
		scheme = defaultScheme
	}

	if len(host) == 0 {
		host = url
		path = ""
		if strings.Contains(url, "/") {
			hostPath := strings.Split(url, "/")
			if len(hostPath) == 2 {
				host = hostPath[0]
				path = "/" + hostPath[1]
			}
		}
	}

	if strings.Contains(host, ":") {
		var portStr string

		host, portStr, err = net.SplitHostPort(host)
		if err != nil {
			return "", "", 0, "", fmt.Errorf(fmt.Sprintf("error extracting host/port from URL '%s'", url)+": %w", err)
		}

		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", "", 0, "", fmt.Errorf(fmt.Sprintf("error extracting port from URL '%s'", url)+": %w", err)
		}
	}

	if scheme == host {
		scheme = defaultScheme
	}

	if len(host) == 0 {
		return "", "", 0, "", fmt.Errorf("expected to extract host from URL '%s'", url)
	}

	return scheme, host, port, path, nil
}
