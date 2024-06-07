## Classic Pong Game For Bison Relay
This project implements a server for a classic Pong game, allowing two players to compete against each other in real-time. It's built using Go and utilizes gRPC for handling client-server communication. In addition, the server establishes a connection with the Bison Relay client, facilitating secure and efficient communication between the game server and clients. This setup ensures not only the management of game instances and player inputs but also the seamless streaming of game updates to clients through the Bison Relay infrastructure.

## Features
  - Real-time Multiplayer: Supports two players competing against each other.
  - Betting System: Players can place bets in Decred (DCR), and the winner takes the pot.
  - Secure Communication: Utilizes Bison Relay for secure and decentralized communication.
  - Go and gRPC: Built using Go and gRPC for efficient client-server communication.

## Prerequisites

  - Go 1.20 or higher
  - Decred wallet with some DCR
  - Bison Relay account

## Installation

Clone the Repository:

```
git clone https://github.com/yourusername/pong-bisonrelay.
cd pong-bisonrelay
```

Install Dependencies:

```
go mod tidy
```

Configure Bison Relay:

  Ensure you have a Bison Relay account set up. Configure the brclient.conf file for JSON-RPC, Example:

```
  [clientrpc]
  # Enable the JSON-RPC clientrpc protocol on the comma-separated list of addresses.
  jsonrpclisten = 127.0.0.1:7676

  # Path to the keypair used for running TLS on the clientrpc interfaces.
  rpccertpath = ~/.brclient/rpc.cert
  rpckeypath = ~/.brclient/rpc.key

  # Path to the certificate used as CA for client-side TLS authentication.
  rpcclientcapath = ~/.brclient/rpc-ca.cert

  # If set to true, generate the rpc-client.cert and rpc-client.key files in the
  # same dir as rpcclientcapath, that should be specified by a client connecting
  # over the clientrpc interfaces. If set to false, then the user is responsible
  # for generating the client CA, and cert files.
  rpcissueclientcert = true
```


### Controls

Move Up: W, Up Arrow
Move Down: S, Down Arrow
