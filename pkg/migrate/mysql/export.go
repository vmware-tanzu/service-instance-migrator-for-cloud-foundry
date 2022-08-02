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

package mysql

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/flow"
	sio "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	mysql "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/mysql/s3"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
)

type BackupDateTimeExtractor func(s string) (string, string, error)
type BackupIDExtractor func(s string) (string, error)
type BackupFilenameExtractor func(s string) (string, error)
type EncryptionKeyExtractor func(s string) (string, error)

func NewExportSequence(api, org, space string, instance *cf.ServiceInstance, om config.OpsManager, downloader s3.ObjectDownloader, executor exec.Executor, exportDir string) flow.Flow {
	cfHome, err := os.MkdirTemp("", instance.GUID)
	if err != nil {
		panic("failed to create CF_HOME")
	}

	return flow.ProgressBarSequence(
		fmt.Sprintf("Exporting %s", instance.Name),
		flow.StepWithProgressBar(
			cf.LoginSourceFoundation(executor, om, api, org, space, cfHome),
			flow.WithDisplay("Logging into source foundation"),
		),
		flow.StepWithProgressBar(
			InstallADBRPlugin(executor, cfHome),
			flow.WithDisplay("Installing adbr plugin"),
		),
		flow.StepWithProgressBar(
			CreateBackup(executor, cfHome, *instance),
			flow.WithDisplay("Creating backup"),
		),
		flow.StepWithProgressBar(
			GetBackupStatus(executor, cfHome, *instance, 30*time.Minute, 10*time.Second),
			flow.WithDisplay("Waiting for backup"),
		),
		flow.StepWithProgressBar(
			GetLatestBackup(executor, cfHome, *instance),
			flow.WithDisplay("Getting latest backup"),
		),
		flow.StepWithProgressBar(
			DownloadBackup(executor, instance, downloader, backupDateTimeExtractor, backupIDExtractor, sio.NewFileSystemHelper(), exportDir),
			flow.WithDisplay("Downloading backup"),
		),
		flow.StepWithProgressBar(
			RetrieveEncryptionKey(executor, om, instance, encryptionKeyExtractor),
			flow.WithDisplay("Retrieving encryption key"),
		),
	)
}

func InstallADBRPlugin(e exec.Executor, cfHome string) flow.StepFunc {
	return func(ctx context.Context, cfg interface{}, dryRun bool) (flow.Result, error) {
		log.Infoln("Installing adbr plugin")
		lines := []string{
			fmt.Sprintf("CF_HOME='%s' cf install-plugin -r CF-Community \"ApplicationDataBackupRestore\" -f\n", cfHome),
		}
		res, err := e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
		if err != nil {
			return exec.Result{}, errors.Wrap(err, "failed to install adbr plugin")
		}

		return res, nil
	}
}

func CreateBackup(e exec.Executor, cfHome string, instance cf.ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, cfg interface{}, dryRun bool) (flow.Result, error) {
		log.Infoln("Creating backup")
		lines := []string{
			fmt.Sprintf("CF_HOME='%s' cf adbr backup %q", cfHome, instance.Name),
		}
		res, err := e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
		if err != nil {
			return exec.Result{}, errors.Wrap(err, fmt.Sprintf("failed to backup instance %q", instance.Name))
		}

		return res, nil
	}
}

func GetBackupStatus(e exec.Executor, cfHome string, instance cf.ServiceInstance, timeout time.Duration, pause time.Duration) flow.StepFunc {
	return func(ctx context.Context, cfg interface{}, dryRun bool) (flow.Result, error) {
		log.Infoln("Waiting for backup")
		res, err := executeForDuration(ctx, e, []string{
			fmt.Sprintf("CF_HOME='%s' cf adbr get-status %q", cfHome, instance.Name),
		}, timeout, pause, func(result exec.Result) bool {
			return strings.Contains(result.Output, "Backup was successful")
		}, func(result exec.Result) (bool, error) {
			return strings.Contains(result.Output, "Backup failed"), fmt.Errorf("adbr failed to backup instance %q", instance.Name)
		})
		if err != nil {
			return exec.Result{}, err
		}

		return res, nil
	}
}

