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

package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/crypto"
)

var (
	data string
	salt string
	key  string
)

func init() {
	flag.StringVar(&data, "data", "", "data to encrypt or decrypt")
	flag.StringVar(&salt, "salt", "", "salt used in encryption")
	flag.StringVar(&key, "key", "", "encryption key used to encrypt data")
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	// Ignore errors; CommandLine is set for ExitOnError.
	_ = flag.CommandLine.Parse(os.Args[2:])
	checkRequiredArgs([]string{"data", "salt", "key"})

	if os.Args[1] == "encrypt" {
		val, err := crypto.Encrypt(data, salt, key)
		if err != nil {
			log.Fatalf("error %s, failed to encrypt data: %s, using key %s, salt: %s", err, data, key, salt)
		}
		fmt.Println(val)
		os.Exit(0)
	}

	if os.Args[1] == "decrypt" {
		val, err := crypto.Decrypt(data, salt, key)
		if err != nil {
			log.Fatalf("error %s, failed to decrypt data: %s, using key %s, salt: %s", err, data, key, salt)
		}
		fmt.Println(val)
		os.Exit(0)
	}

	usage()
}

func checkRequiredArgs(required []string) {
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			_, _ = fmt.Fprintf(os.Stderr, "missing required -%s argument\n", req)
			os.Exit(2)
		}
	}
}

func usage() {
	fmt.Println("Usage: si-crypto [encrypt|decrypt] -data='some data' -salt='1d233fa44' -key='ABIDE'")
	os.Exit(1)
}
