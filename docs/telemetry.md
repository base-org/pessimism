## Pessimism Telemetry


// TODO: Update to use prometheus instead of datadog statsd
### Overview

In order to better understand the internal state of Pessimism, we made available metrics via the [DogStatsd](https://docs.datadoghq.com/developers/dogstatsd/?tab=hostagent) package. DogStatsd is a metrics aggregation service bundled with the Datadog Agent, and implements the StatsD protocol as well as a few Datadog specific extensions.

### Configuration / Setup

To enable metrics, you must first have the Datadog Agent installed on your machine. You can find instructions on how to install the Datadog Agent [here](https://docs.datadoghq.com/agent/).
Follow the instructions for your specific operating system, then you should be able to query the metrics via your telemetry client of choice.
At Coinbase, we use Datadog, so querying the metrics is as simple as going to the [Datadog Metrics Explorer](https://app.datadoghq.com/metric/explorer) and searching for the metric you want to query.

## Metric Naming Conventions
Within the internal/metrics package, we have defined.

## Available Metrics
Currently, the available metrics are as follows:

| Metric Name | Description |
| ----------- | ----------- |

