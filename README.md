# Ndau

This is the implementation of the Ndau chain. See the Ndau design whitepaper for details.  All Ndau transactions are stored on this chain.  An (incomplete) list of possible transactions on the Ndau chain include:

- Transfer
- ChangeTransferKey
- ReleaseFromEndowment
- ChangeEscrowPeriod
- Delegate
- CreditEAI
- GTValidatorChange

#### See the [commands](https://github.com/oneiro-ndev/commands) README for how to build and run locally.  Some of the steps below may no longer applicable since we moved all `/cmd` source to the new repo.

## Install

- pre-requisites:

    1. Install [Go](https://golang.org/doc/install)
    1. Install [Docker](https://docs.docker.com/docker-for-mac/install/)
        - If you don't have an account, you can use one from [here](http://bugmenot.com/view/store.docker.com)
        - Run it from `/Applications`
    1. Clone this repo in `~/go/src/github.com/oneiro-ndev/` (required for Go)
    1. Download machine_user_key from your Oneiro 1password account, put it at the root of your cloned copy of this repo

- install [glide](https://github.com/Masterminds/glide):

    ```shell
    curl https://glide.sh/get | sh
    ```

    - Optionally, install glide using [Brew](https://brew.sh/)

- update dependencies

    ```shell
    glide install
    ```

- check build

    ```shell
    go build ./... && go test ./...
    ```

## Quick Start

There are a number of bash scripts in the `bin` directory which ease the pain of
getting a working system up and running. To get a node going from scratch:

```sh
oneiro-ndev/ndau $ bin/reset.sh && bin/build.sh && bin/init.sh && bin/run.sh
<output redacted>
```

That sequence of commands will remove any existing configuration data, build,
initialize, and start a new node. It will take several minutes and produce quite a
log of output; just bear with it.

Once that's going, you'll see a bunch of docker-compose messages about containers
running. If your environment does not contain a Honeycomb key, you'll then see a bunch
of log messages scrolling past. If it does, output will stop. Either way, your
terminal will still be blocked on the server. Start a new terminal.

```sh
oneiro-ndev/ndau $ glide install && bin/tool-build.sh
<output redacted>
```

This will build the ndau tool. You'll need to run `glide install` because unlike
the node, the tool is not containerized by default, so you need to have the
dependencies locally.

Once you've run the tool builder, the tool (`ndau`) will appear in the `oneiro-ndev/ndau`
directory. It has a fairly deep interface; just run it and any subcommand with
`-h|--help` to investigate.

## Building

It's possible to build and run the ndau node outside its containers without using
any of the scripts, but there are a few gotchas. One of these, as an example,
is that the build scripts inject the current version into both the node and
the tool. If you build either of these programs without the magic flags

```sh
-ldflags "-X github.com/oneiro-ndev/ndau/pkg/version.version=$VERSION"
```

, then the version strings will be unset when you run the applicable version query.

Of course, you also need to set the `$VERSION` environment variable. That's
handled in the `bin/defaults.sh` script, and its default value is

```sh
VERSION=$(git describe --long --tags)
```

One consequence of this is that if you do not commit your recent changes, the
build will change but the version string will not. The build scripts do not
protect you from this. On your own head be it.

## Running outside the containers

Before we start ndaunode, we have to create mock Release From Endowment (RFE) account data so that we have something to transfer into accounts.

- run ndaunode to create RFE mock data in a file:

    ```shell
    % ./ndaunode -make-mocks
    ```

The above command will create mock account data in a file in the config directory, and update the ndau config file to reflect that data.  If you have a chaos node up and running, you can push the mock data onto the chaos chain with the ndaunode command.

- run ndaunode to create RFE mock data on the chaos chain:

    ```shell
    % ./ndaunode -make-chaos-mocks
    ```

Which does the same as -make-mocks but writes the mock data to chaos chain as system variables instead of writing to a file.

Once the mock data is created, ndaunode can be run to start up the node, and tendermint can be run to start up the consensus engine:

- run ndaunode and point to noms db:

    % ./ndaunode -spec http://localhost:8000

- run tendermint to start up consensus engine:

    % ~/go/bin/tendermint node

## Transactions and Blocks

In traditional blockchain parlance, each ndau update is considered to be a "transaction", and each group of transactions that form a single update constructs a new "block". There is a maximum rate of block creation and multiple updates within a single block time will be consolidated.


# `ndau`: Set and retrieve values from the ndau blockchain

## CLI Usage

This is a low-level interface: it operates primarily on raw bytes and strings. While this is fine for debugging, it is not useful for most programs, which will normally want to operate on more structured data. For real use, write your program against the API offered by `pkg/tool`.

### Initial configuration

This tool needs to know the address of a ndau chain node.

```sh
ndau conf [ADDR]
```

`ADDR` must be the http address of a Tendermint node running the ndau chain. If unset, it defaults to the tendermint default port on localhost: `http://localhost:26657`

No other `ndau` command will work until the initial configuration is written.

If you're running a single ndau node, you can find which port it's running on from within the ndaunode directory with:

```sh
bin/defaults.sh docker-compose port tendermint 26657
```

That will generate something like this:

```
TMHOME=/Users/kentquirk/.tendermint/
NDAUHOME=/Users/kentquirk/.ndau/
0.0.0.0:32774
```

Which you can apply this way:

```sh
ndau conf $(bin/defaults.sh docker-compose port tendermint 26657 2>/dev/null)
```

`bin/conf-docker.sh` is simply a shortcut to perform the above command.

### Setting the RFE keys

Release from endowment transactions must be tightly controlled in order for ndau to work properly. At the same time, the blockchain policy council doesn't want to need to sign RFE transactions individually. Therefore, they have the right to delegate the power to release from endowment.

The BPC controls a system variable: `ReleaseFromEndowmentKeys`, which is a list of public keys. Valid RFE transactions must be signed with the private key corresponding to one of these public keys.

For demo purposes, we mock the BPC's role. That is to say that the `ndaunode` subcommands `-make-mocks` and `-make-chaos-mocks` both edit this tool's configuration file to set the `rfe_keys` setting appropriately. If you wish to use RFE transactions, run one of those tools, or if you are interacting with a nonlocal chain, copy the value from the machine on which those tools were initially run.

### Features demo

Once you have the configuration file written, you can finally interact with the ndau chain.

#### Create a new account

```sh
./ndau account new demo
```

#### Claim the account and set a validation key

```sh
$ ./ndau -v account claim demo
{
  "check_tx": {
    "fee": {}
  },
  "deliver_tx": {
    "fee": {}
  },
  "hash": "42CADE2557FA0E948ED73CF1C1C4013774A1912F",
  "height": 4403
}
```

This is a command which actually puts the account onto the ndau chain, with a 0 balance. The `-v` flag simply requests verbosity; it returns the JSON struct returned by the RPC command. Without it, there is no output on success; success is indicated by the exit code.


#### Examine the local account data

```sh
$ cat $(./ndau conf-path)
node = "http://0.0.0.0:32777"

[[accounts]]
  name = "demo"
  address = "ndai9vjzvt6547jirczgi8nrawshvwgk4zzcj8gy9njg64vi"
  [accounts.ownership]
    public = "kgHEIF45MoYR0qDtvUxIJpvwuFPs7avCLE+cp4ZN0TZSbxW7"
    private = "kgHEQOi7pmwV2bQOXFM9X4ItXLpREsnF5fVDlyktGDFinB5aXjkyhhHSoO29TEgmm/C4U+ztq8IsT5ynhk3RNlJvFbs="

  [[accounts.transfer]]
    public = "kgHEIP6AIjEZ7lN0vr58BBDXo+6R8ULpxUJ76ovu/+U7d1+W"
    private = "kgHEQKtqyg1LI8mzYBTXeVYLlcUVyJw0jxcL7N/sWCwVYat9/oAiMRnuU3S+vnwEENej7pHxQunFQnvqi+7/5Tt3X5Y="

[rfe]
  address = "ndnfvutjkcb4vxz73zinvcdx4ezy3pxpzwapfymavapzqyez"
  keys = ["kgHEQPDK/P/zfhQ3f3DigavI2K6v4VJ/aiyZjJZxLbY+JAXTle6Gmwsl8UBFERgjIn0odg5pCSU/g023ntAQncFj3zI=", "kgHEQGwZGJeiMoTUWOucRRh3xHYkhS4euIJtk5TGJta9GMsaBc/butRyfS20ByuogQQ9DR7Q7LdVnW9uk/mioADi74w="]

[nnr]
  address = "ndnpv7awaki9dkkeqcryk958h4jnf2itm3jw8wm3yy4asjxc"
  keys = ["kgHEQMtTPIC0tMhmrhZaP1qbHzvL2KkcytaNK+cXzZYYqSfY9EM3vcH3knX9xxjcgBBuRstCKlrVaEGw7xPOwHvR2Fk=", "kgHEQHjkGdA0je1nj1qxjCq5OYJBH9NtuSnDSGPNSM0f+W7pqAMA3e4CrmGU1F5L5NspvMVVWsv3laDl1z/M+TH+DRw="]
```

- The `node` setting was set by the `conf-docker.sh` script; it's the rpc port
at which the ndau tool can reach the ndau chain

- The `rfe` section was set by `./ndaunode -make-mocks` and `./ndaunode -make-chaos-mocks`. These are private keys authorized to release ndau from the endowment

- The `[[...]]` syntax denotes a single item of a top-level list whose name is contained in the double brackets

- The `accounts` list contains items with a human name, an ndau address, an ownership keypair, and potentially a transfer keypair

#### Examine the blockchain account data

```sh
$ ./ndau -v account query demo
{
  "balance": 0,
  "validationKeys": [
    "kgHEIP6AIjEZ7lN0vr58BBDXo+6R8ULpxUJ76ovu/+U7d1+W"
  ],
  "rewardsTarget": null,
  "incomingRewardsFrom": null,
  "delegationNode": null,
  "lock": null,
  "stake": null,
  "lastEAIUpdate": 593028790000000,
  "lastWAAUpdate": 0,
  "weightedAverageAge": 0,
  "Sequence": 1,
  "settlements": null,
  "RecourseSettings": {
    "Period": 0,
    "ChangesAt": null,
    "Next": null
  },
  "validationScript": null
}
{
  "response": {
    "log": "acct exists: true",
    "value": "jqdCYWxhbmNlAK5WYWxpZGF0aW9uS2V5c5GSAcQg/oAiMRnuU3S+vnwEENej7pHxQunFQnvqi+7/5Tt3X5atUmV3YXJkc1RhcmdldMCzSW5jb21pbmdSZXdhcmRzRnJvbZCuRGVsZWdhdGlvbk5vZGXApExvY2vApVN0YWtlwK1MYXN0RUFJVXBkYXRl0wACG1tGXpmArUxhc3RXQUFVcGRhdGUAsldlaWdodGVkQXZlcmFnZUFnZQCoU2VxdWVuY2UBq1NldHRsZW1lbnRzkLJTZXR0bGVtZW50U2V0dGluZ3ODplBlcmlvZACpQ2hhbmdlc0F0wKROZXh0wLBWYWxpZGF0aW9uU2NyaXB0xAA=",
    "height": "1162"
  }
}
```

Notes about this output:

- `RecourseSettings` is set to the default settlement duration, which is a system variable. It was set during the `change-transfer-key` transaction which assigned the transfer key. Whenever a CTK transaction is signed with the ownership key and the escrow duration is 0, the duration is updated to the default.

- The second JSON object returned is present because we used the `-v` flag. It again contains the raw response from the RPC command.

    - the `log` field says "acct exists: true". If the account were not present on the blockchain, the `log` field would says "does not exist", and the account zero value would have been returned. The `log` field is currently the only way to determine whether an account exists on the blockchain.
    - the `value` field contains the packed representation of the object

#### Release some ndau into the account

The first argument of `rfe` is a floating-point quantity of ndau. For more precision, use the `-napu` flag to set an integer number of napu instead.

The third argument of `rfe` is the index of the key from `rfe_keys` to use to sign the RFE transaction. If `rfe_keys` is unset or the index is out of bounds, the `rfe` command will fail before sending any transaction to the blockchain.

```sh
$ ./ndau -v rfe 10 demo
Release from endowment: 10 ndau to ndaqmatgkap2ff62hkqpwmyzfr6uzdrct6g6mmkk38q3eekk
{
  "check_tx": {
    "fee": {}
  },
  "deliver_tx": {
    "fee": {}
  },
  "hash": "E1848C9ABE066521E951BC716D590A10CAB41998",
  "height": 5320
}
```

#### Transfer ndau from one account to another

Transfers are easy. They have the arguments `qty` `source` `dest`.

```sh
$ ./ndau -v transfer 1 demo prgn
Transfer 1 ndau from ndaqmatgkap2ff62hkqpwmyzfr6uzdrct6g6mmkk38q3eekk to ndart5whvaifbefkmj67qsrzae6bgf42kaga85nf5twhn2fu
{
  "check_tx": {
    "fee": {}
  },
  "deliver_tx": {
    "fee": {}
  },
  "hash": "62EADC24A619289362487658437CCE57384C8053",
  "height": 5711
}
```

For something more interesting, we can create an account on the blockchain by transfering to it. Note that this example requires the [`toml` command](https://github.com/chrisdickinson/toml-cli).

```sh
$ # create a new account without sending anything to the blockchain
$ ./ndau account new demo-receiver
$ # save the address of the new account
$ demo_receiver_addr=$(cat $(./ndau conf-path) | toml | jq '.accounts[] | select(.name == "demo-receiver") | .address' --raw-output) && echo $demo_receiver_addr
$ # transfer 1 ndau from the demo account to the address of the new receiver
$ ./ndau -v transfer .1 demo --to-address=$demo_receiver_addr
Transfer 0.1 ndau from ndai9vjzvt6547jirczgi8nrawshvwgk4zzcj8gy9njg64vi to ndak6qmrbzj35rrc9x2i8dv3icvnewvibb4542zjdv5wwb97
{
  "check_tx": {
    "fee": {}
  },
  "deliver_tx": {
    "fee": {}
  },
  "hash": "E943565F8A06B6A3F3683AA2288B3688E727B2DB",
  "height": 1256
}
$ # now let's look at the receiver on the blockchain
$ ./ndau account query --address=$demo_receiver_addr
{
  "balance": 10000000,
  "validationKeys": null,
  "rewardsTarget": null,
  "incomingRewardsFrom": null,
  "delegationNode": null,
  "lock": null,
  "stake": null,
  "lastEAIUpdate": 593029846000000,
  "lastWAAUpdate": 593029878000000,
  "weightedAverageAge": 29090909,
  "Sequence": 0,
  "settlements": null,
  "RecourseSettings": {
    "Period": 0,
    "ChangesAt": null,
    "Next": null
  },
  "validationScript": null
}
```

Note that the balance remains 0, and the account is unclaimed.

#### Change the settlement settings

We can allow senders to create a "settlement period" which will cause transfers to be delayed before they can be spent.
The default will not always be convenient. A user might want to set a differet settlement period. They might do so like this:

```sh
$ ./ndau account query demo
{
  "balance": 475999000,
  "validationKeys": [
    "kgHEIP6AIjEZ7lN0vr58BBDXo+6R8ULpxUJ76ovu/+U7d1+W"
  ],
  "rewardsTarget": null,
  "incomingRewardsFrom": null,
  "delegationNode": null,
  "lock": null,
  "stake": null,
  "lastEAIUpdate": 593030087000000,
  "lastWAAUpdate": 0,
  "weightedAverageAge": 0,
  "Sequence": 11,
  "settlements": null,
  "RecourseSettings": {
    "Period": 3600000000,
    "ChangesAt": null,
    "Next": null
  },
  "validationScript": null
}
```

The escrow settings will now change to 1 hour. Settlement periods only change after the current escrow period has expired.

If we now send another .2 ndau to our account above:

```sh
$ ./ndau -v transfer .2 demo --to-address=$demo_receiver_addr
Transfer 0.2 ndau from ndai9vjzvt6547jirczgi8nrawshvwgk4zzcj8gy9njg64vi to ndak6qmrbzj35rrc9x2i8dv3icvnewvibb4542zjdv5wwb97
{
  "check_tx": {
    "fee": {}
  },
  "deliver_tx": {
    "fee": {}
  },
  "hash": "36CAFF6E703983641561A9DFB3EEF2495AFE410A",
  "height": 1286
}
```

We can see the unsettled amount in the target account:

```sh
$ ./ndau account query --address=$demo_receiver_addr
{
  "balance": 130000000,
  "validationKeys": null,
  "rewardsTarget": null,
  "incomingRewardsFrom": null,
  "delegationNode": null,
  "lock": null,
  "stake": null,
  "lastEAIUpdate": 593029846000000,
  "lastWAAUpdate": 593030185000000,
  "weightedAverageAge": 284384615,
  "Sequence": 0,
  "settlements": [
    {
      "Qty": 20000000,
      "Expiry": 593033785000000
    }
  ],
  "RecourseSettings": {
    "Period": 0,
    "ChangesAt": null,
    "Next": null
  },
  "validationScript": null
}
```


### Changing the validator set

`ndau cvc PUBKEY POWER` sends a command validator change. On the real blockchain, this is likely to be disabled, but it works for now.

### Querying the node status

You may simply wish to know some information about the node's internal state. `ndau info` asks for that state, and pretty-prints the results.

## API Usage

The `pkg/tool` package provides a high-level interface with which external programs can interact with the ndau chain.

After some research, we decided that [MessagePack](https://msgpack.org/index.html) is the best available canonical data format for the ndau chain, and [msgp](https://github.com/tinylib/msgp) is the best available library for interface use.

To use it:

2. Get the tool: `go get github.com/tinylib/msgp && go install github.com/tinylib/msgp`.
1. Write the structs you wish to encode in a single file.
3. Add the following snippet to your struct file: `//go:generate msgp`.
4. Run `go generate ./...` to generate the implementations for `msgp.Encodable` and `msgp.Decodable`, which satisfy the high-level interface provided by this package.
5. You're all set! Pass your structs directly to the get/set/etc. functions.

As a matter of style, the methods of this class which accept and return `[]byte` should be avoided in favor of strucured data whenever possible. They may become deprecated and vanish in future versions of this tool.

### Schemas

MessagePack is fundamentally schemaless. We are likely to search for and/or write a schema discovery/validation tool at some point in the future, but that time has not yet come.

## Tendermint/ndau RESTful API server

The `ndautool` can function as a Tendermint/ndau chain API server.  To start in this mode use the `server` command with a port argument, like so:

```sh
% ndau server 8005
```

To test the server once it's started, you can make a `status` query in your browser such as the following, given the server started on port 8005 as above:

http://localhost:8005/status

This query should return JSON similar to the following:

```sh
{
    "node_info": {
        "pub_key": {
            "type": "ed25519",
            "data": "767B6318D0DD94F6A24BC36A27D43C84E11943663C80874DB8B611A8D8C0BBCF"
        },
        "listen_addr": "172.17.0.6:26656",
        "network": "test-chain-C4FkcS",
        "version": "0.18.0",
        "channels": "QCAhIiMwOAA=",
        "moniker": "`hostname`",
        "other": [
            "wire_version=0.7.3",
            "p2p_version=0.5.0",
            "consensus_version=v1/0.2.2",
            "rpc_version=0.7.0/3",
            "tx_index=on",
            "rpc_addr=tcp://0.0.0.0:26657"
        ]
    },
    "pub_key": {
        "type": "ed25519",
        "data": "99B0A0D14E63EE299F38DB3E9DD6E529E1B51864C63D7855C67F0BFD48BCF9AB"
    },
    "latest_block_hash": "D39E083AB7F38DB772037E53CB7689228C8B547E",
    "latest_app_hash": "933360BC6EC909869C10781FE97B1F22B9B57156",
    "latest_block_height": 2356,
    "latest_block_time": "2018-05-25T18:57:59.757Z",
    "syncing": false,
    "validator_status": {
        "voting_power": 10
    }
}
```