func GetLatestBackup(e exec.Executor, cfHome string, instance cf.ServiceInstance) flow.StepFunc {
	return func(ctx context.Context, cfg interface{}, dryRun bool) (flow.Result, error) {
		log.Infoln("Getting latest backup")
		lines := []string{
			fmt.Sprintf("CF_HOME='%s' cf adbr list-backups %q -l 1", cfHome, instance.Name),
		}
		res, err := e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
		if err != nil {
			return exec.Result{}, fmt.Errorf("failed to list adbr backups: %w", err)
		}

		return res, nil
	}
}

func DownloadBackup(e exec.Executor, instance *cf.ServiceInstance, downloader s3.ObjectDownloader, dateTimeExtractor BackupDateTimeExtractor, idExtractor BackupIDExtractor, fso sio.FileSystemOperations, exportDir string) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		m, ok := c.(*config.Migration)
		if !ok {
			log.Fatal("failed to convert type to config.Migration")
		}

		var conf mysql.Config
		cfg := config.NewMapDecoder(conf).Decode(*m, "mysql").(mysql.Config)

		if len(cfg.Type) == 0 {
			return exec.Result{}, fmt.Errorf("required property 'backup_type' is not set in si-migrator.yml for mysql backups")
		}

		if cfg.BackupDirectory == "" {
			cfg.BackupDirectory = exportDir
		}
		log.Infof("Downloading latest backup to %q", cfg.BackupDirectory)

		switch cfg.Type {
		case mysql.SCP:
			return exec.Result{}, scpDownload(ctx, cfg, e, instance, dateTimeExtractor, idExtractor, dryRun)
		case mysql.S3:
			return exec.Result{}, s3Download(ctx, cfg, e, instance, downloader, dateTimeExtractor, idExtractor, fso)
		case mysql.Minio:
			return exec.Result{}, minioDownload(ctx, cfg, e, instance, dateTimeExtractor, idExtractor, dryRun)
		}

		return exec.Result{}, fmt.Errorf("failed to backup strategy for type '%s'", cfg.Type)
	}
}

func RetrieveEncryptionKey(e exec.Executor, om config.OpsManager, instance *cf.ServiceInstance, encryptionKeyExtractor EncryptionKeyExtractor) flow.StepFunc {
	return func(ctx context.Context, c interface{}, dryRun bool) (flow.Result, error) {
		log.Infof("Retrieving encryption key")
		if err := checkRequiredParams("missing required param: %q is not set",
			map[string]string{
				"instance guid":      instance.GUID,
				"instance backup id": instance.BackupID,
			}); err != nil {
			var wrappedErr interface{ Unwrap() error }
			if errors.As(err, &wrappedErr) {
				unwrappedErr := errors.Unwrap(err)
				return exec.Result{}, unwrappedErr
			}
			return exec.Result{}, err
		}

		cfDeploymentName, err := findDeploymentName(ctx, e, om, "^cf-")
		if err != nil {
			return exec.Result{}, err
		}

		credhubAdminSecret, err := findCredhubAdminSecret(ctx, e, om, cfDeploymentName)
		if err != nil {
			return exec.Result{}, err
		}

		log.Debugf("Got credhub admin secret %q", credhubAdminSecret)
		res, err := bosh.Run(e, ctx, om, "deps", "--column=name", "|", "grep", "pivotal-mysql")
		if err != nil {
			return exec.Result{}, errors.Wrap(err, "failed to get pivotal-mysql deployment name")
		}
		status := res.Status
		log.Debugln(status.Output)
		mysqlDeploymentName := strings.TrimSuffix(status.Output, "\t\n")
		log.Debugf("Getting encryption key from %q", mysqlDeploymentName)
		sshCmd := strings.Join(
			[]string{
				fmt.Sprintln("\"/var/vcap/packages/credhub-cli/bin/credhub api https://credhub.service.cf.internal:8844 --ca-cert /var/vcap/jobs/adbr-api/config/credhub_ca.pem && \\"),
				fmt.Sprintf("/var/vcap/packages/credhub-cli/bin/credhub login --client-name credhub_admin_client --client-secret %s && \\", credhubAdminSecret),
				fmt.Sprintf("/var/vcap/packages/credhub-cli/bin/credhub get -n /tanzu-mysql/backups/%s_%s -q\"", instance.GUID, instance.BackupID),
			}, "\n")
		res, err = bosh.Run(e, ctx, om, "-d", mysqlDeploymentName, "ssh", "dedicated-mysql-broker/0", "-c", sshCmd)
		if err != nil {
			return exec.Result{}, errors.Wrap(err, fmt.Sprintf("failed to get encryption key from credhub: cmd: %q", sshCmd))
		}
		status = res.Status
		key, err := encryptionKeyExtractor(status.Output)
		if err != nil {
			return exec.Result{}, err
		}

		log.Debugf("Encrypt key is %q", key)
		instance.BackupEncryptionKey = key

		return exec.Result{}, nil
	}
}

