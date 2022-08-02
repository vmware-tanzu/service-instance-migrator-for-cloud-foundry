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

package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	mysql "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/config"
)

func NewGetObjectInput(key string, bucket string) *s3.GetObjectInput {
	return &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
}

func NewDownloader(cr config.MigrationReader) (*s3manager.Downloader, error) {
	m, err := cr.GetMigration()
	if err != nil {
		return nil, err
	}
	var conf mysql.Config
	cfg := config.NewMapDecoder(conf).Decode(*m, "mysql").(mysql.Config)
	sess := session.Must(session.NewSession(aws.NewConfig().
		WithMaxRetries(3).
		WithCredentials(credentials.NewStaticCredentials(
			cfg.S3.AccessKeyID,
			cfg.S3.SecretAccessKey,
			"",
		)).
		WithEndpoint(cfg.S3.Endpoint).
		WithRegion(cfg.S3.Region).
		WithDisableSSL(cfg.S3.Insecure).
		WithS3ForcePathStyle(cfg.S3.ForcePathStyle)),
	)

	return s3manager.NewDownloader(sess), nil
}
