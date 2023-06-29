## Pessimism Telemetry

Pessimism uses [Prometheus](https://prometheus.io/docs/introduction/overview/) for telemetry. The application spins up a metrics server on a specified port (default 7300) and exposes metrics at the `/metrics` endpoint. 

### Metrics Configuration
Configuration depends on the client that you are running, but the general idea is that you need to configure a Prometheus client to scrape the metrics server to view metrics in a dashboard such as Grafana or Datadog. Here at Coinbase, we use Datadog for our metrics dashboards, so I will run through an example using Datadog.

#### Datadog Agent (Macos)
First, ensure that you have installed the datadog agent on the machine that is running the metrics server. You can find instructions on how to do that [here](https://docs.datadoghq.com/agent/basic_agent_usage/osx/?tab=agentv6v7).
Next, you will need to configure the datadog agent to scrape the metrics server. You can do this by adding the following to your datadog agent configuration file (usually located at `~/.datadog-agent/datadog.yaml`):
```yaml 
api_key: <YOUR_API_KEY>
site: datadoghq.com
tags: 
    - host:<YOUR_HOSTNAME>
```

Next you'll need to update the datadog prometheus configuration file at `~/.datadog-agent/conf.d/prometheus.d/conf.yaml` to include the following:
```yaml
instances:
    - prometheus_url: http://localhost:7300/metrics
      namespace: "pessimism"
      metrics:
        - *
```

Restart the agent via `sudo launchctl stop com.datadoghq.agent` and `sudo launchctl start com.datadoghq.agent`.

After a few minutes, you should be able to see metrics in Datadog by searching for `pessimism` in the metrics explorer, filtering by your hostname set above in step 1.

#### Datadog Agent (Docker)
First, make sure you have docker installed locally. You can find instructions on how to do that [here](https://docs.docker.com/get-docker/).
Next, pull the datadog agent image from dockerhub:  :q

Next, run the datadog agent container with the following command:
```bash
docker run -d --cgroupns host \
    --pid host \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    -v /proc/:/host/proc/:ro \
    -v /sys/fs/cgroup/:/host/sys/fs/cgroup:ro \
    -e DD_API_KEY=${DD_API_KEY} \
    -e DD_SITE="datadoghq.com" \
    -e PROMETHEUS_PORT=7300 \
    -e PROMETHEUS_ENDPOINT="0.0.0.0:7300/metrics" \
    -e NAMESPACE="pessimism_local" \
datadog/agent:latest
```

You should start seeing metrics in Datadog after a few minutes.

## Documentation
To generate documentation for the metrics, run `make docs` from the root of the repository. This will generate markdown 
which can be pasted directly below to keep current metric documentation up to date.

## Current Metrics
|                  METRIC                   |                      DESCRIPTION                       |  LABELS   |  TYPE   |
|-------------------------------------------|--------------------------------------------------------|-----------|---------|
| pessimism_up                              | 1 if the service is up                                 |           | gauge   |
| pessimism_invariants_active_invariants    | Number of active invariants                            |           | gauge   |
| pessimism_etl_active_pipelines            | Number of active pipelines                             |           | gauge   |
| pessimism_invariants_invariant_runs_total | Number of times a specific invariant has been run      | invariant | counter |
| pessimism_alarms_generated_total          | Number of total alarms generated for a given invariant | invariant | counter |
| pessimism_node_errors_total               | Number of node errors caught                           | node      | counter |