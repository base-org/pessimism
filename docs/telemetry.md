## Pessimism Telemetry

Pessimism uses [Prometheus](https://prometheus.io/docs/introduction/overview/) for telemetry. The application spins up a metrics server on a specified port (default 7300) and exposes the `/metrics` endpoint. 

### Local Testing
To verify that metrics are being collected locally, curl the metrics endpoint via `curl localhost:7300/metrics`. The response should display all custom and system metrics.

### Server Configuration
The default configuration within `config.env.template` should be suitable in most cases, however if you do not want to run the metrics server, set `METRICS_ENABLED=0` and the metrics server will not be started. This is useful mainly for testing purposes. 

## Generating Documentation
To generate documentation for metrics, run `make docs` from the root of the repository. This will generate markdown 
which can be pasted directly below to keep current system metric documentation up to date.

## Current Metrics
|                  METRIC                   |                      DESCRIPTION                       |  LABELS   |  TYPE   |
|-------------------------------------------|--------------------------------------------------------|-----------|---------|
| pessimism_up                              | 1 if the service is up                                 |           | gauge   |
| pessimism_invariants_active_invariants    | Number of active invariants                            |           | gauge   |
| pessimism_etl_active_pipelines            | Number of active pipelines                             |           | gauge   |
| pessimism_invariants_invariant_runs_total | Number of times a specific invariant has been run      | invariant | counter |
| pessimism_alarms_generated_total          | Number of total alarms generated for a given invariant | invariant | counter |
| pessimism_node_errors_total               | Number of node errors caught                           | node      | counter |