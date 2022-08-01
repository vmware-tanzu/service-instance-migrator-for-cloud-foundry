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

## Run the integration tests

```shell
make test-e2e
```

Run `make help` for all other tasks.
