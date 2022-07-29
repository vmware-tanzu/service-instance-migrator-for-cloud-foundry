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

package db

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/net/ssh"
	"sync"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/jmoiron/sqlx"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Repository

// Repository calls operations on the CCDB
type Repository interface {
	ServiceInstanceExists(guid string) (bool, error)
	CreateServiceInstance(si cfclient.ServiceInstance, targetSpace cfclient.Space, targetPlan cfclient.ServicePlan, targetService cfclient.Service, key string) error
	DeleteServiceInstance(spaceGUID string, serviceInstanceGUID string) (bool, error)
	CreateServiceBinding(binding cfclient.ServiceBinding, appGUID string, encryptionKey string) error
}

// NewCCDBConnection will establish a connection to the CCDB on the host with the specified
// username and password
func NewCCDBConnection(dbHost string, dbUsername string, dbPassword string,
	tunnelHost string, tunnelUser string, tunnelPassword string, tunnelPrivateKey string,
	tunnelRequired bool) (*sqlx.DB, error) {

	log.Debugln("About to create SSH NewTunnel")
	sshTunnel, err := ssh.NewTunnel(dbHost, tunnelHost, tunnelUser, tunnelPassword, tunnelPrivateKey, tunnelRequired)

	if err != nil {
		return nil, err
	}

	hostPort := fmt.Sprintf("%s:3306", dbHost)
	wg := &sync.WaitGroup{}
	if sshTunnel != nil {
		wg.Add(1)
		log.Debugln("About to start tunnel")
		go func() {
			err := sshTunnel.Start(wg)
			if err != nil {
				panic(err)
			}
		}()
		hostPort = fmt.Sprintf("localhost:%d", sshTunnel.Local.Port)
	}

	wg.Wait()
	log.Debugln("About to connect to DB")
	connectStr := fmt.Sprintf("%s:%s@(%s)/ccdb", dbUsername, dbPassword, hostPort)
	return sqlx.Connect("mysql", connectStr)
}

// CloudController is the concrete struct that will implement Repository
type CloudController struct {
	DB         *sqlx.DB
	SaltLength int
}

// CalculateSaltLength will look in the CCDB for an existing salt
// to get its length, as it will be static throughout the DB
func (d *CloudController) CalculateSaltLength() error {
	if d.SaltLength > 0 {
		return nil
	}

	type lenVal struct {
		Len int `db:"salt_length"`
	}

	retVal := []lenVal{}
	err := d.DB.Select(&retVal, "SELECT LENGTH(`salt`) AS salt_length FROM `service_instances` LIMIT 1;")
	if err != nil {
		return err
	}

	if len(retVal) > 0 {
		d.SaltLength = retVal[0].Len
		return nil
	}

	err = d.DB.Select(&retVal, "SELECT LENGTH(`salt`) AS salt_length FROM `service_brokers` LIMIT 1;")
	if err != nil {
		return err
	}

	if len(retVal) > 0 {
		d.SaltLength = retVal[0].Len
		return nil
	}

	return errors.New("could not determine length of salt string")
}

// GenerateSalt will create a random hex string with the appropriate
// length for the foundation
func (d *CloudController) GenerateSalt() (string, error) {
	err := d.CalculateSaltLength()
	if err != nil {
		return "", err
	}
	saltBytes := make([]byte, d.SaltLength/2)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", saltBytes), nil
}

// NewNullString will return an empty sql.NullString if s is empty, and an
// sql.NullString with s in it if not
func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// NewNullInt64 will return an empty sql.NullInt64 if i is 0, and an sql.NullInt64
// with i in it if it is not
func NewNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}

	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}
