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

package cc

import (
	// initialize the MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

type DefaultCloudControllerServiceFactory struct {
	holder           ClientHolder
	manifestExporter ManifestExporter
}

func NewCloudControllerServiceFactory(h ClientHolder, manifestExporter ManifestExporter) DefaultCloudControllerServiceFactory {
	return DefaultCloudControllerServiceFactory{
		holder:           h,
		manifestExporter: manifestExporter,
	}
}

func (f DefaultCloudControllerServiceFactory) NewCloudControllerService(cfg *Config, isExport bool) (Service, error) {
	ccdb, err := NewDatabaseFactory(cfg).NewCCDB(isExport)
	if err != nil {
		return nil, err
	}

	return NewCloudControllerService(ccdb, f.holder.CFClient(isExport), f.manifestExporter), nil
}
