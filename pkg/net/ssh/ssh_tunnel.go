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

package ssh

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// NewTunnel will create a string of the form "host:port", and will
// establish an SSH tunnel if necessary to create a string that can be used to
// connect to a remote machine
func NewTunnel(host, tunnelHost, tunnelUser,
	tunnelPassword, tunnelPrivateKey string, tunnelRequired bool) (*SSH, error) {
	if !tunnelRequired {
		log.Debugln("SSH tunneling not required, skipping tunnel creation")
		return nil, nil
	}

	if tunnelHost == "" || tunnelUser == "" || (tunnelPassword == "" && tunnelPrivateKey == "") {
		return nil, errors.New("tunneling is required, but the tunnel information was not specified")
	}

	localPort := findLocalPort()
	if localPort < 0 {
		return nil, errors.New("no available ports on localhost")
	}

	localServer := Endpoint{
		Host: "localhost",
		Port: localPort,
	}

	remoteServer := Endpoint{
		Host: host,
		Port: 3306,
	}

	tunnelServer := Endpoint{
		Host: tunnelHost,
		Port: 22,
	}

	var authMethods []ssh.AuthMethod
	if tunnelPrivateKey != "" {
		log.Debugf("Using private key %s", tunnelPrivateKey)
		pem, err := os.ReadFile(tunnelPrivateKey)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(pem)
		if err != nil {
			return nil, err
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if tunnelPassword != "" {
		authMethods = append(authMethods, ssh.Password(tunnelPassword))
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	config := ssh.ClientConfig{
		Timeout: 5 * time.Second,
		User:    tunnelUser,
		Auth:    authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	return &SSH{
		Config: &config,
		Local:  &localServer,
		Server: &tunnelServer,
		Remote: &remoteServer,
	}, nil
}

func findLocalPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return -1
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
