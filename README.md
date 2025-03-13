## Pong on Bison Relay

A Pong implementation with betting capabilities, built on top of Bison Relay's privacy-preserving network platform.

### Features

 - 🏓 Real-time Pong gameplay with terminal-based and flutter UI
 - 💰 Betting system with DCR transactions
 - 🚦 Matchmaking system with waiting rooms
 - 🔔 In-game notifications system

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

### UI Client (Flutter)
```bash
# Navigate to the Flutter UI directory
cd pongui
```

For detailed information about building the Flutter UI, please see the [Flutter UI Documentation](pongui/README.md).

### Bot
```bash
go build -o pongbot ./cmd/pongbot
```

### Terminal Client
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

When running the client (Flutter app or pongclient) for the first time, it will create a new configuration file with default values. These default values should match the RPC configuration of your bruig/brclient setup.

The client config is located at: `{appdata}/.pongui/pongui.conf`

```ini
serveraddr={server_ip_address}:50051
rpcurl=wss://127.0.0.1:7676/ws
servercertpath=/home/{user}/.brclient/rpc.cert
clientcertpath=/home/{user}/.brclient/rpc-client.cert
clientkeypath=/home/{user}/.brclient/rpc-client.key
grpcservercert=/home/{user}/server.cert
rpcuser=whatever_username_you_want
rpcpass=some_strong_password
```

When running the bot for the first time it will create a conf file with default values, which might need to be adjusted.

The bot config is located at: `{appdata}/.pongbot/pongbot.conf`

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

Same for the client: `{appdata}/.pongclient/pongclient.conf`

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

## Running

Start the pongbot server or connect to an existing one.

```bash
./pongbot
```

Start a client instance (terminal UI)

```bash
./pongclient
```

Start the Flutter UI client

```bash
# Navigate to the Flutter UI directory
cd pongui
```

## Gameplay

1. First, you must send a tip to the bot to establish your bet amount (in DCR)
2. After tipping, you can create or join a waiting room
3. You can only join waiting rooms with the same bet amount as your tip
4. In the waiting room, you can:
   - Get ready/unready
   - Leave the waiting room
5. When both players are ready, the game starts automatically
6. Play using W/S or arrow keys (Up/Down)
7. First player to score 3 points wins the match
8. Winner takes all bets

## ⚠️ Warning

- **Ensure Channel Liquidity**: Make sure your client has enough liquidity to receive the bet amount if you win
- **Check Channel Status**: You can verify your channel liquidity in the Bison Relay application under the Network tab, in the Channels section
- **Minimum Bet**: The minimum bet is determined by the bot configuration (default: 0.00000001 DCR | 1 atom)
- **Connection Issues**: Ensure your Bison Relay client is properly connected before starting a game

## Betting System

- **Configurable Minimum Bet**  
  `-minbetamt` flag sets minimum wager (default: 0.00000001 DCR)
  
- **Free-to-Play Mode**  
  `-isf2p=true` enables simulated bets without real funds (disabled by default)

- **Secure Payment Handling**  
  Bot processes transactions through Bison Relay's RPC client

## gRPC API

For detailed information about the gRPC API, please see the [gRPC API Documentation](pongrpc/README.md).


