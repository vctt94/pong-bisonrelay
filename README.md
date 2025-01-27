## Pong on Bison Relay

A Pong implementation with betting capabilities, built on top of Bison Relay's privacy-preserving network platform.

### Features

🏓 Real-time Pong gameplay with terminal-based and flutter UI
💰 Betting system with DCR transactions
🚦 Matchmaking system with waiting rooms
🔔 In-game notifications system

### System Architecture:
- gRPC API handles game state synchronization
- Bison Relay RPC Client handles payment transactions
- Two UI options
  - Terminal UI (BubbleTea)
  - Flutter UI

## Getting Started

```
Prerequisites
Go 1.22+
Bison Relay client (with RPC configured)
Flutter for UI
```

## Build

### Bot
```bash
go build -o pongbot ./cmd/pongbot
```

### Client
```bash
go build -o pongclient ./cmd/pongclient
```

## Configuration

Ensure your Bison Relay client configuration (brclient.conf or bruig.conf) contains:
```ini
[clientrpc]
jsonrpclisten = 127.0.0.1:7676
rpccertpath = /home/{user}/.brclient/rpc.cert
rpckeypath = /home/{user}/.brclient/rpc.key
rpcuser = whatever_username_you_want
rpcpass = some_strong_password
rpcauthmode = basic
rpcclientcapath = /home/{user}/.brclient/rpc-ca.cert
rpcissueclientcert = 1
```

when running the bot for the first time it will create a conf file with default values, which might need to be adjusted.

The one created is located at: `{appdata}/pongbot.conf`

```ini
datadir=/home/{user}/.pongbot
isf2p=false
minbetamt=0.00000001
rpcurl=wss://127.0.0.1:7676/ws
grpchost=localhost
grpcport=50051
httpport=8888
servercertpath=/home/{user}/.brclient/rpc.cert
clientcertpath=/home/{user}/.brclient/rpc-client.cert
clientkeypath=/home/{user}/.brclient/rpc-client.key
rpcuser=whatever_username_you_want
rpcpass=some_strong_password
debug=debug
```

Same for the client: `{appdata}/pongclient.conf`

```ini
serveraddr=localhost:50051
rpcurl=wss://127.0.0.1:7676/ws
servercertpath=/home/{user}/.brclient/rpc.cert
clientcertpath=/home/{user}/.brclient/rpc-client.cert
clientkeypath=/home/{user}/.brclient/rpc-client.key
grpcservercert=/home/{user}/server.cert
rpcuser=whatever_username_you_want
rpcpass=some_strong_password
```

Running

Start the pongbot server or connect to an existing one.

```bash
./pongbot
```

Start a client instance

```bash
./pongclient
```

Gameplay

```
1. Create or join a waiting room
2. Send tip to bot to set bet amount (DCR)
3. Get Ready (space key)
4. Wait for opponent
5. Play using W/S or arrow keys
6. Winner takes all
```

gRPC API

Key endpoints

 - SendInput:          Player Inputs 
 - StartGameStream:    Update game state stream
 - StartNtfnStream:    Notifications stream
 - GetWaitingRoom:     Single room details
 - GetWaitingRooms:    All available rooms list
 - CreateWaitingRoom:  New Waiting Room
 - JoinWaitingRoom:    Join existing room


## Betting System

- **Configurable Minimum Bet**  
  `-minbetamt` flag sets minimum wager (default: 0.00000001 DCR)
  
- **Free-to-Play Mode**  
  `-isf2p=true` enables simulated bets without real funds (disabled by default)

- **Secure Payment Handling**  
  Bot processes transactions through Bison Relay's RPC client
