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
	"io/ioutil"
	"testing"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestCalculateSaltLength(t *testing.T) {
	spec.Run(t, "CalculateSaltLength", testCalculateSaltLength, spec.Report(report.Terminal{}))
}

func testCalculateSaltLength(t *testing.T, when spec.G, it spec.S) {
	var (
		dbConn *sql.DB
		ccdb   *db.CloudController
		mock   sqlmock.Sqlmock
	)

	it.Before(func() {
		RegisterTestingT(t)

		var err error

		dbConn, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		dbx := sqlx.NewDb(dbConn, "mysql")
		ccdb = &db.CloudController{
			DB: dbx,
		}
	})

	it.After(func() {
		dbConn.Close()
	})

	when("There are existing service instances", func() {
		it.Before(func() {
			rows := sqlmock.NewRows([]string{"salt_length"}).AddRow(8)
			mock.ExpectQuery("FROM `service_instances`").WillReturnRows(rows)
		})

		it("returns the appropriate salt length", func() {
			err := ccdb.CalculateSaltLength()
			Expect(err).NotTo(HaveOccurred())
			Expect(ccdb.SaltLength).To(Equal(8))
		})
	})

	when("There are no SIs but existing service brokers", func() {
		it.Before(func() {
			siRows := sqlmock.NewRows([]string{"salt_length"})
			mock.ExpectQuery("FROM `service_instances`").WillReturnRows(siRows)
			sbRows := sqlmock.NewRows([]string{"salt_length"}).AddRow(8)
			mock.ExpectQuery("FROM `service_brokers`").WillReturnRows(sbRows)
		})

		it("returns the appropriate salt length", func() {
			err := ccdb.CalculateSaltLength()
			Expect(err).NotTo(HaveOccurred())
			Expect(ccdb.SaltLength).To(Equal(8))
		})
	})

	when("There are no SIs and no service brokers", func() {
		it.Before(func() {
			siRows := sqlmock.NewRows([]string{"salt_length"})
			mock.ExpectQuery("FROM `service_instances`").WillReturnRows(siRows)
			sbRows := sqlmock.NewRows([]string{"salt_length"})
			mock.ExpectQuery("FROM `service_brokers`").WillReturnRows(sbRows)
		})

		it("returns the appropriate salt length", func() {
			err := ccdb.CalculateSaltLength()
			Expect(err).To(HaveOccurred())
			Expect(ccdb.SaltLength).To(BeZero())
		})
	})
}
