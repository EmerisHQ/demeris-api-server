# API server

[![codecov](https://codecov.io/gh/emerishq/demeris-api-server/branch/main/graph/badge.svg?token=8P0G6VQDPE)](https://codecov.io/gh/emerishq/demeris-api-server)
[![Build status](https://github.com//emerishq/demeris-api-server/workflows/Build/badge.svg?token=8P0G6VQDPE)](https://github.com//emerishq/demeris-api-server/commits/main)
[![Tests status](https://github.com//emerishq/demeris-api-server/workflows/Tests/badge.svg)](https://github.com//emerishq/demeris-api-server/commits/main)
[![Lint](https://github.com/emerishq/demeris-api-server/Lint/badge.svg?token=8P0G6VQDPE)](https://github.com//emerishq/demeris-api-server/commits/main)

REST API entry-point to the Emeris back-end.  
At its simplest the api-server is a translation layer between JSON REST and chain-specific RPC.

## Actions

* `make`  
  Build and generate a binary.

* `make generate-swagger`  
  Generate `swagger.yaml` under `docs`.
  Alternatively, you can get a generated copy as a [Github action artifact](https://github.com/emerishq/demeris-api-server/actions/workflows/swagger.yml).

* `make generate-mocks`  
  Generate mock testing files under the `mocs` package.

## Dependencies & Licenses

The list of non-{Cosmos, AiB, Tendermint} dependencies and their licenses are:

|Module   	                  |License          |
|---	                      |---  	        |
|gin-gonic/gin   	          |MIT   	        |
|go-playground/validator   	  |MIT   	        |
|jmoiron/sqlx   	          |MIT   	        |
|go.uber.org/zap   	          |MIT           	|
|sigs.k8s.io/controller-runtime |MIT            |
|sony/sonyflake               |MIT              |
|lib/pq                       |Open use         |
|patrickmn/go-cache        |MIT              |