## Pessimism API

### Overview
The Pessimism API is a RESTful HTTP API that allows users to interact with the Pessimism application. The API is built using the [go-chi](https://github.com/go-chi/chi) framework and is served using the native [http package](https://pkg.go.dev/net/http). The API is designed to be modular and extensible, allowing for the addition of new endpoints and functionality with relative ease.

Currently, interactive endpoint documentation is hosted via [Swagger UI](https://swagger.io/tools/swagger-ui/) at [https://base-org.github.io/pessimism/](https://base-org.github.io/pessimism/). 

### Configuration
The API can be customly configured using environment variables stored in a `config.env` file. The following environment variables are used to configure the API:
- `SERVER_HOST`: The host address to serve the API on (eg. `localhost`)
- `SERVER_PORT`: The port to serve the API on (eg. `8080`)
- `SERVER_KEEP_ALIVE`: The keep alive second duration for the server (eg. `10`)
- `SERVER_READ_TIMEOUT`: The read timeout second duration for the server (eg. `10`)
- `SERVER_WRITE_TIMEOUT`: The write timeout second duration for the server (eg. `10`)

### Authorization and Authentication
TBD