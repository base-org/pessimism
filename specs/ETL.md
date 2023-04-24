# ETL

The Pessimism ETL is a generalized abstraction for a DAG based component system that continuously reads live chain data into inputs designed for consumption by the Risk Engine in the form of intertwined “pipelines”.


## Component Metadata
### Egress Handler
All component types use an `egressHandler` type for routing transit data to actively subscribed downstream components. 

### Ingress Handler
Some component types ()


### UUID
All components have a UUID that stores critical identification data. Component IDs are used by higher order abstractions to:
* Represent a component DAG 
* Understand when component duplicates occur in the system

### Pipe
Pipes are used to perform local arbitrary computations on some input data. Once input data processing has been completed, the output data is then submitted to its respective destination(s) using an egressHandler.

#### Attributes
* A communication channel with a pipeline manager
* Ingress handler that other components can write to
* `TransformFunc` - A processing function that performs some data translation/transformation on respective inputs
* An egressHandler that stores dependencies to write to (i,e. Other pipeline components, invariant engine)
* A specified output data type

### Oracle 
Oracles are responsible for collecting data from some external third party _(e.g, L1 geth node, L2 rollup node, etc.)_. 

#### Attributes
* A communication channel with the pipeline manager
* Poller/subscription logic that performs real-time data reads on some third-party source
* An egressHandler that stores dependencies to write to (i,e. Other pipeline components, invariant engine)
* A specified output data type

* (Optional) Interface with some storage (postgres, mongo, etc.) to persist lively extracted data
* (Optional) Backtest support for polling some data between some starting and ending block heights

### (TBD) Aggregator
Aggregators are used to solve the problem where a pipe or an invariant input will require multiple data points to perform an execution sequence. Since aggregators are subscribing to more than one data stream with different output frequencies, they must employ a synchronization policy for collecting and propagating multi-data inputs within a highly asynchronous environment.