func findDeploymentName(ctx context.Context, e exec.Executor, om config.OpsManager, pattern string) (string, error) {
	res, err := bosh.Run(e, ctx, om, "deps", "--column=name", "|", "grep", "'"+pattern+"'", "|", "tr", "-d", "'\\t\\n'")
	if err != nil {
		return "", errors.Wrap(err, "failed to get deployments")
	}
	status := res.Status
	log.Debugln(status.Output)
	deployment := strings.TrimSuffix(status.Output, "\t\n")
	match, _ := regexp.MatchString(pattern, deployment)
	if !match {
		return "", fmt.Errorf("failed to find deployment name")
	}

	return deployment, nil
}

func findCredhubAdminSecret(ctx context.Context, e exec.Executor, opsman config.OpsManager, deploymentName string) (string, error) {
	log.Debugf("Getting credhub admin credentials from %q", deploymentName)
	res, err := om.Run(e, ctx, opsman,
		fmt.Sprintf("curl -s -p /api/v0/deployed/products/%s/credentials/.uaa.credhub_admin_client_client_credentials | "+
			"jq -r .credential.value.password", deploymentName))
	if err != nil {
		return "", errors.Wrap(err, "failed to find credhub admin credentials")
	}
	status := res.Status
	credhubAdminSecret := strings.TrimSuffix(status.Output, "\n")

	return credhubAdminSecret, nil
}

func scpDownload(ctx context.Context, cfg mysql.Config, e exec.Executor, instance *cf.ServiceInstance, dateTimeExtractor BackupDateTimeExtractor, idExtractor BackupIDExtractor, dryRun bool) error {
	if err := checkRequiredParams("required param %q is not set in si-migrator.yml for scp backup",
		map[string]string{
			"backup directory":      cfg.BackupDirectory,
			"username":              cfg.SCP.Username,
			"hostname":              cfg.SCP.Hostname,
			"destination directory": cfg.SCP.DestinationDirectory,
			"private key":           cfg.SCP.PrivateKey,
		}); err != nil {
		var wrappedErr interface{ Unwrap() error }
		if errors.As(err, &wrappedErr) {
			unwrappedErr := errors.Unwrap(err)
			return unwrappedErr
		}
		return err
	}

	lines := []string{
		fmt.Sprintf("echo downloading backup %s", instance.GUID),
	}

	var backupDate, backupTime, backupID string
	var err error
	if !dryRun {
		backupDate, backupTime, backupID, err = extractBackupDetails(e.LastResult().Output, dateTimeExtractor, idExtractor)
		if err != nil {
			return err
		}
		instance.BackupID = backupID
		instance.BackupDate = backupDate
		instance.BackupTime = backupTime
	}

	log.Debugf("backupDate: %s, backupTime: %s, backupID: %s", backupDate, backupTime, backupID)

	backupDir := filepath.Join(cfg.BackupDirectory, instance.GUID, backupID)
	instance.BackupFile = filepath.Join(backupDir, "mysql-backup.tar.gpg")
	lines = append(lines, []string{
		fmt.Sprintf("mkdir -p %s", backupDir),
		fmt.Sprintf("scp -i %s %s@%s:%s/p.mysql/service-instance_%s/%s/%s_%s.tar %s",
			cfg.SCP.PrivateKey,
			cfg.SCP.Username,
			cfg.SCP.Hostname,
			cfg.SCP.DestinationDirectory,
			instance.GUID,
			backupDate,
			instance.GUID,
			backupID,
			cfg.BackupDirectory,
		),
		fmt.Sprintf("tar xvf %s/%s_%s.tar -C %s", cfg.BackupDirectory, instance.GUID, backupID, backupDir),
		fmt.Sprintf("mv %s/*.tar.gpg %s", backupDir, instance.BackupFile),
		fmt.Sprintf("rm -f %s/%s_%s.tar", cfg.BackupDirectory, instance.GUID, backupID),
	}...)

	_, err = e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		return errors.Wrap(err, "failed to download backup")
	}

	return nil
}

