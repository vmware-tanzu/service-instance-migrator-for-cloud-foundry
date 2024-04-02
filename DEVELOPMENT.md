# Development

This directory contains documentation for Developers in order to start
[contributing](CONTRIBUTING.md) to the Service Instance Migrator for Cloud Foundry.

This project uses Go 1.17+ and Go modules. Clone the repo to any directory.

## Build and run all checks

Before submitting any PRs to upstream, make sure to run:

```shell
make all
```

## Build the project

To build into the local ./bin dir:

```shell
make build
```

## Run the unit tests

```shell
make test
```

## Run integration tests

The integration tests target just one foundation. Two orgs are created to simulate the effect of having multiple
foundations in order to save costs.

### Run the end-to-end test suite

The e2e tests can be executed to test the following migrators:

- [credhub](https://network.pivotal.io/products/credhub-service-broker/)
- [sqlserver](https://github.com/cloudfoundry-attic/mssql-server-broker/)
- [mysql](https://network.pivotal.io/products/pivotal-mysql/)
- [ecs](https://network.pivotal.io/products/ecs-service-broker/)

Create a `si-migrator.yml` file under [test/e2e](test/e2e) directory. You can copy our
[example](test/e2e/si-migrator.yml.example) and refer to the [reference](https://github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/blob/main/README.md#config-reference) documentation.

Run the tests

```shell
make test-e2e
```

If you only want to test the migration of a single service, then include the migration config for just that service. You also do not need to migrate from one
foundation to another. You can specify the same `url` for the Ops Manager in both `source` and `target` foundations of the config and that will export from one 
space to another.  Here's an example:

```sh
cat ./test/e2e/si-migrator.yml <<EOF
debug: true
exclude_orgs:
  - system
  - p-spring-cloud-services
  - p-dataflow
foundations:
  source:
    url: https://opsman.east.acme.com
    username: admin
    password: 'XPrQWi9xvwta$Ng'
    hostname: opsman.east.acme.com
    private_key: ~/.ssh/acme-east-prod
    ssh_user: "ubuntu"
  target:
    url: https://opsman.east.acme.com
    username: admin
    password: 'XPrQWi9xvwta$Ng'
    hostname: opsman.east.acme.com
    private_key: ~/.ssh/acme-east-prod
    ssh_user: "ubuntu"
migration:
  use_default_migrator: true
  migrators:
   - name: mysql
     migrator:
       backup_type: scp
       scp:
         username: ubuntu
         hostname: opsman.east.acme.com
         port: 22
         destination_directory: /home/ubuntu
         private_key: ~/.ssh/acme-east-prod
EOF
```

Then we can test exporting just mysql and the user-provided service instance bindings

```sh
make test-export-space
```

Run `make help` for all other tasks.
