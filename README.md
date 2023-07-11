# pessimism
__Because you can't always be optimistic__

_Pessimism_ is a public good monitoring service that allows for [Op-Stack](https://stack.optimism.io/) and EVM compatible blockchains to be continuously assessed for real-time threats using custom defined user invariant rule sets. To learn about Pessimism's architecture, please advise the documentation. 

<!-- Badge row 1 - status -->

[![GitHub contributors](https://img.shields.io/github/contributors/base-org/pessimism)](https://github.com/base-org/pessimism/graphs/contributors)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/w/base-org/pessimism)](https://github.com/base-org/pessimism/graphs/contributors)
[![GitHub Stars](https://img.shields.io/github/stars/base-org/pessimism.svg)](https://github.com/base-org/pessimism/stargazers)
![GitHub repo size](https://img.shields.io/github/repo-size/base-org/pessimism)
[![GitHub](https://img.shields.io/github/license/base-org/pessimism?color=blue)](https://github.com/base-org/pessimism/blob/main/LICENSE)

<!-- Badge row 2 - detailed status -->

[![GitHub pull requests by-label](https://img.shields.io/github/issues-pr-raw/base-org/pessimism)](https://github.com/base-org/pessimism/pulls)
[![GitHub Issues](https://img.shields.io/github/issues-raw/base-org/pessimism.svg)](https://github.com/base-org/pessimism/issues)

**Warning:**
Pessimism is currently experimental and very much in development. It means Pessimism is currently unstable, so code will change and builds can break over the coming months. If you come across problems, it would help greatly to open issues so that we can fix them as quickly as possible.

## Setup
To use the template, run the following the command(s):
1. Create local config file (`config.env`) to store all necessary environmental variables. There's already an example `config.env.template` in the repo that stores default env vars.

2. [Download](https://go.dev/doc/install) or upgrade to `golang 1.19`.

3. Install all project golang dependencies by running `go mod download`.

# To Run
1. Compile pessimism to machine binary by running the following project level command(s):
    * Using Make: `make build-app`

2. To run the compiled binary, you can use the following project level command(s):
    * Using Make: `make run-app`
    * Direct Call: `./bin/pessimism`

## Linting
[golangci-lint](https://golangci-lint.run/) is used to perform code linting. Configurations are defined in [.golangci.yml](./.golangci.yml)
It can be ran using the following project level command(s):
* Using Make: `make lint`
* Direct Call: `golangci-lint run`

## Testing

### Unit Tests
Unit tests are written using the native [go test](https://pkg.go.dev/testing) library with test mocks generated using the golang native [mock](https://github.com/golang/mock) library. These tests live throughout the the project's `/internal` directory and are named with the suffix `_test.go`.

Unit tests can run using the following project level command(s):
* Using Make: `make test`
* Direct Call: `go test ./...`

### Integration Tests
Integration tests are written that leverage the existing [op-e2e](https://github.com/ethereum-optimism/optimism/tree/develop/op-e2e) testing framwork for spinning up pieces of the bedrock system. Additionally, the [httptest](https://pkg.go.dev/net/http/httptest) library is used to mock downstream alerting services (e.g. Slack's webhook API). These tests live in the project's `/e2e` directory.

Integration tests can run using the following project level command(s):
* Using Make: `make e2e-test`
* Direct Call: `go test ./e2e/...`

## Bootstrap Config
A bootstrap config file is used to define the initial state of the pessimism service. The file must be `json` formatted with its directive defined in the `BOOTSTRAP_PATH` env var. (e.g. `BOOTSTRAP_PATH=./genesis.json`)

### Example File
```
[
    {
          "network": "layer1",
          "pipeline_type": "live",
          "type": "contract_event", 
          "start_height": null,
          "alert_destination": "slack",
          "invariant_params": {
              "address": "0xfC0157aA4F5DB7177830ACddB3D5a9BB5BE9cc5e",
              "args": ["Transfer(address, address, uint256)"]
        }
    },
    {
        "network": "layer1",
        "pipeline_type": "live",
        "type": "balance_enforcement", 
        "start_height": null,
        "alert_destination": "slack",
        "invariant_params": {
            "address": "0xfC0157aA4F5DB7177830ACddB3D5a9BB5BE9cc5e",
            "lower": 1,
            "upper": 2
       }
    }
]
```



## Spawning an invariant session
To learn about the currently supported invariants and how to spawn them, please advise the [invariants' documentation](./docs/invariants.md).