func s3Download(ctx context.Context, cfg mysql.Config, e exec.Executor, instance *cf.ServiceInstance, downloader s3.ObjectDownloader, dateTimeExtractor BackupDateTimeExtractor, idExtractor BackupIDExtractor, fso sio.FileSystemOperations) error {
	if err := checkRequiredParams("required param %q is not set in si-migrator.yml for minio backup",
		map[string]string{
			"backup directory":  cfg.BackupDirectory,
			"url":               cfg.S3.Endpoint,
			"access key id":     cfg.S3.AccessKeyID,
			"secret access key": cfg.S3.SecretAccessKey,
			"bucket name":       cfg.S3.BucketName,
			"bucket path":       cfg.S3.BucketPath,
			"region":            cfg.S3.Region,
		}); err != nil {
		var wrappedErr interface{ Unwrap() error }
		if errors.As(err, &wrappedErr) {
			unwrappedErr := errors.Unwrap(err)
			return unwrappedErr
		}
		return err
	}

	backupDate, backupTime, backupID, err := extractBackupDetails(e.LastResult().Output, dateTimeExtractor, idExtractor)
	if err != nil {
		return err
	}
	instance.BackupID = backupID
	instance.BackupDate = backupDate
	instance.BackupTime = backupTime

	key := fmt.Sprintf("%s/service-instance_%s/%s/%s_%s.tar",
		cfg.S3.BucketPath,
		instance.GUID,
		backupDate,
		instance.GUID,
		backupID,
	)

	backupDir := filepath.Join(cfg.BackupDirectory, instance.GUID, backupID)

	err = fso.Mkdir(backupDir)
	if err != nil {
		return err
	}

	backupFile := filepath.Join(cfg.BackupDirectory, fmt.Sprintf("%s_%s.tar", instance.GUID, instance.BackupID))

	log.Infof("Downloading backup to %q", backupFile)
	file, err := fso.Create(backupFile)
	if err != nil {
		return err
	}

	_, err = downloader.DownloadWithContext(ctx, file, s3.NewGetObjectInput(key, cfg.S3.BucketName))
	if err != nil {
		log.Errorf("Failed to download mysql backup, %v", err)
		return err
	}

	tarFile, err := fso.Open(backupFile)
	if err != nil {
		return err
	}
	defer func(tarFile *os.File) {
		_ = tarFile.Close()
	}(tarFile)

	err = fso.Untar(backupDir, tarFile)
	if err != nil {
		log.Errorf("Failed to extract file %q to %q, %v", tarFile.Name(), backupDir, err)
		return err
	}

	backupFilename := "mysql-backup.tar.gpg"
	instance.BackupFile = filepath.Join(backupDir, backupFilename)

	return renameBackupFile(fso, backupDir, backupFilename)
}

