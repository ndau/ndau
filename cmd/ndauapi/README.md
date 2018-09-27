# ndauapi

This tool provides an http interface to an ndau node.

# Design

This tool uses a [boneful](https://github.com/kentquirk/boneful) service, based on the [bone router](https://github.com/go-zoo/bone).

Configuration is provided with environment variables specifying the following

  * How much logging you want (error, warn, info, debug).
  * The protocol, host and port of the ndau node's rpc port. Required.
  * And the port to listen on.

Communication between this program and tendermint is firstly done with the tool pkg and indirectly with [Tendermint's RPC client](https://github.com/tendermint/tendermint/tree/master/rpc/client), which is based on JSON RPC.

Testing depends on a test net to be available and as such are not very pure unit tests. A TODO is to find a suitable mock. 

# Getting started

```shell
./build.sh
NDAUAPI_NODE_ADDRESS=http://127.0.0.1:31001 ./ndauapi
```

# Basic Usage

```shell
# get node status
curl localhost:3030/status
```

# Testing in VSCode

Please include this in your VSCode config to run individual tests. Replace the IP and port with your ndau node's IP and Tendermint RPC port.

```json
    "go.testEnvVars": {
        "NDAUAPI_NODE_ADDRESS": "http://127.0.0.1:31001"
    },
```

# API

The following is automatically generated. Please do not edit the README.md file. Any changes above this section should go in (README-template.md).

> TODO: Please note that this documentation implementation is incomplete and is missing complete responses.


---
# `/`

This service provides the API for Tendermint and Chaos/Order/ndau blockchain data



* [Status](#status)

* [Health](#health)

* [Net Info](#net info)

* [Genesis](#genesis)

* [ABCI Info](#abci info)

* [Num Unconfirmed Transactions](#num unconfirmed transactions)

* [Dump Consensus State](#dump consensus state)

* [Get Block](#get block)

* [Get Block Chain](#get block chain)

* [Node List](#node list)

* [Node List](#node list)

* [Address List](#address list)




---
## Status

### `GET /status`

_Returns the status of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "node_info": {
            "id": "",
            "listen_addr": "",
            "network": "",
            "version": "",
            "channels": "",
            "moniker": "",
            "other": null
          },
          "sync_info": {
            "latest_block_hash": "",
            "latest_app_hash": "",
            "latest_block_height": 0,
            "latest_block_time": "0001-01-01T00:00:00Z",
            "catching_up": false
          },
          "validator_info": {
            "address": "",
            "pub_key": null,
            "voting_power": 0
          }
        }
```



---
## Health

### `GET /health`

_Returns the health of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## Net Info

### `GET /net`

_Returns the network information of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "listening": false,
          "listeners": null,
          "n_peers": 0,
          "peers": null
        }
```



---
## Genesis

### `GET /genesis`

_Returns the genesis block of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "genesis": null
        }
```



---
## ABCI Info

### `GET /abci`

_Returns info on the ABCI interface._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "response": {}
        }
```



---
## Num Unconfirmed Transactions

### `GET /unconfirmed`

_Returns the number of unconfirmed transactions on the chain._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "node_info": {
            "id": "",
            "listen_addr": "",
            "network": "",
            "version": "",
            "channels": "",
            "moniker": "",
            "other": null
          },
          "sync_info": {
            "latest_block_hash": "",
            "latest_app_hash": "",
            "latest_block_height": 0,
            "latest_block_time": "0001-01-01T00:00:00Z",
            "catching_up": false
          },
          "validator_info": {
            "address": "",
            "pub_key": null,
            "voting_power": 0
          }
        }
```



---
## Dump Consensus State

### `GET /consensus`

_Returns the current Tendermint consensus state in JSON_








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "round_state": null,
          "peers": null
        }
```



---
## Get Block

### `GET /block`

_Returns the block in the chain at the given height._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 height | Query | Height of the block in chain to return. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "block_meta": null,
          "block": null
        }
```



---
## Get Block Chain

### `GET /blockchain`

_Returns a sequence of blocks starting at min_height and ending at max_height_




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 start | Query | Height at which to begin retrieval of blockchain sequence. | string
 end | Query | Height at which to end retrieval of blockchain sequence. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "last_height": 0,
          "block_metas": null
        }
```



---
## Node List

### `GET /nodes`

_Returns a list of all nodes._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "nodes": null
        }
```



---
## Node List

### `GET /nodes/:id`

_Returns a single node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "id": "",
          "listen_addr": "",
          "network": "",
          "version": "",
          "channels": "",
          "moniker": "",
          "other": null
        }
```



---
## Address List

### `POST /accounts`

_Returns a list of addresses._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```
