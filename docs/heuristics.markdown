---
layout: page
title: Heuristics
permalink: /heuristics
---

# Heuristics

## Balance Enforcement

The hardcoded `balance_enforcement` heuristic checks the native ETH balance of some address every `n` milliseconds and alerts to slack if the account's balance is ever less than `lower` or greater than `upper` value. This heuristic is useful for monitoring hot wallets and other accounts that should always have a balance above a certain threshold.

### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| address | string | The address to check the balance of |
| lower | float | The ETH lower bound of the balance |
| upper | float | The ETH upper bound of the balance |

### Example Deploy Request

```bash
curl --location --request POST 'http://localhost:8080/v0/heuristic' \
--header 'Content-Type: text/plain' \
--data-raw '{
  "method": "run",
  "params": {
    "network": "layer1",
    "pipeline_type": "live",
    "type": "balance_enforcement", 
    "start_height": null,
    "alert_destination": "slack",
    "heuristic_params": {
        "address": "0xfC0157aA4F5DB7177830ACddB3D5a9BB5BE9cc5e",
        "lower": 1,
        "upper": 2
   }
}
}'
```

## Contract Event

The hardcoded `contract_event` heuristic scans newly produced blocks for a specific contract event and alerts to slack if the event is found. This heuristic is useful for monitoring for specific contract events that should never occur.

### Parameters

| Name    | Type     | Description                                        |
|---------|----------|----------------------------------------------------|
| address | string   | The address of the contract to scan for the events |
| args    | []string | The event signatures to scan for                   |

**NOTE:** The `args` field is an array of string event declarations (eg. `Transfer(address,address,uint256)`). Currently Pessimism makes no use of contract ABIs so the manually specified event declarations are not validated for correctness. If the event declaration is incorrect, the heuristic session will never alert but will continue to scan.

### Example Deploy Request

```
curl --location --request POST 'http://localhost:8080/v0/heuristic' \
--header 'Content-Type: text/plain' \
--data-raw '{
  "method": "run",
  "params": {
    "network": "layer1",
    "pipeline_type": "live",
    "type": "contract_event", 
    "start_height": null,
    "alert_destination": "slack",
    "heuristic_params": {
        "address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "args": ["Transfer(address,address,uint256)"]
   }
}
}'
```

## Withdrawal Safety

**NOTE:** This heuristic currently requires an active RPC connection to both L1/L2 networks as well as synced OP Indexer instance. Eventually withdrawal safety will be extended to run for `WithdrawalFinalized` and `MessagePassed` events. Using `MessagePassed` wouldn't require using an active OP Indexer instance since the event is emitted on the L2ToL1MessagePasser contract and doesn't have to be correlated from an L1 event.

The hardcoded `withdrawal_safety` heuristic runs a suite of security invariants upon detection of a withdraw event. The invariants ran are as follows:
1. The L1 withdrawal proven hash is present in the L2ToL1MessagePasser contract's internal state
2. The withdraw amount isn't greater than the `OptimismPortal` contract's balance
3. The withdraw amount isn't within `x%` of the `OptimismPortal` contract's balance
4. The withdraw message hash is a valid L2 message hash


### Parameters

| Name              | Type   | Description                                     |
|-------------------|--------|-------------------------------------------------|
| l1_portal_address | string | The address of the L1Portal contract            |
| l2_to_l1_address  | string | The address of the L2ToL1MessagePasser contract |

### Example Deploy Request

```bash
curl --location --request POST 'http://localhost:8080/v0/heuristic' \
--header 'Content-Type: text/plain' \
--data-raw '{
  "method": "run",
 "params": {
  "network": "layer1",
  "pipeline_type": "live",
  "type": "withdrawal_safety",
  "start_height":  null,
  "alert_destination": "slack",
    "heuristic_params": {
      "l1_portal_address": "0x111",
      "l2_to_l1_address":  "0x333",
    },
}'
```

## Fault Detection

**NOTE:** This heuristic requires an active RPC connection to both L1 and L2 networks. Furthermore, the Pessimism implementation of fault-detector assumes that a submitted L2 output on L1 will correspond to a canonical block on L2.

The hardcoded `fault_detector` heuristic scans for active `OutputProposed` events on an L1 Output Reader contract. Once an event is detected, the heuristic implementation proceeds to reconstruct a local state output for the corresponding L2 block. If there is a mismatch between the L1 output and the local state output, the heuristic alerts.

### Parameters

| Name              | Type   | Description                                     |
|-------------------|--------|-------------------------------------------------|
| l2_output_address | string | The address of the L1 output oracle             |
| l2_to_l1_address  | string | The address of the L2ToL1MessagePasser contract |

### Example Deploy Request

```bash
curl --location --request POST 'http://localhost:8080/v0/heuristic' \
--header 'Content-Type: text/plain' \
--data-raw '{
  "method": "run",
 "params": {
  "network": "layer1",
  "pipeline_type": "live",
  "type": "fault_detector",
  "start_height":  null,
  "alert_destination": "slack",
    "heuristic_params": {
      "l2_output_address":  "0x111",
      "l2_to_l1_address":   "0x333",
    },
}'
```
