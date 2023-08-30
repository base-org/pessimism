---
layout: page
title: Alerting
permalink: /architecture/alerting
---

{% raw %}
<script src="https://cdn.jsdelivr.net/npm/mermaid@10.3.0/dist/mermaid.min.js"></script>
{% endraw %}

## Overview
The alerting subsystem will receive alerts from the `EngineManager` and publish them to the appropriate alerting destinations. The alerting subsystem will also be responsible for managing the lifecycle of alerts. This includes creating, updating, and removing alerting entries for heuristic sessions.

## Diagram

{% raw %}
<div class="mermaid">
graph LR

subgraph EM["Engine Manager"]
    alertingRelay
end

subgraph AM["Alerting Manager"]
    alertingRelay --> |Alert|EL
    EL[eventLoop] --> |Alert SUUID|AS["AlertStore"]
    AS --> |Alert Policy|EL
    EL --> |Submit alert|SR["SeverityRouter"]
    SR --> SH["Slack"]
    SR --> PH["PagerDuty"]
    SR --> CPH["CounterParty Handler"]

end
CPH --> |"HTTP POST"|TPH["Third Party API"]
SH --> |"HTTP POST"|SlackAPI("Slack Webhook API")
PH --> |"HTTP POST"|PagerDutyAPI("PagerDuty API")

</div>
{% endraw %}

### Alert
An `Alert` type stores all necessary metadata for external consumption by a downstream entity. 
### Alert Store
The alert store is a persistent storage layer that is used to store alerting entries. As of now, the alert store only supports configurable alert destinations for each alerting entry. Ie:
```
    (SUUID) --> (AlertDestination)
```

### Alert Destinations
An alert destination is a configurable destination that an alert can be sent to. As of now this only includes _Slack_. In the future however, this will include other third party integrations.

#### Slack
The Slack alert destination is a configurable destination that allows alerts to be sent to a specific Slack channel. The Slack alert destination will be configured with a Slack webhook URL. The Slack alert destination will then use this URL to send alerts to the specified Slack channel.


#### PagerDuty
The PagerDuty alert destination is a configurable destination that allows alerts to be sent to a specific PagerDuty services via the use of integration keys. Pessimism also uses the SUUID associated with an alert as a deduplication key for PagerDuty. This is done to ensure that PagerDuty will not be spammed with duplicate or incidents. 


### Alert CoolDowns
To ensure that alerts aren't spammed to destinations once invoked, a time based cooldown value (`cooldown_time`) can be defined within the  `alert_params` of a heuristic session config. This time value determines how long a heuristic session must wait before being allowed to alert again. 

An example of this is shown below:
```json
    {
      "network": "layer1",
      "pipeline_type": "live",
      "type": "balance_enforcement",
      "start_height": null,
      "alerting_params": {
        "cooldown_time": 10,
        "message": "",
        "destination": "slack"
      },
      "heuristic_params": {
        "address": "0xfC0157aA4F5DB7177830ACddB3D5a9BB5BE9cc5e",
        "lower": 1,
        "upper": 2
      }
    }
```

### Alert Messages
Pessimism allows for the arbitrary customization of alert messages. This is done by defining an `message` value string within the `alerting_params` of a heuristic session bootstrap config or session creation request. This is critical for providing additional context on alerts that allow for easier ingestion by downstream consumers (i.e, alert responders). 