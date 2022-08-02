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

package db_test

import (
	"database/sql"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestServiceBinding(t *testing.T) {
	spec.Run(t, "ServiceBinding", testServiceBinding, spec.Report(report.Terminal{}))
}

func testServiceBinding(t *testing.T, when spec.G, it spec.S) {
	var (
		dbConn *sql.DB
		ccdb   db.Repository
		mock   sqlmock.Sqlmock
	)

	it.Before(func() {
		RegisterTestingT(t)

		var err error

		dbConn, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		dbx := sqlx.NewDb(dbConn, "mysql")
		ccdb = &db.CloudController{
			DB:         dbx,
			SaltLength: 8,
		}
	})

	it.After(func() {
		dbConn.Close()
	})

	when("creating a service binding", func() {
		var (
			key string
			sb  cfclient.ServiceBinding
		)

		it.Before(func() {
			key = "supersecretencryptionkey"
			sb = cfclient.ServiceBinding{
				Guid: "1324",
				Credentials: map[string]string{
					"foo": "bar",
				},
				AppGuid:             "abcd",
				ServiceInstanceGuid: "wxyz",
			}
		})

		when("the db is working as expected", func() {
			it.Before(func() {
				mock.ExpectExec("INSERT INTO service_bindings").WillReturnResult(sqlmock.NewResult(1, 1))
			})

			it("works", func() {
				err := ccdb.CreateServiceBinding(sb, "dbca", key)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
}
