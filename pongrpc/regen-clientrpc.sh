#!/bin/sh

BINDIR=$(mktemp -d)

build_protoc_gen_go() {
    mkdir -p $BINDIR
    export GOBIN=$BINDIR
    go install github.com/golang/protobuf/protoc-gen-go
}

generate() {
    protoc --go_out=. --go-grpc_out=. pong.proto
    protoc --dart_out=grpc:../pongui/flutterui/pongui/lib/grpc/generated -I. pong.proto
}

# Build the bins from the main module, so that clientrpc doesn't need to
# require all client and tool dependencies.
(cd .. && build_protoc_gen_go)
GENPATH="$BINDIR:$PATH"
PATH=$GENPATH generate