func minioDownload(ctx context.Context, cfg mysql.Config, e exec.Executor, instance *cf.ServiceInstance, dateTimeExtractor BackupDateTimeExtractor, idExtractor BackupIDExtractor, dryRun bool) error {
	if err := checkRequiredParams("required param %q is not set in si-migrator.yml for minio backup",
		map[string]string{
			"backup directory": cfg.BackupDirectory,
			"url":              cfg.Minio.URL,
			"access key":       cfg.Minio.AccessKey,
			"secret key":       cfg.Minio.SecretKey,
			"bucket name":      cfg.Minio.BucketName,
			"bucket path":      cfg.Minio.BucketPath,
		}); err != nil {
		var wrappedErr interface{ Unwrap() error }
		if errors.As(err, &wrappedErr) {
			unwrappedErr := errors.Unwrap(err)
			return unwrappedErr
		}
		return err
	}

	command := "mc alias"
	if cfg.Minio.Insecure {
		command = "mc alias --insecure"
	}
	lines := []string{
		fmt.Sprintf("%s set %s %s %s %s", command, cfg.Minio.Alias, cfg.Minio.URL, cfg.Minio.AccessKey, cfg.Minio.SecretKey),
	}

	var backupDate, backupTime, backupID string
	var err error

	if !dryRun {
		backupDate, backupTime, backupID, err = extractBackupDetails(e.LastResult().Output, dateTimeExtractor, idExtractor)
		if err != nil {
			return err
		}
		instance.BackupID = backupID
		instance.BackupDate = backupDate
		instance.BackupTime = backupTime
	}

	log.Debugf("backupDate: %s, backupTime: %s, backupID: %s", instance.BackupDate, instance.BackupTime, instance.BackupID)

	command = "mc cp -q"
	if cfg.Minio.Insecure {
		command = "mc cp -q --insecure"
	}
	backupDir := filepath.Join(cfg.BackupDirectory, instance.GUID, backupID)
	instance.BackupFile = filepath.Join(backupDir, "mysql-backup.tar.gpg")
	lines = append(lines, []string{
		fmt.Sprintf("mkdir -p %s", backupDir),
		fmt.Sprintf("%s %s/%s/%s/service-instance_%s/%s/%s_%s.tar %s",
			command,
			cfg.Minio.Alias,
			cfg.Minio.BucketName,
			cfg.Minio.BucketPath,
			instance.GUID,
			backupDate,
			instance.GUID,
			backupID,
			cfg.BackupDirectory,
		),
		fmt.Sprintf("tar xvf %s/%s_%s.tar -C %s", cfg.BackupDirectory, instance.GUID, backupID, backupDir),
		fmt.Sprintf("mv %s/*.tar.gpg %s", backupDir, instance.BackupFile),
		fmt.Sprintf("rm -f %s/%s_%s.tar", cfg.BackupDirectory, instance.GUID, backupID),
	}...)

	_, err = e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		return errors.Wrap(err, "failed to download backup")
	}

	return nil
}

func backupIDExtractor(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("couldn't extract backup id, output is empty")
	}

	lines := strings.Split(s, "\n")
	if len(lines) < 3 {
		return "", fmt.Errorf("couldn't extract backup id, not enough lines in output")
	}

	fields := strings.Fields(lines[2])
	if len(fields) < 7 {
		return "", fmt.Errorf("couldn't extract backup id, not enough fields to parse in output")
	}

	parts := strings.Split(fields[0], "_")
	if len(parts) < 2 {
		return "", fmt.Errorf("couldn't extract backup id, no underscore in fields")
	}

	return parts[1], nil
}

