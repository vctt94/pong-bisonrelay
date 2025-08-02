# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

### Server (Bot)
```bash
go build -o pongbot ./cmd/pongbot
```

### Terminal Client
```bash
go build -o pongclient ./cmd/pongclient
```

### Flutter UI
```bash
# Native desktop
flutter build linux
flutter build macos  
flutter build windows

# Cross-platform build
flutter build android
flutter build ios

# Development
flutter run
```

### Database CLI
```bash
go build -o server_db_cli ./cmd/server_db_cli
```

### Development Scripts
Use provided scripts in `/scripts/` for development workflow:
- `dev-script.sh` - Sets up tmux session with dcrd, dcrwallet, dcrlnd, brserver
- `bot-script.sh` - Launch bot + brclient in tmux
- `pongui-script.sh` - Launch Flutter UI

## Testing
```bash
go test ./...
go test ./ponggame -v
go test ./server -v
flutter test  # for Flutter UI
```

## Architecture Overview

### System Components
- **Pong Game Engine** (`ponggame/`) - Core game logic, physics, waiting rooms
- **gRPC API** (`pongrpc/`) - Game state synchronization via protobuf-defined services
- **Server** (`server/`) - Bot managing games, betting, and user sessions
- **Clients** - Terminal UI (`cmd/pongclient/`) and Flutter UI (`pongui/`)
- **Go bindings for Flutter** (`golib/`) - Native Go integration with Flutter via ffi

### Key Architecture Details
- **Game Sessions**: Per-player frame channels via gRPC streaming
- **Betting System**: DCR payments via Bison Relay RPC, minimum configurable bet
- **Matchmaking**: Waiting rooms with player ready/unready states
- **Real-time Sync**: Game state updates push via gRPC streams
- **Persistence**: BoltDB for game/player data, UUID-based game IDs

### Core Data Flow
1. User tips bot to establish bet amount → coinflip notification
2. User joins/creates waiting room → broadcast to players
3. Players ready up → waiting room → game instance created
4. Lock-in bets → coinflip determines server/client roles
5. Game frames streamed via gRPC → real-time state sync
6. Game ends → winner takes all via automated payouts

### Required Dependencies
- Go 1.23.4+ (check go.mod for exact version)
- Flutter SDK for UI development
- Bison Relay client (brclient/bruig) with RPC configured
- Decred blockchain stack (dcrd, dcrwallet, dcrlnd)

### Configuration Files
- **Bot**: `~/.pongbot/pongbot.conf` (auto-created)
- **Terminal Client**: `~/.pongclient/pongclient.conf` (auto-created)  
- **Flutter UI**: `~/.pongui/pongui.conf` (auto-created)
- **Bison Relay**: Must have RPC enabled (see README.md)