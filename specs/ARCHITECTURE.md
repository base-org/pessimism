# Pessimism Architecture

## Overview
There are *three subsystems* that drive Pessimism’s architecture:
1. [ETL](./ETL.md) - Modularized data extraction system for retrieving and processing external chain data in the form of a DAG known as the Pipeline DAG
2. [Risk Engine](./RISK_ENGINE.md) - Logical execution platform that runs a set of invariants on the data funneled from the Pipeline DAG
3. [Alerting](./ALERTING.md) - Alerting system that is used to notify users of invariant failures

These systems will be accessible by a client through the use of a restful HTTP API that has direct access to both primary subsystems.

The API will be supported to allow Pessimism developers to:
1. Start invariant sessions
2. Update existing invariant sessions
3. Remove invariant sessions

## Diagram
![high level component diagram](./assets/high_level_diagram.png)

