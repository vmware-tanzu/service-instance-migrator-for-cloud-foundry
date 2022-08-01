# Service Instance Migrator for Cloud Foundry

[![build workflow](https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/actions/workflows/build.yml)

## Overview

The `service-instance-migrator` is a command-line tool for migrating [Service Instances](https://docs.cloudfoundry.org/devguide/services/) from one [Cloud Foundry](https://docs.cloudfoundry.org/) (CF) or [Tanzu Application Service](https://tanzu.vmware.com/application-service) (TAS) to another. The `service-instance-migrator` currently only supports TAS deployments because it relies on the Ops Manager API to execute some of its underlying commands. However, this dependency is mostly to ease the burden on providing all the configuration necessary for running a migration. And it's a near-term goal to remove Ops Manager as a required dependency.

### Supported Service Instance Types

The following service types are currently implemented:

- SQL Server
- Tanzu MySQL
- Credhub Service Broker
- ECS service
- Scheduler

More to come in the future:

- RabbitMQ (on-demand & shared)
- Redis (on-demand & shared)
- Spring Cloud Gateway
- Spring Cloud Config Server
- Spring Cloud Registry
- SSO
- SMB
- AppD

## Getting Started

### Prerequisites

The `service-instance-migrator` relies on the following tools to execute in a shell during the migration process. Please
ensure these are installed prior to running any `export` or `import` commands.

- [om](https://github.com/pivotal-cf/om)
- [bosh-cli](https://bosh.io/docs/cli-v2)
- [cf-cli](https://code.cloudfoundry.org/cli)
- [credhub-cli](https://github.com/cloudfoundry-incubator/credhub-cli)
- [jq](https://stedolan.github.io/jq)

### Download latest release

Download the `service-instance-migrator-<OS>-amd64.tar.gz` from the most recent release listed on the [Service Instance Migrator for Cloud Foundry releases](https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/releases) page.

Following are the instructions for installing version `v0.0.8`.

#### For macOS

```shell
VERSION=v0.0.8
wget -q https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/releases/download/${VERSION}/service-instance-migrator-darwin-amd64.tgz
tar -xvf service-instance-migrator-darwin-amd64.tgz -C /usr/local/bin
chmod +x /usr/local/bin/service-instance-migrator
```

#### For linux

```shell
VERSION=v0.0.8
wget -q https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/releases/download/${VERSION}/service-instance-migrator-linux-amd64.tgz
tar -xvf service-instance-migrator-darwin-amd64.tgz -C /usr/local/bin
chmod +x /usr/local/bin/service-instance-migrator
```

### Build from source

See the [development guide](./DEVELOPMENT.md) for instructions to build from source.

## Documentation

The `service-instance-migrator` requires user credentials or client credentials to communicate with the Cloud Foundry Cloud Controller API.

The configuration for the CLI is specified in a file called [si-migrator.yml](si-migrator.yml.example) which can be overridden with the following environment variables.

- `SI_MIGRATOR_CONFIG_FILE` will override cli config file location [default: `./si-migrator.yml`]
- `SI_MIGRATOR_CONFIG_HOME` will override cli config directory location [default: `.`, `$HOME`, or `$HOME/.config/si-migrator`]

Create a copy of [si-migrator.yml](si-migrator.yml.example) and place in any of the locations above. You can leave out any migrators that do not apply to your CF deployments.

### Config Reference

```yaml
debug: false
export_dir: "/tmp/tas-export"
exclude_orgs: [system] # optional, you can also use include_orgs for migrating specific orgs
domains_to_replace:
  apps.src.tas.example.com: apps.dst.tas.example.com
ignore_service_keys: false # optional, don't create any service keys on import
source_bosh: # optional, will be fetched from opsman if not set
  url: https://10.1.0.1
  all_proxy: sssh+socks5://some-user@opsman-1.example.com:22?private-key=/path/to/ssh-key
  root_ca_cert: |
    a trusted cert
target_bosh: # optional, will be fetched from opsman if not set
  url: https://10.1.0.2
  all_proxy: sssh+socks5://some-user@opsman-2.example.com:22?private-key=/path/to/ssh-key
  root_ca_cert: |
    a trusted cert
source_api: # optional, will be fetched from opsman if not set
  url: https://api.src.tas.example.com
  # admin or client credentials (not both)
  username: ""
  password: ""
  client_id: client-with-cloudcontroller-admin-permissions
  client_secret: client-secret
target_api: # optional, will be fetched from opsman if not set
  url: https://api.dst.tas.example.com
  # admin or client credentials (not both)
  username: ""
  password: ""
  client_id: client-with-cloudcontroller-admin-permissions
  client_secret: client-secret
foundations:
  source:
    url: https://opsman-1.example.com
    client_id: ""
    client_secret: ""
    username: admin
    password: REDACTED
    hostname: opsman-1.example.com
    private_key: /Users/user/.ssh/opsman1
    ssh_user: ubuntu
  target:
    url: https://opsman-2.example.com
    client_id: ""
    client_secret: ""
    username: admin
    password: REDACTED
    hostname: opsman-2.example.com
    private_key: /Users/user/.ssh/opsman2
    ssh_user: ubuntu
migration:
  use_default_migrator: true # optional, use the default managed service migrator when no supported migrator exists
  migrators:
    - name: ecs
      migrator:
        source_ccdb: # optional, will be fetched from opsman if not set
          db_host: 192.168.2.21 # optional
          db_username: ccdb-username # optional
          db_password: ccdb-password # optional
          db_encryption_key: REDACTED # optional, used to decrypt credentials in the source ccdb
          ssh_host: 10.213.135.16 # optional, will use opsman host if not set
          ssh_username: jumpbox-user # optional, will use opsman user and ssh key if not set
          ssh_password: "" # optional
          ssh_private_key: /Users/user/.ssh/jumpbox # optional, will use opsman ssh key if not set
          ssh_tunnel: true # optional, will set to true and opsman will be used for the tunnel if not set
        target_ccdb: # optional, will be fetched from opsman if not set
          db_host: 192.168.4.21 # optional
          db_username: ccdb-username # optional
          db_password: ccdb-password # optional
          db_encryption_key: REDACTED # optional, used to decrypt credentials in the target ccdb
          ssh_host: 10.213.135.16 # optional, will use opsman host if not set
          ssh_username: jumpbox-user # optional, will use opsman user and ssh key if not set
          ssh_password: "" # optional
          ssh_private_key: /Users/user/.ssh/jumpbox # optional, will use opsman ssh key if not set
          ssh_tunnel: true # optional, will set to true and opsman will be used for the tunnel if not set
    - name: sqlserver
      migrator:
        source_ccdb: # optional, will be fetched from opsman if not set
          db_host: 192.168.2.21 # optional
          db_username: ccdb-username # optional
          db_password: ccdb-password # optional
          db_encryption_key: REDACTED # optional, used to decrypt credentials in the source ccdb
          ssh_host: 10.213.135.16 # optional, will use opsman host if not set
          ssh_username: jumpbox-user # optional, will use opsman user and ssh key if not set
          ssh_password: "" # optional
          ssh_private_key: /Users/user/.ssh/jumpbox # optional, will use opsman ssh key if not set
          ssh_tunnel: true # optional, will set to true and opsman will be used for the tunnel if not set
        target_ccdb: # optional, will be fetched from opsman if not set
          db_host: 192.168.4.21 # optional
          db_username: ccdb-username # optional
          db_password: ccdb-password # optional
          db_encryption_key: REDACTED # optional, used to decrypt credentials in the target ccdb
          ssh_host: 10.213.135.16 # optional, will use opsman host if not set
          ssh_username: jumpbox-user # optional, will use opsman user and ssh key if not set
          ssh_password: "" # optional
          ssh_private_key: /Users/user/.ssh/jumpbox # optional, will use opsman ssh key if not set
          ssh_tunnel: true # optional, will set to true and opsman will be used for the tunnel if not set
    - name: mysql
      migrator:
        backup_type: minio # one of [scp, s3, minio]
        backup_directory: /tmp # optional (defaults to export directory)
        scp:
          username: backuphost-username
          hostname: backuphost.example.com
          port: 22
          destination_directory: /path/to/backups/mysql-tas1
          private_key: /Users/user/.ssh/backup-host
        s3:
          endpoint: https://s3.us-east-1.amazonaws.com
          access_key_id: REDACTED
          secret_access_key: REDACTED
          region: us-east-1
          bucket_name: mysql-tas1
          bucket_path: p.mysql
          insecure: false
          force_path_style: true
        minio:
          alias: tas1ecstestdrive
          url: https://object.ecstestdrive.com
          access_key: REDACTED
          secret_key: REDACTED
          bucket_name: mysql-tas1
          bucket_path: p.mysql
          insecure: false
```

The `source_api` and `target_api` as well as `source_bosh` and `target_bosh` stanzas will be looked up from Ops Manager,
so it's not required to set them. Command line flags will always override any values found in the config file.

The `service-instance-migrator` retrieves the encryption key and credentials for the `cloud-controller` database used by
the `ecs` and `sqlserver` migrations if you do not specify these values. It does, however, add some extra time to the
migration to retrieve them.

### Commands

#### Export

Running `export` without any flags will export all service instances of all supported types from the source foundation.
This may take a long time depending on how many service instances you have in your source foundation.

```shell
service-instance-migrator export
```

#### Import

Running `import` does the opposite of `export` and as you may have guessed, uses the output from export as it's input.
This command will take all the service instances found in the export directory and attempt to import them into the target foundation.

```shell
service-instance-migrator import
```

Check out the [docs](./docs/si-migrator.md) to see usage for all the commands.

## Logs

By default, all log output is appended to `/tmp/si-migrator.log`. You can override this location by setting the
`SI_MIGRATOR_LOG_FILE` environment variable.

## Example Demos

To run the demos under [hack](hack), you need to set some environment variables. We suggest first installing
[direnv](https://direnv.net/) and creating a `.envrc` file under the root of the project.
Here's an example `.envrc` which exports the required environment variables:

```shell
export CF_TAS1=$HOME/.cf_tas1
export CF_TAS2=$HOME/.cf_tas2

export CF_SOURCE_HOME="$CF_TAS1"
export CF_TARGET_HOME="$CF_TAS2"
export CF_HOME="$CF_TAS1"

export CF_SOURCE_SYS_DOMAIN="sys.tas1.vmware.com"
export CF_SOURCE_APPS_DOMAIN="apps.tas1.vmware.com"
export CF_SOURCE_USERNAME="admin"
export CF_SOURCE_PASSWORD="your-cf-admin-password"
export CF_SOURCE_ORG=tas1
export CF_SOURCE_SPACE=si-migrator-test-space

export CF_TARGET_SYS_DOMAIN="sys.tas1.vmware.com"
export CF_TARGET_APPS_DOMAIN="apps.tas1.vmware.com"
export CF_TARGET_USERNAME="admin"
export CF_TARGET_PASSWORD="your-cf-admin-password"
export CF_TARGET_ORG=tas2
export CF_TARGET_SPACE=si-migrator-test-space
```

Install the [credhub service broker](https://network.pivotal.io/products/credhub-service-broker/). Then, run
`./hack/credhub-demo.sh` to demo migrating a credhub service instance.

## Contributing

The Service Instance Migrator for Cloud Foundry project team welcomes contributions from the community. Before you start working with Service Instance Migrator for Cloud Foundry, please
read our [Developer Certificate of Origin](https://cla.vmware.com/dco). All contributions to this repository must be
signed as described on that page. Your signature certifies that you wrote the patch or have the right to pass it on
as an open-source patch. For more detailed information, refer to [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Refer to [LICENSE](LICENSE) for details.
