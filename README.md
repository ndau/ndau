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

Note that this doesn't actually interact with the ndau chain; it simply creates a named keypair in the local config file.

#### Set the account's transfer key

```sh
$ ./ndau -v account change-transfer-key demo
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

This is the command which actually puts the account onto the ndau chain, with a 0 balance. The `-v` flag simply requests verbosity; it returns the JSON struct returned by the RPC command. Without it, there is no output on success; success is indicated by the exit code.


#### Examine the local account data

```sh
$ cat $(./ndau conf-path)
node = "0.0.0.0:32768"
rfe_keys = ["kgHEQINkiAStVH+IYRqbIsWLnt0AnkwmpdOaFUVxF8gQVu5UeTIBgl+DT6tlsVNmt4Bih6j4C/98Bo36tU+wYThuBDU=", "kgHEQHDoaCjbGBPnK9WTA7ilfPquZ/m0vWesE6X6hr/2pyJF5Gkkx5aobS2WNcEwVdTJPUrxaaIAm55lMDO3uu2QXpk="]

[[accounts]]
  name = "demo"
  address = "ndart5whvaifbefkmj67qsrzae6bgf42kaga85nf5twhn2fu"
  [accounts.ownership]
    public = "kgHEINabBrNCZOowBrCwnHrDVotfoxVYzMNsd23De42DgdJl"
    private = "kgHEQO9kMOQPnebLzyRPMDz4lUgf8avp5uELNHM0EBDvp+Zv1psGs0Jk6jAGsLCcesNWi1+jFVjMw2x3bcN7jYOB0mU="
  [accounts.transfer]
    public = "kgHEIPLE0yICydwzkMJDZAcn+MFmf9wtpTlYUUT3BIoxiat2"
    private = "kgHEQOc+mQJzJ3taAMoqh88Pb83Yf+hv/Lh3uMexCnH3xhmQ8sTTIgLJ3DOQwkNkByf4wWZ/3C2lOVhRRPcEijGJq3Y="
```

- The `node` setting was set by the `conf-docker.sh` script; it's the rpc port
at which the ndau tool can reach the ndau chain

- The `rfe_keys` are set by `./ndaunode -make-mocks` and `./ndaunode -make-chaos-mocks`. These are private keys authorized to release ndau from the endowment

- The `[[...]]` syntax denotes a single item of a top-level list whose name is contained in the double brackets

- The `accounts` list contains items with a human name, an ndau address, an ownership keypair, and potentially a transfer keypair

#### Examine the blockchain account data

```sh
$ ./ndau -v account query demo
{
  "Balance": 0,
  "TransferKey": {},
  "RewardsTarget": null,
  "DelegationNode": null,
  "Lock": null,
  "Stake": null,
  "LastWAAUpdate": 0,
  "WeightedAverageAge": 0,
  "Sequence": 0,
  "Escrows": null,
  "EscrowSettings": {
    "Duration": 1296000000000,
    "ChangesAt": null,
    "Next": null
  }
}
{
  "response": {
    "log": "exists",
    "value": "i6dCYWxhbmNlAKtUcmFuc2ZlcktleZIBxCCt0dBvjZjokY9eN5MW3w/5y77kRW+mjxPRtiYparr5D61SZXdhcmRzVGFyZ2V0wK5EZWxlZ2F0aW9uTm9kZcCkTG9ja8ClU3Rha2XArUxhc3RXQUFVcGRhdGUAsldlaWdodGVkQXZlcmFnZUFnZQCoU2VxdWVuY2UAp0VzY3Jvd3OQrkVzY3Jvd1NldHRpbmdzg6hEdXJhdGlvbtMAAAEtv56gAKlDaGFuZ2VzQXTApE5leHTA",
    "height": "4533"
  }
}
```

Notes about this output:

- `TransferKey` is visualized as an empty struct. This is a known bug: [oneiro-ndev/signature#4](https://github.com/oneiro-ndev/signature/issues/4). If it were unset, it would be `null`.

- `EscrowSettings` is set to the default escrow duration, which is a system variable. It was set during the `change-transfer-key` transaction which assigned the transfer key. Whenever a CTK transaction is signed with the ownership key and the escrow duration is 0, the duration is updated to the default.

- The second JSON object returned is present because we used the `-v` flag. It again contains the raw response from the RPC command.

    - the `log` field says "exists". If the account were not present on the blockchain, the `log` field would says "does not exist", and the account zero value would have been returned. The `log` field is currently the only way to determine whether an account exists on the blockchain.
    - the `value` field contains the packed representation of the object

#### Release some ndau into the account

The first argument of `rfe` is a floating-point quantity of ndau. For more precision, use the `-napu` flag to set an integer number of napu instead.

The third argument of `rfe` is the index of the key from `rfe_keys` to use to sign the RFE transaction. If `rfe_keys` is unset or the index is out of bounds, the `rfe` command will fail before sending any transaction to the blockchain.

```sh
$ ./ndau -v rfe 10 demo 0
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
$ ./ndau -v transfer 1 demo --to_address=$demo_receiver_addr
Transfer 1 ndau from ndaqmatgkap2ff62hkqpwmyzfr6uzdrct6g6mmkk38q3eekk to ndahqajzp8h5ke5nf8gr5fj6ewuh86hhud6ymw5p3fx7jsvq
{
  "check_tx": {
    "fee": {}
  },
  "deliver_tx": {
    "fee": {}
  },
  "hash": "AF7704986E10F4E1CB80DFD27754081D2071321C",
  "height": 6262
}
$ # now let's look at the receiver on the blockchain
$ ./ndau account query --address=$demo_receiver_addr
{
  "Balance": 0,
  "TransferKey": null,
  "RewardsTarget": null,
  "DelegationNode": null,
  "Lock": null,
  "Stake": null,
  "LastWAAUpdate": 15418601000000,
  "WeightedAverageAge": 0,
  "Sequence": 0,
  "Escrows": [
    {
      "Qty": 100000000,
      "Expiry": 16714601000000
    }
  ],
  "EscrowSettings": {
    "Duration": 0,
    "ChangesAt": null,
    "Next": null
  }
}
```

Note that the balance remains 0, but an item has been added to the escrows list containing the appropriate number of napu, and an expiry time based on the source's escrow settings.

#### Change the escrow settings

We just demo'd how accounts have default escrow settings, and how transfers are only credited to the account balance once the escrow period ends. However, the default will not always be convenient. A user might want to set a much shorter escrow period. They might do so like this:

```sh
$ ./ndau account change-escrow-period demo 1h
$ ./ndau account query demo
{
  "Balance": 800000000,
  "TransferKey": {},
  "RewardsTarget": null,
  "DelegationNode": null,
  "Lock": null,
  "Stake": null,
  "LastWAAUpdate": 0,
  "WeightedAverageAge": 0,
  "Sequence": 2,
  "Escrows": null,
  "EscrowSettings": {
    "Duration": 1296000000000,
    "ChangesAt": 16715346000000,
    "Next": 3600000000
  }
}
```

The escrow settings will now change to 1 hour, after the current escrow period has expired.

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
