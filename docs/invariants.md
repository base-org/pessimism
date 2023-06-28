# Invariants


## Balance Enforcement
The hardcoded `balance_enforcement` invariant checks the native ETH balance of some address every `n` milliseconds and alerts to slack if the account's balance is ever less than `lower` or greater than `upper` value. This invariant is useful for monitoring hot wallets and other accounts that should always have a balance above a certain threshold.

### Parameters
| Name | Type | Description |
| ---- | ---- | ----------- |
| address | string | The address to check the balance of |
| lower | float | The ETH lower bound of the balance |
| upper | float | The ETH upper bound of the balance |

### Example Deploy Request
```
curl --location --request POST 'http://localhost:8080/v0/invariant' \
--header 'Content-Type: text/plain' \
--data-raw '{
  "method": "run",
  "params": {
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
}'
```

## Contract Event
The hardcoded `contract_event` invariant scans newly produced blocks for a specific contract event and alerts to slack if the event is found. This invariant is useful for monitoring for specific contract events that should never occur.

### Parameters
| Name | Type | Description |
| ---- | ---- | ----------- |
| address | string | The address of the contract to scan for the events |
| args | []string | The event signatures to scan for |

**NOTE:** The `args` field is an array of string event declarations (eg. `Transfer(address,address,uint256)`). Currently Pessimism makes no use of contract ABIs so the manually specified event declarations are not validated for correctness. If the event declaration is incorrect, the invariant session will never alert but will continue to scan. 


### Example Deploy Request
```
curl --location --request POST 'http://localhost:8080/v0/invariant' \
--header 'Content-Type: text/plain' \
--data-raw '{
  "method": "run",
  "params": {
    "network": "layer1",
    "pipeline_type": "live",
    "type": "contract_event", 
    "start_height": null,
    "alert_destination": "slack",
    "invariant_params": {
        "address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "args": ["Transfer(address,address,uint256)"]
   }
}
}'