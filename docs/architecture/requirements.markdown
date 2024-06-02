---
layout: page
title: Requirements
permalink: /requirements
---

{% raw %}
<script src="https://cdn.jsdelivr.net/npm/mermaid@10.3.0/dist/mermaid.min.js"></script>
{% endraw %}

## Overview
The following service dependencies are required for running a pessimism instance:

### L1 Node
An ethereum json rpc node (e.g, go-ethereum, reth) is required for reading and syncing from recent chain state. It is best recommended to use an archival node if you're running any heuristic sessions that requires more than the last 128 blocks. Otherwise full node syncing should be sufficient assuming you're starting the application at most recent block state.

### L2 Node
An OP Stack execution node (i.e, op-reth, op-geth) is required for reading and syncing from recent chain state. It is best recommended to use an archival node if you're running any heuristic sessions that requires more than the last 128 blocks. Otherwise full node syncing should be sufficient assuming you're starting the application at most recent block state. 

### OP Indexer
The [OP Indexer](https://github.com/ethereum-optimism/optimism/tree/develop/indexer) is used as a stateful dependency for ingesting key native bridge metadata (i.e, L2 -> L1 withdrawals, supply counts).  This is only required if you'd like native bridge heuristic support (`withdrawal_enforcement`, `supply_enforcement`).
