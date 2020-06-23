# Vulcanize DB

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-contract-watcher)](https://goreportcard.com/report/github.com/vulcanize/eth-contract-watcher)

> Tool for watching contract data on Ethereum using contract ABIs


## Table of Contents
1. [Background](#background)
1. [Install](#install)
1. [Usage](#usage)
1. [Testing](#testing)
1. [Contributing](#contributing)
1. [License](#license)


## Background
An Ethereum smart contract's variables are stored in the storage [patricia trie](https://eth.wiki/en/fundamentals/patricia-tree) associated with the contract's account.
Contract variable values can be read directly from the state database if their keys are known, or indirectly through public getter methods the contract exposes.
It is not always feasible to determine the key for a particular storage variable without prior knowledge of the contract source code, as the mapping is dependent on the order of variable declaration and on the compiler used.
Public getter methods provide a more transparent means of accessing storage data, but while it is good practice for contracts to expose such methods it is not required and in some cases is purposely avoided.
An archival node is required for accessing historical data contract variable values as the Ethereum state and storage tries are pruned below the gc limit of a full node.

Contracts also have the ability to emit data in the form of events, allowing the data to be stored in [receipt logs](https://github.com/ethereum/go-ethereum/blob/master/core/types/log.go#L31).
Receipt logs provide a cheaper alternative for storing small amounts of contract data outside of the storage trie, where they remain available below the gc limit of a full node.
Receipt logs can have up to four topics, the first topic is reserved for the event signature while the rest are available
for up to three indexed arguments of the event. Non-indexed argument values are encoded within the `data` field of the log.

E.g. for `event Transfer(address indexed _from, address indexed _to, uint256 _value)`

The first log topic is the keccak256 hash of the event: `keccak256("Transfer(address,address,uint256)")`
The second topic will contain the `_from` address, the third topic will contain the `_to` address, and the data field will contain
the `_value` amount.

Both methods and events are defined in a contract's [ABI](https://solidity.readthedocs.io/en/v0.5.3/abi-spec.html),
as such we can use any contract ABI to automate fetching, transformation, and indexing of the contract's data into an external database with minimal additional configuration.

`eth-contract-wacther` is a generic contract watcher that takes advantage of this.
It can watch any and all events for a given contract provided the contract's ABI is available.
It also provides some state variable coverage by automating polling of public getter methods, with some restrictions:
1. The method must have 2 or less arguments
1. The method's arguments must all be of type address or bytes32 (hash)
1. The method must return a single value

In the future we intend to expand the functionality to support direct watching of storage slots by leveraging a [state-diffing Ethereum client](https://github.com/vulcanize/go-ethereum/tree/statediff_at_anyblock-1.9.13)

Note:

`eth-contract-wacther` builds on the headers synced into Postgres by [eth-header-sync](https://github.com/vulcanize/eth-header-sync), it uses these to direct the fetching of contract data and anchors the data
to the headers through foreign keys. It needs to be ran in parallel (or tandem) to an eth-header-sync process operating on the same Postgres database.

`eth-contract-wacther` requires the contract ABI be available on Etherscan if it is not provided in the config file by the user.

If method polling is turned on we require an archival node at the ETH ipc endpoint in our config, otherwise we only need to connect to a full node.

## Install

1. [Dependencies](#dependencies)
1. [Building the project](#building-the-project)
1. [Setting up the database](#setting-up-the-database)
1. [Configuring a synced Ethereum node](#configuring-a-synced-ethereum-node)
1. [Runing eth-header-sync](#running-eth-header-sync)

### Dependencies
 - Go 1.12+
 - Postgres 11.2
 - Ethereum Node
   - [go-ehereum](https://github.com/ethereum/go-ethereum/releases) (1.8.23+)
   - [openethereum](https://github.com/openethereum/openethereum/releases) (1.8.11+)
 - [eth-header-sync](https://github.com/vulcanize/eth-header-sync) (0.0.10+)

### Building the project
Download the codebase to your local `GOPATH` via:

`go get github.com/vulcanize/eth-contract-watcher`

Move to the project directory:

`cd $GOPATH/src/github.com/vulcanize/eth-contract-watcher`

Be sure you have enabled Go Modules (`export GO111MODULE=on`), and build the executable with:

`make build`

If you need to use a different dependency than what is currently defined in `go.mod`, it may helpful to look into [the replace directive](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive).
This instruction enables you to point at a fork or the local filesystem for dependency resolution.

If you are running into issues at this stage, ensure that `GOPATH` is defined in your shell.
If necessary, `GOPATH` can be set in `~/.bashrc` or `~/.bash_profile`, depending upon your system.
It can be additionally helpful to add `$GOPATH/bin` to your shell's `$PATH`.

### Setting up the database
1. Install Postgres
1. Create a superuser for yourself and make sure `psql --list` works without prompting for a password.
1. `createdb vulcanize_public`
1. `cd $GOPATH/src/github.com/vulcanize/eth-header-sync`
1.  Run the migrations: `make migrate HOST_NAME=localhost NAME=vulcanize_public PORT=5432`
    - There are optional vars `USER=username` and `PASS=password` if the database user is not the default user `postgres` and/or a password is present
    - To rollback a single step: `make rollback NAME=vulcanize_public`
    - To rollback to a certain migration: `make rollback_to MIGRATION=n NAME=vulcanize_public`
    - To see status of migrations: `make migration_status NAME=vulcanize_public`

    * See below for configuring additional environments
    
In some cases (such as recent Ubuntu systems), it may be necessary to overcome failures of password authentication from
localhost. To allow access on Ubuntu, set localhost connections via hostname, ipv4, and ipv6 from peer/md5 to trust in: /etc/postgresql/<version>/pg_hba.conf

(It should be noted that trusted auth should only be enabled on systems without sensitive data in them: development and local test databases)

### Running eth-header-sync
The contract watcher runs in parallel (or in tandem) to [eth-header-sync](https://github.com/vulcanize/eth-header-sync).
eth-header-sync validated headers direct fetching of contract data and anchor the contract data to them in Postgres using foreign keys.
More information on how to run the header sync can be found in that repositories [README](https://github.com/vulcanize/eth-header-sync/blob/master/README.md).

## Usage

After building the binary, run as

`./eth-contract-watcher watch --config=./environments/example.toml`

### Configuration

The config file linked to in the `--config` cli flag should have the below format

```toml
  [database]
    name     = "vulcanize_public"
    hostname = "localhost"
    port     = 5432

  [client]
    rpcPath  = "/Users/user/Library/Ethereum/geth.ipc"

  [contract]
    network  = ""
    addresses  = [
        "contractAddress1",
        "contractAddress2"
    ]
    [contract.contractAddress1]
        abi    = 'ABI for contract 1'
        startingBlock = 982463
    [contract.contractAddress2]
        abi    = 'ABI for contract 2'
        events = [
            "event1",
            "event2"
        ]
		eventArgs = [
			"arg1",
			"arg2"
		]
        methods = [
            "method1",
			"method2"
        ]
		methodArgs = [
			"arg1",
			"arg2"
		]
        startingBlock = 4448566
        piping = true
````

- `database` fields hold the paramaters for connection to the Postgres database
- `client.rpcPath` is the RPC path to an Ethereum full or archival node
- The `contract` section defines which contracts we want to watch and with which conditions.
- `network` is only necessary if the ABIs are not provided and wish to be fetched from Etherscan.
    - Empty or nil string indicates mainnet
    - "ropsten", "kovan", and "rinkeby" indicate their respective networks
- `addresses` lists the contract addresses we are watching and is used to load their individual configuration parameters
- `contract.<contractAddress>` are the sub-mappings which contain the parameters specific to each contract address
    - `abi` is the ABI for the contract; if none is provided the application will attempt to fetch one from Etherscan using the provided address and network
    - `events` is the list of events to watch
        - If this field is omitted or no events are provided then by default *all* events extracted from the ABI will be watched
        - If event names are provided then only those events will be watched
    - `eventArgs` is the list of arguments to filter events with
        - If this field is omitted or no eventArgs are provided then by default watched events are not filtered by their argument values
        - If eventArgs are provided then only those events which emit at least one of these values as an argument are watched
    - `methods` is the list of methods to poll
        - If this is omitted or no methods are provided then by default NO methods are polled
        - If method names are provided then those methods will be polled, provided
            1) Method has two or less arguments
            1) Arguments are all of address or hash types
            1) Method returns a single value
    - `methodArgs` is the list of arguments to limit polling methods to
        - If this field is omitted or no methodArgs are provided then by default methods will be polled with every combination of the appropriately typed values that have been collected from watched events
        - If methodArgs are provided then only those values will be used to poll methods
    - `startingBlock` is the block we want to begin watching the contract, usually the deployment block of that contract
    - `piping` is a boolean flag which indicates whether or not we want to pipe return method values forward as arguments to subsequent method calls

At the very minimum, for each contract address an ABI and a starting block number need to be provided (or just the starting block if the ABI can be reliably fetched from Etherscan).
With just this information we will be able to watch all events at the contract, but with no additional filters and no method polling.

### Output

Transformed events and polled method results are committed to Postgres in schemas and tables generated according to the contract abi.

Schemas are created for each contract using the naming convention `<sync-type>_<lowercase contract-address>`.
Under this schema, tables are generated for watched events as `<lowercase event name>_event` and for polled methods as `<lowercase method name>_method`.
The 'method' and 'event' identifiers are tacked onto the end of the table names to prevent collisions between methods and events of the same lowercase name.

### Example:

Modify `./environments/example.toml` to replace the empty `rpcPath` with a path that points to an ethjson_rpc endpoint (e.g. a local geth node ipc path or an Infura url).
This endpoint should be for an archival eth node if we want to perform method polling as this configuration is currently set up to do. To work with a non-archival full node,
remove the `balanceOf` method from the `0x8dd5fbce2f6a956c3022ba3663759011dd51e73e` (TrueUSD) contract.

If you are operating a header sync vDB, run:

 `./eth-contract-watcher contractWatcher --config=./environments/example.toml --mode=header`

If instead you are operating a full sync vDB and provided an archival node IPC path, run in full mode:

 `./eth-contract-watcher contractWatcher --config=./environments/example.toml --mode=full`

This will run the contractWatcher and configures it to watch the contracts specified in the config file. Note that
by default we operate in `header` mode but the flag is included here to demonstrate its use.

The example config we link to in this example watches two contracts, the ENS Registry (0x314159265dD8dbb310642f98f50C066173C1259b) and TrueUSD (0x8dd5fbCe2F6a956C3022bA3663759011Dd51e73E).

Because the ENS Registry is configured with only an ABI and a starting block, we will watch all events for this contract and poll none of its methods. Note that the ENS Registry is an example
of a contract which does not have its ABI available over Etherscan and must have it included in the config file.

The TrueUSD contract is configured with two events (`Transfer` and `Mint`) and a single method (`balanceOf`), as such it will watch these two events and use any addresses it collects emitted from them
to poll the `balanceOf` method with those addresses at every block. Note that we do not provide an ABI for TrueUSD as its ABI can be fetched from Etherscan.

For the ENS contract, it produces and populates a schema with four tables"
`header_0x314159265dd8dbb310642f98f50c066173c1259b.newowner_event`
`header_0x314159265dd8dbb310642f98f50c066173c1259b.newresolver_event`
`header_0x314159265dd8dbb310642f98f50c066173c1259b.newttl_event`
`header_0x314159265dd8dbb310642f98f50c066173c1259b.transfer_event`

For the TrusUSD contract, it produces and populates a schema with three tables:

`header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.transfer_event`
`header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.mint_event`
`header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.balanceof_method`

Column ids and types for these tables are generated based on the event and method argument names and types and method return types, resulting in tables such as:

Table "header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.transfer_event"

|  Column    |         Type          | Collation | Nullable |                                           Default                                            | Storage  | Stats target | Description |
|:----------:|:---------------------:|:---------:|:--------:|:--------------------------------------------------------------------------------------------:|:--------:|:------------:|:-----------:|
| id         | integer               |           | not null | nextval('header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.transfer_event_id_seq'::regclass) | plain    |              |             |
| header_id  | integer               |           | not null |                                                                                              | plain    |              |             |
| token_name | character varying(66) |           | not null |                                                                                              | extended |              |             |
| raw_log    | jsonb                 |           |          |                                                                                              | extended |              |             |
| log_idx    | integer               |           | not null |                                                                                              | plain    |              |             |
| tx_idx     | integer               |           | not null |                                                                                              | plain    |              |             |
| from_      | character varying(66) |           | not null |                                                                                              | extended |              |             |
| to_        | character varying(66) |           | not null |                                                                                              | extended |              |             |
| value_     | numeric               |           | not null |                                                                                              | main     |              |             |


Table "header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.balanceof_method"

|   Column   |         Type          | Collation | Nullable |                                            Default                                             | Storage  | Stats target | Description |
|:----------:|:---------------------:|:---------:|:--------:|:----------------------------------------------------------------------------------------------:|:--------:|:------------:|:-----------:|
| id         | integer               |           | not null | nextval('header_0x8dd5fbce2f6a956c3022ba3663759011dd51e73e.balanceof_method_id_seq'::regclass) | plain    |              |             |
| token_name | character varying(66) |           | not null |                                                                                                | extended |              |             |
| block      | integer               |           | not null |                                                                                                | plain    |              |             |
| who_       | character varying(66) |           | not null |                                                                                                | extended |              |             |
| returned   | numeric               |           | not null |                                                                                                | main     |              |             |

The addition of '_' after column names is to prevent collisions with reserved Postgres words.

Also notice that the contract address used for the schema name has been down-cased.

## Testing
- Replace the empty `rpcPath` in the `environments/testing.toml` with a path to a full node's eth_jsonrpc endpoint (e.g. local geth node ipc path or infura url)
    - Note: must be mainnet
    - Note: integration tests require configuration with an archival node
- `make test` will run the unit tests and skip the integration tests
- `make integrationtest` will run just the integration tests
- `make test` and `make integrationtest` setup a clean `vulcanize_testing` db


## Contributing
Contributions are welcome!

VulcanizeDB follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/1/4/code-of-conduct).

## License
[AGPL-3.0](LICENSE) © Vulcanize Inc
