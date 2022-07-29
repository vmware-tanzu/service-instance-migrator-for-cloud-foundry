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
	"errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCreateServiceInstance(t *testing.T) {
	spec.Run(t, "CreateServiceInstance", testCreateServiceInstance, spec.Report(report.Terminal{}))
}

func TestServiceInstanceExists(t *testing.T) {
	spec.Run(t, "ServiceInstanceExists", testServiceInstanceExists, spec.Report(report.Terminal{}))
}

func TestDeleteServiceInstance(t *testing.T) {
	spec.Run(t, "DeleteServiceInstance", testDeleteServiceInstance, spec.Report(report.Terminal{}))
}

func testServiceInstanceExists(t *testing.T, when spec.G, it spec.S) {
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

	when("retrieving a service instance", func() {
		var (
			si cfclient.ServiceInstance
		)

		it.Before(func() {
			si = cfclient.ServiceInstance{
				Guid: "123",
				Name: "some-si",
			}
		})

		when("the service instance exists in the database", func() {
			it.Before(func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow("123")
				mock.ExpectQuery(`SELECT id FROM service_instances`).WithArgs("123").WillReturnRows(rows)
			})

			it("it exists", func() {
				exists, err := ccdb.ServiceInstanceExists(si.Guid)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		when("the service instance cannot be found in the database", func() {
			it.Before(func() {
				rows := sqlmock.NewRows([]string{"id"})
				mock.ExpectQuery(`SELECT id FROM service_instances`).WithArgs("123").WillReturnRows(rows)
			})

			it("it does not exist", func() {
				exists, err := ccdb.ServiceInstanceExists(si.Guid)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})
	})
}

func testCreateServiceInstance(t *testing.T, when spec.G, it spec.S) {
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

	when("creating a service instance", func() {
		var (
			key           string
			si            cfclient.ServiceInstance
			targetPlan    cfclient.ServicePlan
			targetService cfclient.Service
			targetSpace   cfclient.Space
		)

		it.Before(func() {
			key = "supersecretencryptionkey"

			si = cfclient.ServiceInstance{
				Guid:            "123",
				Name:            "my-si",
				SpaceGuid:       "abc",
				ServiceGuid:     "001",
				ServicePlanGuid: "999",
				Credentials: map[string]interface{}{
					"foo": "bar",
				},
				Tags: []string{"tag-1"},
			}

			targetPlan = cfclient.ServicePlan{
				Guid:        "999",
				Name:        "my-plan",
				UniqueId:    "qwerty",
				ServiceGuid: "001",
			}

			targetService = cfclient.Service{
				Guid:              "001",
				Label:             "my-service",
				ServiceBrokerGuid: "def",
			}

			targetSpace = cfclient.Space{
				Guid:             "abc",
				Name:             "my space",
				OrganizationGuid: "org-id",
			}
		})

		when("the db is working as expected", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO service_usage_events`).WithArgs(sqlmock.AnyArg(), "CREATED", "org-id", "abc", "my space", "123", "my-si", "managed_service_instance", "999", "my-plan", "001", "my-service").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			})

			it("works", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		when("the target space cannot be found in the DB", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"})
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the target plan cannot be found in the DB", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"})
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the transaction fails to start", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin().WillReturnError(errors.New("test-error"))
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the service instance fails to be inserted", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnError(errors.New("test-error"))
				mock.ExpectRollback()
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the service usage event fails to be inserted", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO service_usage_events`).WithArgs(sqlmock.AnyArg(), "CREATED", "org-id", "abc", "my space", "123", "my-si", "managed_service_instance", "999", "my-plan", "001", "my-service").WillReturnError(errors.New("test-error"))

				mock.ExpectRollback()
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the commit fails", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO service_usage_events`).WithArgs(sqlmock.AnyArg(), "CREATED", "org-id", "abc", "my space", "123", "my-si", "managed_service_instance", "999", "my-plan", "001", "my-service").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("test-error"))
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the rollback fails", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnError(errors.New("test-error"))
				mock.ExpectRollback().WillReturnError(errors.New("test-rollback-error"))
			})

			it("fails", func() {
				err := ccdb.CreateServiceInstance(si, targetSpace, targetPlan, targetService, key)
				Expect(err).To(HaveOccurred())
			})
		})
	})
}

func testDeleteServiceInstance(t *testing.T, when spec.G, it spec.S) {
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
			DB: dbx,
		}
	})

	it.After(func() {
		dbConn.Close()
	})

	when("deleting a service instance", func() {
		var (
			si    cfclient.ServiceInstance
			space cfclient.Space
		)

		it.Before(func() {
			si = cfclient.ServiceInstance{
				Guid: "86a002bb-7b8c-4961-8194-7ece614c2638",
			}
			space = cfclient.Space{
				Guid:             "58d29ff4-5d7e-4357-b244-60b7e4c34025",
				Name:             "my-space",
				OrganizationGuid: "my-org",
			}
		})

		when("the db is working as expected", func() {
			it.Before(func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				sharedInstancesRows := sqlmock.NewRows([]string{"service_instance_guid", "target_space_guid"})
				mock.ExpectQuery(regexp.QuoteMeta(db.GetServiceInstanceSharesQuery)).WithArgs(si.Guid, space.Guid).WillReturnRows(sharedInstancesRows)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("58d29ff4-5d7e-4357-b244-60b7e4c34025").WillReturnRows(spaceRows)
				mock.ExpectQuery(`SELECT id FROM service_instances`).WithArgs(si.Guid).WillReturnRows(rows)
				mock.ExpectExec(db.DeleteServiceBindingsSQLStatement).WithArgs(si.Guid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(db.DeleteServiceKeysSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(db.DeleteServiceInstanceOperationsSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(db.DeleteServiceInstanceSQLStatement)).WithArgs(si.Guid, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			})

			it("deletes the service instance", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).NotTo(HaveOccurred())
				Expect(isDeleted).To(BeTrue())
			})
		})

		when("the service instance is not bound to any apps", func() {
			it.Before(func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				sharedInstancesRows := sqlmock.NewRows([]string{"service_instance_guid", "target_space_guid"})
				mock.ExpectQuery(regexp.QuoteMeta(db.GetServiceInstanceSharesQuery)).WithArgs(si.Guid, space.Guid).WillReturnRows(sharedInstancesRows)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("58d29ff4-5d7e-4357-b244-60b7e4c34025").WillReturnRows(spaceRows)
				mock.ExpectQuery(`SELECT id FROM service_instances`).WithArgs(si.Guid).WillReturnRows(rows)
				mock.ExpectExec(db.DeleteServiceBindingsSQLStatement).WithArgs(si.Guid).WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec(db.DeleteServiceKeysSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(db.DeleteServiceInstanceOperationsSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(db.DeleteServiceInstanceSQLStatement)).WithArgs(si.Guid, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			})

			it("still deletes the service instance", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).NotTo(HaveOccurred())
				Expect(isDeleted).To(BeTrue())
			})
		})

		when("the service instance is shared across spaces", func() {
			it.Before(func() {
				mock.ExpectBegin()
				sharedInstancesRows := sqlmock.NewRows([]string{"service_instance_guid", "target_space_guid"}).AddRow(si.Guid, space.Guid)
				mock.ExpectQuery(regexp.QuoteMeta(db.GetServiceInstanceSharesQuery)).WithArgs(si.Guid, space.Guid).WillReturnRows(sharedInstancesRows)
				mock.ExpectCommit()
			})

			it("does not delete the service instance", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, db.ErrUnsupportedOperation)).To(BeTrue())
				Expect(isDeleted).To(BeFalse())
			})
		})

		when("no rows were deleted from service_instance_operations", func() {
			it.Before(func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				sharedInstancesRows := sqlmock.NewRows([]string{"service_instance_guid", "target_space_guid"})
				mock.ExpectQuery(regexp.QuoteMeta(db.GetServiceInstanceSharesQuery)).WithArgs(si.Guid, space.Guid).WillReturnRows(sharedInstancesRows)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("58d29ff4-5d7e-4357-b244-60b7e4c34025").WillReturnRows(spaceRows)
				mock.ExpectQuery(`SELECT id FROM service_instances`).WithArgs(si.Guid).WillReturnRows(rows)
				mock.ExpectExec(db.DeleteServiceBindingsSQLStatement).WithArgs(si.Guid).WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec(db.DeleteServiceKeysSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(db.DeleteServiceInstanceOperationsSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec(regexp.QuoteMeta(db.DeleteServiceInstanceSQLStatement)).WithArgs(si.Guid, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			})

			it("still deletes the service instance", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).NotTo(HaveOccurred())
				Expect(isDeleted).To(BeTrue())
			})
		})

		when("no rows were deleted from service_instances", func() {
			it.Before(func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				sharedInstancesRows := sqlmock.NewRows([]string{"service_instance_guid", "target_space_guid"})
				mock.ExpectQuery(regexp.QuoteMeta(db.GetServiceInstanceSharesQuery)).WithArgs(si.Guid, space.Guid).WillReturnRows(sharedInstancesRows)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("58d29ff4-5d7e-4357-b244-60b7e4c34025").WillReturnRows(spaceRows)
				mock.ExpectQuery(`SELECT id FROM service_instances`).WithArgs(si.Guid).WillReturnRows(rows)
				mock.ExpectExec(db.DeleteServiceBindingsSQLStatement).WithArgs(si.Guid).WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec(db.DeleteServiceKeysSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(db.DeleteServiceInstanceOperationsSQLStatement).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec(regexp.QuoteMeta(db.DeleteServiceInstanceSQLStatement)).WithArgs(si.Guid, 1).WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			})

			it("still deletes the service instance", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).NotTo(HaveOccurred())
				Expect(isDeleted).To(BeTrue())
			})
		})

		when("the commit fails", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO service_usage_events`).WithArgs(sqlmock.AnyArg(), "CREATED", "org-id", "abc", "my space", "123", "my-si", "managed_service_instance", "999", "my-plan", "001", "my-service").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("test-error"))
			})

			it("fails", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).To(HaveOccurred())
				Expect(isDeleted).To(BeFalse())
			})
		})

		when("the rollback fails", func() {
			it.Before(func() {
				spaceRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id FROM spaces`).WithArgs("abc").WillReturnRows(spaceRows)

				planRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`SELECT id FROM service_plans`).WithArgs("999").WillReturnRows(planRows)

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO service_instances`).WithArgs("123", "my-si", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), `["tag-1"]`, sqlmock.AnyArg()).WillReturnError(errors.New("test-error"))
				mock.ExpectRollback().WillReturnError(errors.New("test-rollback-error"))
			})

			it("fails", func() {
				isDeleted, err := ccdb.DeleteServiceInstance(space.Guid, si.Guid)
				Expect(err).To(HaveOccurred())
				Expect(isDeleted).To(BeFalse())
			})
		})
	})
}
