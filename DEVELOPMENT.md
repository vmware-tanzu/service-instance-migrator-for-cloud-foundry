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
[example](test/e2e/si-migrator.yml.example) and refer to the [reference](README.md#config-reference) documentation.

Run the tests

```shell
make test-e2e
```

Run `make help` for all other tasks.
