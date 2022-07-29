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

package config

const (
	SCP   = "scp"
	Minio = "minio"
	S3    = "s3"
)

type Config struct {
	Type            string `yaml:"backup_type"`
	BackupDirectory string `yaml:"backup_directory"`
	Minio           struct {
		Alias      string `yaml:"alias" default:"minio"`
		URL        string `yaml:"url"`
		AccessKey  string `yaml:"access_key"`
		SecretKey  string `yaml:"secret_key"`
		Insecure   bool   `yaml:"insecure,omitempty"`
		BucketName string `yaml:"bucket_name"`
		BucketPath string `yaml:"bucket_path" default:"p.mysql"`
		Api        string `yaml:"api" default:"S3v4"`
		Path       string `yaml:"path" default:"auto"`
	} `yaml:"minio,omitempty"`
	S3 struct {
		Endpoint        string `yaml:"endpoint"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
		Region          string `yaml:"region" default:"us-east-1"`
		BucketName      string `yaml:"bucket_name"`
		BucketPath      string `yaml:"bucket_path" default:"p.mysql"`
		Insecure        bool   `yaml:"insecure,omitempty"`
		ForcePathStyle  bool   `yaml:"force_path_style,omitempty"`
	} `yaml:"s3,omitempty"`
	SCP struct {
		Username             string `yaml:"username"`
		Hostname             string `yaml:"hostname"`
		DestinationDirectory string `yaml:"destination_directory"`
		Port                 int    `yaml:"port"`
		PrivateKey           string `yaml:"private_key"`
	} `yaml:"scp,omitempty"`
}
