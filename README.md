# Alice TSS Example

Based on the [Alice](https://github.com/getamis/alice) library, this program demonstrates a Threshold Signature Scheme (TSS) implementation using [go-libp2p](https://github.com/libp2p/go-libp2p) for peer-to-peer networking.

The system provides three main TSS operations:

1. `dkg`: generate shares
2. `signer`: sign a message
3. `reshare`: refresh shares

## Configuration

All nodes require a configuration file with the following structure:

**config.yml**
```yaml
port: 10001
rpc: 1234

store:
  type: "badger"
  path: "./node.test/badger1"
```

### Configuration Parameters

1. `port`: P2P networking port that this node will listen on
2. `rpc`: HTTP port that the JSON-RPC server is exposed on
3. `store.type`: Database type (currently supports "badger" and "mock")
4. `store.path`: Directory path where the Badger database files are stored

### DKG
#### Request

Besides the common inputs, DKG will need another two inputs.

1. `rank`: The rank of this node during HTSS algorithm.
2. `threshold`: The threshold that needed to generate a valid signature.

```shell
curl --request POST \
  --url http://127.0.0.1:1234/tss \
  --header 'Content-Type: application/json' \
  --data '{
	"jsonrpc":"2.0",
	"method": "signer.RegisterDKG",
	"params": [
	],
	"id": "12"
}'
```

#### Output

```json
{
	"jsonrpc": "2.0",
	"result": {
		"Data": {
			"hash": "0x5a73c8fb1b418fdd33985b0b3a8561243abbb5cf1af3f0a368502939e3a4d658",
			"config": {
				"rank": 0,
				"threshold": 2
			}
		}
	},
	"id": "12"
}
```
1. `hash`: The hash of the DKG. Use `hash` to get data.
    ```shell
    curl --request POST \
      --url http://127.0.0.1:1234/tss \
      --header 'Content-Type: application/json' \
      --data '{
      "jsonrpc": "2.0",
      "method": "signer.GetDKG",
      "params": [
        {
          "key": "hash"
        }
      ],
      "id": "12"
    }'   
    ```
    ```json
    {
        "jsonrpc": "2.0",
        "result": {
            "Data": {
                "share": "m/qLSTBTc/f7My0iPs5woPKfoKgT+XFnStmX4owvxUgRJ2fgdnKxztwfYMNpO1aE8JqjZnA9lsHw+gLl4yWEWA==",
                "pubkey": {
                    "X": "d890e326fc2ea4f67d8eb6dc451779836fe7a15a2643b901d342f76ba06d7674",
                    "Y": "d637a8b69734453627a4d9c324f007b45c819c8240d6e75ed4adb66ede844b16"
                },
                "publicKey": "02d890e326fc2ea4f67d8eb6dc451779836fe7a15a2643b901d342f76ba06d7674",
                "address": "0x6dc09db941ff502d1ed186cb72e863dc405787a8",
                "bks": {
                    "QmTnNGyMB9ZzPVWnnxHMuvUpNHEEe1iDiih2KAzuX8yoSQ": {
                        "X": "109729591954224079959826078399798641625124977056152422372303221817148775946583",
                        "Rank": 0
                    },
                    "QmWSwYK1spmsKsSk4tvZ5UdsahhQkJZ9Tv5KewF3dCXeTG": {
                        "X": "81027363746734659626980593804036585447339663816871029076125049049095054333520",
                        "Rank": 0
                    },
                    "QmYY7udrgptw5NbiBnujvAwnm2vVxS6yet1iMCrERwi2h5": {
                        "X": "42332349435963829328874129794257492944267227575524735938800760087749784934735",
                        "Rank": 0
                    }
                }
            }
        },
        "id": "12"
    }
    ```
2. Result DKG
   1. `share`: The respective encrypted share of the node. The value of share in these output files must be different.
   2. `pubkey`: The public key. The value of public key in these output files must be the same.
   3. `address`: Address of public key.
   4. `bks`: The Birkhoff parameter of all nodes. Each Birkhoff parameter contains x coordinate and the rank.

### Signer
#### Request

Signer will need another three inputs.

1. `hash`: The hash of the DKG.
2. `pubkey`: The public key generated from DKG.
3. `msg`: The message to be signed.

e.g.

```shell
curl --request POST \
  --url http://127.0.0.1:1234/tss \
  --header 'Content-Type: application/json' \
  --data '{
	"jsonrpc": "2.0",
	"method": "signer.SignMessage",
	"params": [
		{
			"data": {
				"hash":"hash",
				"pubkey": "pubkey",
				"message": "msg"
			}
		}
	],
	"id": "12"
}'
```

#### Output
```json
{
	"jsonrpc": "2.0",
	"result": {
		"Data": "0x064e6b2999d1c97a9b73f17d4ec5730a3e5c8c4b240aab0b09f31b18de80dc8a"
	},
	"id": "12"
}
```
After signing, we will have a `hash` to get signature. And the value of the signature (both `r` and `s`).

Request get signature
```shell
curl --request POST \
  --url http://127.0.0.1:1234/tss \
  --header 'Content-Type: application/json' \
  --data '{
	"jsonrpc": "2.0",
	"method": "signer.GetKey",
	"params": [
		{
			"key": "hash"
		}
	],
	"id": "12"
}'
```
Response
```json
{
	"jsonrpc": "2.0",
	"result": {
		"Data": {
			"hash": "hash",
			"r": "r",
			"s": "s"
		}
	},
	"id": "12"
}
```

### Reshare
#### Request

Reshare will need another two inputs.

1. `hash`: The hash of the DKG.
2. `pubkey`: The public key generated from DKG.

```shell
curl --request POST \
  --url http://127.0.0.1:1234/tss \
  --header 'Content-Type: application/json' \
  --data '{
	"jsonrpc":"2.0",
	"method": "signer.Reshare",
	"params": [
		{
			"data": {
				"hash":"hash",
				"pubkey": "pubkey"
			}
		}
	],
	"id": "12"
}'
```

#### Output
```json
{
	"jsonrpc": "2.0",
	"result": {
		"Data": "hash"
	},
	"id": "12"
}
```

After reshare, the value of new share is rotated and different with the old one and each share in the output will be stored and replaced in DB.

## Build

To build the TSS binary, run the following command in the project root directory:

```shell
make tss
```

This will create an executable at `./cmd/tss`.

### Prerequisites

- Go 1.19 or later
- Make sure to initialize git submodules if this is a fresh clone:
  ```shell
  make init
  ```

## Usage

After building the binary, you need to start multiple nodes to participate in the TSS protocol. Here's how to set up a 3-node cluster:

### Node Setup

Each node requires:
- A unique configuration file
- A keystore file for the node's identity
- A password for the keystore (can be provided via flag or prompt)

### Starting the Nodes

**Node A** (using provided config):
```yaml
# config/id-10001-input.yml
port: 10001
rpc: 1234

store:
  type: "badger"
  path: "./node.test/badger1"
```
```shell
./cmd/tss --config ./config/id-10001-input.yml --keystore ./node.test/keystore/1 --password <password>
```

**Node B** (using provided config):
```yaml
# config/id-10002-input.yml
port: 10002
rpc: 1235

store:
  type: "badger"
  path: "./node.test/badger2"
```
```shell
./cmd/tss --config ./config/id-10002-input.yml --keystore ./node.test/keystore/2 --password <password>
```

**Node C** (using provided config):
```yaml
# config/id-10003-input.yml
port: 10003
rpc: 1236

store:
  type: "badger"
  path: "./node.test/badger3"
```
```shell
./cmd/tss --config ./config/id-10003-input.yml --keystore ./node.test/keystore/3 --password <password>
```

### Command Line Options

- `--config`: Path to the configuration file
- `--keystore`: Path to the keystore file for node identity
- `--password`: Password for the keystore file
- `--port`: Override the RPC port from config
- `--self-host`: Run in self-hosted mode (disables mDNS discovery)

### Network Discovery

The nodes use mDNS (multicast DNS) for automatic peer discovery on the local network. Make sure all nodes are running on the same network segment for automatic discovery to work.

## Creating Keystore Files

Before starting nodes, you need to create keystore files for each node's identity. You can create these using standard Ethereum keystore tools or generate them programmatically using the project's utilities.
