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

This service provides the API for Tendermint and Chaos/Order/ndau blockchain data.

It is organized into several sections:

* /account returns data about specific accounts
* /block returns information about blocks on the blockchain
* /chaos returns information from the chaos chain
* /node provides information about node operations
* /order returns information from the order chain
* /transaction allows querying individual transactions on the blockchain
* /tx provides tools to build and submit transactions

Each of these, in turn, has several endpoints within it.




* [AccountByID](#accountbyid)

* [AccountsFromList](#accountsfromlist)

* [AccountEAIRate](#accounteairate)

* [AccountByID](#accountbyid)

* [BlockHash](#blockhash)

* [BlockHeight](#blockheight)

* [BlockRange](#blockrange)

* [ChaosSystemNames](#chaossystemnames)

* [ChaosSystemKey](#chaossystemkey)

* [ChaosHistoryKey](#chaoshistorykey)

* [NodeStatus](#nodestatus)

* [NodeHealth](#nodehealth)

* [NodeNetInfo](#nodenetinfo)

* [NodeGenesis](#nodegenesis)

* [NodeABCIInfo](#nodeabciinfo)

* [NodeNumUnconfirmedTransactions](#nodenumunconfirmedtransactions)

* [DumpConsensusState](#dumpconsensusstate)

* [NodeList](#nodelist)

* [NodeID](#nodeid)

* [OrderHash](#orderhash)

* [OrderHeight](#orderheight)

* [OrderHistory](#orderhistory)

* [CurrentOrderData](#currentorderdata)

* [TransactionByHash](#transactionbyhash)

* [TxChangeValidation](#txchangevalidation)

* [TxChangeSettlement](#txchangesettlement)

* [TxClaimAccount](#txclaimaccount)

* [TxClaimNodeRewards](#txclaimnoderewards)

* [TxCreditEAI](#txcrediteai)

* [TxDelegate](#txdelegate)

* [TxLock](#txlock)

* [TxNominateNodeReward](#txnominatenodereward)

* [TxNotify](#txnotify)

* [TxRegisterNode](#txregisternode)

* [TxReleaseFromEndowment](#txreleasefromendowment)

* [TxSetRewardsDest](#txsetrewardsdest)

* [TxStake](#txstake)

* [TxTransfer](#txtransfer)

* [TxTransferAndLock](#txtransferandlock)

* [TxSubmit](#txsubmit)




---
## AccountByID

### `GET /account/account/:accountid`

_Returns current state of an account given its address._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "balance": 0,
          "transferKeys": null,
          "rewardsTarget": null,
          "incomingRewardsFrom": null,
          "delegationNode": null,
          "lock": null,
          "stake": null,
          "lastEAIUpdate": 0,
          "lastWAAUpdate": 0,
          "weightedAverageAge": 0,
          "Sequence": 0,
          "settlements": null,
          "settlementSettings": {
            "Period": 0,
            "ChangesAt": null,
            "Next": null
          },
          "validationScript": null,
          "address": ""
        }
```



---
## AccountsFromList

### `POST /account/accounts`

_Returns current state of several accounts given a list of addresses._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## AccountEAIRate

### `POST /account/eai/rate`

_Returns eai rates for a collection of account information._

Accepts an array of rate requests that includes an address
field; this field may be any string (the account information is not
checked). It returns an array of rate responses, which includes
the address passed so that responses may be correctly correlated
to the input.



_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | []routes.EAIRateRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        [
          {
            "address": "accountAddress",
            "weightedAverageAge": 7776000000000,
            "lock": {
              "noticePeriod": 15552000000000,
              "unlocksOn": null
            }
          }
        ]
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        [
          {
            "address": "accountAddress",
            "eairate": 6000000
          }
        ]
```



---
## AccountByID

### `GET /account/history/:accountid`

_Returns the balance history of an account given its address._

The history includes the timestamp, new balance, and transaction ID of each change to the account's balance.
The result is reverse sorted chronologically from the current time, and supports paging by time.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 limit | Query | Maximum number of transactions to return; default=10. | string
 before | Query | Timestamp (ISO-3339) to start looking backwards; default=now. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## BlockHash

### `GET /block/hash/:blockhash`

_Returns the block in the chain with the given hash._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 blockhash | Query | Hash of the block in chain to return. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "block_meta": null,
          "block": null
        }
```



---
## BlockHeight

### `GET /block/height/:height`

_Returns the block in the chain at the given height._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 height | Query | Height of the block in chain to return. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "block_meta": null,
          "block": null
        }
```



---
## BlockRange

### `GET /block/range/:first/:last`

_Returns a sequence of blocks starting at first and ending at last_




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 first | Path | Height at which to begin retrieval of blocks. | int
 last | Path | Height at which to end retrieval of blocks. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "last_height": 0,
          "block_metas": null
        }
```



---
## ChaosSystemNames

### `GET /chaos/system/names`

_Returns all current named system variables on the chaos chain._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## ChaosSystemKey

### `GET /chaos/system/:key`

_Returns the current value of a system variable from the chaos chain._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 key | Path | Name of the system variable. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        ""
```



---
## ChaosHistoryKey

### `GET /chaos/history/:key`

_Returns the history of changes to a value of a chaos chain system variable._

The history includes the timestamp, new value, and transaction ID of each change to the account's balance.
The result is reverse sorted chronologically from the current time, and supports paging by time.


_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 key | Path | Name of the system variable. | string
 limit | Query | Maximum number of values to return; default=10. | string
 before | Query | Timestamp (ISO-3339) to start looking backwards; default=now. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## NodeStatus

### `GET /node/status`

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
## NodeHealth

### `GET /node/health`

_Returns the health of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## NodeNetInfo

### `GET /node/net`

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
## NodeGenesis

### `GET /node/genesis`

_Returns the genesis document of the current node._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "genesis": null
        }
```



---
## NodeABCIInfo

### `GET /node/abci`

_Returns info on the node's ABCI interface._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "response": {}
        }
```



---
## NodeNumUnconfirmedTransactions

### `GET /node/unconfirmed`

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
## DumpConsensusState

### `GET /node/consensus`

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
## NodeList

### `GET /node/nodes`

_Returns a list of all nodes._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "nodes": null
        }
```



---
## NodeID

### `GET /node/:id`

_Returns a single node._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 id | Path | the NodeID as a hex string | string






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
## OrderHash

### `GET /order/hash/:ndauhash`

_Returns the collection of data from the order chain as of a specific ndau blockhash._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 ndauhash | Path | Hash from the ndau chain. | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "marketPrice": 0,
          "targetPrice": 0,
          "floorPrice": 0,
          "endowmentSold": 0,
          "totalNdau": 0,
          "USD": ""
        }
```



---
## OrderHeight

### `GET /order/height/:ndauheight`

_Returns the collection of data from the order chain as of a specific ndau block height._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 ndauheight | Path | Height from the ndau chain. | int






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "marketPrice": 0,
          "targetPrice": 0,
          "floorPrice": 0,
          "endowmentSold": 0,
          "totalNdau": 0,
          "USD": ""
        }
```



---
## OrderHistory

### `GET /order/history/`

_Returns an array of data from the order chain at periodic intervals over time, sorted chronologically._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 limit | Query | Maximum number of values to return; default=100, max=1000. | string
 period | Query | Duration between samples (ex: 1d, 5m); default=1d. | string
 before | Query | Timestamp (ISO-3339) to end (exclusive); default=now. | string
 after | Query | Timestamp (ISO-3339) to start (inclusive); default=before-(limit*period). | string






_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        []
```



---
## CurrentOrderData

### `GET /order/current`

_Returns current order chain data for key parameters._

Returns current order chain information for 5 parameters:
* Market price
* Target price
* Floor price
* Total ndau sold from the endowment
* Total ndau in circulation







_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {
          "marketPrice": 16.85,
          "targetPrice": 17,
          "floorPrice": 2.57,
          "endowmentSold": 291900000000000,
          "totalNdau": 314159300000000,
          "USD": "USD"
        }
```



---
## TransactionByHash

### `GET /transaction/:txhash`

_Returns a transaction given its tx hash._








_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxChangeValidation

### `POST /tx/changevalidation`

_Returns a prepared ChangeValidation transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxChangeValidationRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxChangeSettlement

### `POST /tx/changesettlement`

_Returns a prepared ChangeSettlement transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxChangeSettlementRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxClaimAccount

### `POST /tx/claimaccount`

_Returns a prepared ClaimAccount transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxClaimAccountRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxClaimNodeRewards

### `POST /tx/claimnoderewards`

_Returns a prepared ClaimNodeRewards transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxClaimNodeRewardsRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxCreditEAI

### `POST /tx/crediteai`

_Returns a prepared CreditEAI transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxCreditEAIRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxDelegate

### `POST /tx/delegate`

_Returns a prepared Delegate transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxDelegateRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxLock

### `POST /tx/lock`

_Returns a prepared Lock transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxLockRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxNominateNodeReward

### `POST /tx/nominatenodereward`

_Returns a prepared NominateNodeReward transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxNominateNodeRewardRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxNotify

### `POST /tx/notify`

_Returns a prepared Notify transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxNotifyRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxRegisterNode

### `POST /tx/registernode`

_Returns a prepared RegisterNode transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxRegisterNodeRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxReleaseFromEndowment

### `POST /tx/releasefromendowment`

_Returns a prepared ReleaseFromEndowment transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxReleaseFromEndowmentRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxSetRewardsDest

### `POST /tx/setrewardsdest`

_Returns a prepared SetRewardsDest transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxSetRewardsDestRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxStake

### `POST /tx/stake`

_Returns a prepared Stake transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxStakeRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxTransfer

### `POST /tx/transfer`

_Returns a prepared Transfer transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxTransferRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxTransferAndLock

### `POST /tx/transferandlock`

_Returns a prepared TransferAndLock	transaction for signature._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.TxTransferAndLockRequest




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```



---
## TxSubmit

### `POST /tx/submit`

_Submits a prepared transaction._




_**Parameters:**_

Name | Kind | Description | DataType
---- | ---- | ----------- | --------
 body | Body |  | routes.PreparedTx




_**Consumes:**_ `[application/json]`


_**Reads:**_
```json
        {}
```


_**Produces:**_ `[application/json]`


_**Writes:**_
```json
        {}
```
