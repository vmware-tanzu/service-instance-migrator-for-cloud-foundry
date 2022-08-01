# Development

This doc explains how to set up a development environment in order to start
[contributing](CONTRIBUTING.md) to the Service Instance Migrator for Cloud Foundry.

This project uses Go 1.17+ and Go modules. Clone the repo to any directory.

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
