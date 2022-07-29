# service-instance-migrator-for-cloud-foundry

[![build workflow](https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/actions/workflows/build.yml)

## Overview

The `service-instance-migrator` is a command-line tool for migrating [Service Instances](https://docs.cloudfoundry.org/devguide/services/) from one [Cloud Foundry](https://docs.cloudfoundry.org/) (CF) or [Tanzu Application Service](https://tanzu.vmware.com/application-service) (TAS) to another. The `service-instance-migrator` currently only supports TAS deployments because it relies on the Ops Manager API to execute some of its underlying commands. However, this dependency is mostly to ease the burden on providing all the configuration necessary for running a migration. And it's a near-term goal to remove Ops Manager as a required dependency.

## Prerequisites

The `service-instance-migrator` relies on the following tools to execute in a shell during the migration process. Please
ensure these are installed prior to running any `export` or `import` commands.

- [om](https://github.com/pivotal-cf/om)
- [bosh-cli](https://bosh.io/docs/cli-v2)
- [cf-cli](https://code.cloudfoundry.org/cli)
- [credhub-cli](https://github.com/cloudfoundry-incubator/credhub-cli)
- [jq](https://stedolan.github.io/jq)

## Supported Service Instance Types

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

## CC API Object to Migration Process Mapping

- Applications                      - App-Migrator
- Application Environment Variables - App-Migrator
- Buildpacks                        - Other
- Default Security Groups           - Other
- Feature Flags                     - Other
- Private Domains                   - CF-Mgmt
- Shared Domains                    - CF-Mgmt
- Routes                            - App-Migrator
- Route Mappings                    - App-Migrator
- Quota Definitions                 - CF-Mgmt
- Application Security Groups       - CF-Mgmt
- Services                          - Other
- Service Brokers                   - Other
- Service Plans                     - CF-Mgmt
- Service Plan Visibility           - CF-Mgmt
- Service Keys                      - Other
- Managed Service Instances         - Service-Instance-Migrator
- User Provided Services            - Service-Instance-Migrator
- Service Bindings                  - App-Migrator
- Orgs                              - CF-Mgmt
- Spaces                            - CF-Mgmt
- Space Quotas                      - CF-Mgmt
- Isolation Segments                - CF-Mgmt
- Stacks                            - Other
- Local UAA Users/Clients           - Other
- LDAP Users                        - CF-Mgmt
- Roles                             - CF-Mgmt

## Documentation

The `service-instance-migrator` requires user credentials or client credentials to communicate with the Cloud Foundry Cloud Controller API.

Create a `$HOME/.config/si-migrator/si-migrator.yml` using the following template:

```yaml
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

The locations of this file can be overridden using environment variables:

- `SI_MIGRATOR_CONFIG_FILE` will override cli config file location [default: `$HOME/si-migrator.yml`]
- `SI_MIGRATOR_CONFIG_HOME` will override cli config directory location [default: `.`, `$HOME`, or `$HOME/.config`]

The `source_api` and `target_api` and configuration under the `sqlserver` and `ecs` migrators will be looked up from Ops Manager,
so it's not required to set them. Command line flags always override any values found in the config file. Make sure to update the directories and domains as needed.

We will retrieve the encryption key and credentials for the cloud-controller database for the `ecs` and `sqlserver` migrations if
you do not specify these values, however, it does add extra time to the migration to retrieve them.

### Export

Running `export` without any flags will export all service instances of all supported types from the source foundation. This may take a long time depending on how many service instances you have in your source foundation.

```shell
service-instance-migrator export
```

### Import

Running `import` does the opposite of `export` and as you may have guessed, uses the output from export as it's input. This command will take all the service instances found in the export directory and attempt to import them into the target foundation.

```shell
service-instance-migrator import
```

Check out the [docs](./docs/si-migrator.md) to see usage for all the commands.

## Logs

By default, all log output is appended to `/tmp/si-migrator.log`. You can override this location by setting the
`SI_MIGRATOR_LOG_FILE` environment variable.

## For Developers

This project uses Go 1.16+ and Go modules. Clone the repo to any directory.

Build and run all checks

```shell
make all
```

Build the project

```shell
make install
```

Run the tests

```shell
make test
```

To create a release for `darwin`, `linux` and `windows`.

```shell
make release
```

Run `make help` for all other tasks.

## Contributing

The service-instance-migrator-for-cloud-foundry project team welcomes contributions from the community. Before you start working with service-instance-migrator-for-cloud-foundry, please
read our [Developer Certificate of Origin](https://cla.vmware.com/dco). All contributions to this repository must be
signed as described on that page. Your signature certifies that you wrote the patch or have the right to pass it on
as an open-source patch. For more detailed information, refer to [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Refer to [LICENSE](LICENSE) for details.
