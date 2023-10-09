---
layout: page
title: Alert Routing
permalink: /alert-routing
---


## Overview

The alert routing feature enables users to define a number of alert destinations
and then route alerts to those destinations based on the alert's severity.
For example, a user may want to send all alerts to Slack but only send high
severity alerts to PagerDuty.

## How it works

Alerts are routed to destinations based on the severity of the given heuristic.
When a heuristic is deployed, the user must specify the severity of the alert
that the heuristic will produce. When the heuristic is run, the alert is routed
to the configured destinations based on the severity of the alert. For example,
if a heuristic is configured to produce a high severity alert, the alert will be
routed to all configured destinations that support high severity alerts.

Each severity level is configured independently for each alert destination.
A user can add any number of alert configurations per severity.

Located in the root directory you'll find a file named `alerts-template.yaml`.
This file contains a template for configuring alert routing. The template contains
a few examples on how you might want to configure your alert routing.

## Supported Alert Destinations

Pessimism currently supports the following alert destinations:

| Name      | Description                         |
|-----------|-------------------------------------|
| slack     | Sends alerts to a Slack channel     |
| pagerduty | Sends alerts to a PagerDuty service |

## Alert Severity

Pessimism currently defines the following severities for alerts:

| Severity | Description                                                                 |
|----------|-----------------------------------------------------------------------------|
| low      | Alerts that may not require immediate attention                             |
| medium   | Alerts that could be hazardous, but may not be completely destructive       |
| high     | Alerts that require immediate attention and could result in a loss of funds |

## PagerDuty Severity Mapping

PagerDuty supports the following severities: `critical`, `error`, `warning`,
and `info`. Pessimism maps the Pessimism severities to
[PagerDuty severities](https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTgx-send-an-alert-event)
as follows ([ref](../internal/core/alert.go)):

| Pessimism | PagerDuty |
|-----------|-----------|
| low       | warning   |
| medium    | error     |
| high      | critical  |
