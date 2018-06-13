# `ndau`: Set and retrieve values from the ndau blockchain

## CLI Usage

This is a low-level interface: it operates primarily on raw bytes and strings. While this is fine for debugging, it is not useful for most programs, which will normally want to operate on more structured data. For real use, write your program against the API offered by `pkg/tool`.

### Initial configuration

This tool needs to know the address of a ndau chain node.

```sh
ndau conf [ADDR]
```

`ADDR` must be the http address of a Tendermint node running the ndau chain. If unset, it defaults to the tendermint default port on localhost: `http://localhost:46657`

No other `ndau` command will work until the initial configuration is written.

If you're running a single ndau node, you can find which port it's running on from within the ndaunode directory with:

```sh
bin/defaults.sh docker-compose port tendermint 46657
```

That will generate something like this:

```
TMHOME=/Users/kentquirk/.tendermint/
NDAUHOME=/Users/kentquirk/.ndau/
0.0.0.0:32774
```

Which you can apply this way:

```sh
ndau conf $(bin/defaults.sh docker-compose port tendermint 46657 2>/dev/null)
```

### Changing the validator set

`ndau gtvc PUBKEY POWER` sends a globally trusted validator change. This won't work forever, but it does nicely for now.

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
        "listen_addr": "172.17.0.6:46656",
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
            "rpc_addr=tcp://0.0.0.0:46657"
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
