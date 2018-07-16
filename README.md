# Ndau

This is the implementation of the Ndau chain. See the Ndau design whitepaper for details.  All Ndau transactions are stored on this chain.  An (incomplete) list of possible transactions on the Ndau chain include:

- Transfer
- ChangeTransferKey
- ReleaseFromEndowment
- ChangeEscrowPeriod
- Delegate
- ComputeEAI
- GTValidatorChange


## Install

- install [glide](https://github.com/Masterminds/glide):

    ```shell
    curl https://glide.sh/get | sh
    ```

- update dependencies

    ```shell
    glide install
    ```

- check build

    ```shell
    go build ./... && go test ./...
    ```

## Quick Start

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