func backupDateTimeExtractor(s string) (string, string, error) {
	if len(s) == 0 {
		return "", "", fmt.Errorf("couldn't extract datetime, output is empty")
	}

	lines := strings.Split(s, "\n")
	if len(lines) < 3 {
		return "", "", fmt.Errorf("couldn't extract datetime, not enough lines in output")
	}

	fields := strings.Fields(lines[2])
	if len(fields) < 7 {
		log.Errorf("couldn't extract datetime, fields are: %+v", fields)
		return "", "", fmt.Errorf("couldn't extract datetime, not enough fields to parse in output")
	}

	// Wed Nov 24 21:04:52 UTC 2021
	day, _ := strconv.Atoi(fields[3])
	backupDate := fmt.Sprintf("%s-%s-%s %s", fields[6], fields[2], fmt.Sprintf("%02d", day), fields[4])
	dt, err := time.Parse("2006-Jan-02 15:04:05", backupDate)
	if err != nil {
		return "", "", err
	}

	return dt.Format("2006/01/02"), dt.Format("15:04:05"), nil
}

func encryptionKeyExtractor(input string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("couldn't extract encryption key, output is empty")
	}
	log.Debugln(input)
	scanner := bufio.NewScanner(strings.NewReader(input))
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		lines := strings.Split(string(data), "\n")
		if atEOF {
			return 0, nil, io.EOF
		}
		for _, l := range lines {
			fields := strings.Split(l, "| ")
			return len(l) + 1, []byte(fields[1]), nil
		}
		return 0, nil, nil
	}
	scanner.Split(split)

	fields := make([]string, 0, 6)
	for scanner.Scan() {
		if scanner.Text() != "" {
			fields = append(fields, scanner.Text())
		}
	}
	if scanner.Err() != nil {
		return "", scanner.Err()
	}

	return strings.TrimSuffix(fields[len(fields)-2], "\r"), nil
}

func extractBackupDetails(s string, dateTimeExtractor BackupDateTimeExtractor, idExtractor BackupIDExtractor) (string, string, string, error) {
	var backupDate, backupTime, backupID string
	var err error
	backupDate, backupTime, err = dateTimeExtractor(s)
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to extract the date and time from backup")
	}

	backupID, err = idExtractor(s)
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to extract the id from backup")
	}

	return backupDate, backupTime, backupID, nil
}

func checkRequiredParams(msg string, params map[string]string) error {
	var wrappedErr error
	for name, value := range params {
		if value == "" {
			if wrappedErr != nil {
				wrappedErr = fmt.Errorf("%s: %w", fmt.Sprintf(msg, name), wrappedErr)
			} else {
				wrappedErr = fmt.Errorf(msg, name)
			}
		}
	}

	return wrappedErr
}

func executeForDuration(ctx context.Context, s exec.Executor, lines []string, timeout time.Duration, pause time.Duration, doneCondition func(exec.Result) bool, errorCondition func(exec.Result) (bool, error)) (exec.Result, error) {
	done := make(chan bool)
	var err error
	var res exec.Result
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			res, err = s.Execute(child, strings.NewReader(strings.Join(lines, "\n")))
			if err != nil {
				cancel()
				return
			}

			if ok, condErr := errorCondition(res); ok {
				log.Errorln("failed to backup service instance")
				err = condErr
				cancel()
				return
			}

			if doneCondition(res) || res.DryRun {
				close(done)
				continue
			}

			time.Sleep(pause)
		}
	}()
	select {
	case <-done:
		return res, nil
	case <-child.Done():
		return res, err
	case <-time.After(timeout):
		close(done)
		return res, errors.New("timed out waiting for backup")
	}
}

func renameBackupFile(fso sio.FileSystemOperations, backupDir, backupFilename string) error {
	reStr := ".*.gpg"
	re := regexp.MustCompile(reStr)
	err := fso.Rename(re, backupDir, backupFilename)
	if err != nil {
		log.Errorf("Failed to rename any files matching %q in %q to %q, %v", reStr, backupDir, backupFilename, err)
		return fmt.Errorf("failed to rename any files matching %q in %q to %q: %w", reStr, backupDir, backupFilename, err)
	}

	return nil
}
