# Dapla API reconcilers

This repository contains the reconcilers for the Dapla API.

We are deeply inspired by [Nais API reconcilers](https://github.com/nais/api-reconcilers)

The main purpose is to create team resources, permissions and maintain them.

## Local development

[dapla-api](../api/README.md) is a dependency for this project, and you will need to
have [mise](https://mise.jdx.dev/) installed
on your system.

To run the reconciler locally, you need to have the nais/api project cloned and running.
See the [dapla-api README local development](../api/README.md) for more information.

Given that a lot of the reconcilers are using external services, most of these requires authentication and access to
these services.
So ensure that you configure and provide a proper environment for the reconcilers to run.
You may use the example configuration file to skip the boring process of figuring it out:

```shell
mise install         # Install required dependencies
cp .env.example .env # Copy the example configuration file
```

To run the reconciler locally, you can use the following command:

```shell
mise run local
```

This will build the reconciler and run it locally.
It sets an environment variable to communicate with the nais/api project running locally.

Run `mise run test` to run the tests.

### Local kind cluster setup (only relevant if doing stuff against Kubernetes, e.g. the namespace reconciler)

TODO

## Architecture

The project contains a set of reconcilers which are run on schedule or triggered by events.
A manager is responsible for running the reconcilers and handling the errors.

The manager will listen for pubsub events and trigger the correct reconcilers when needed.

All state and data is stored in Dapla API, and the communication with the API is done through GRPC.
