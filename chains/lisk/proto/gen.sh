#!/bin/sh

go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null
go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc 2>/dev/null

protoc -I=. --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. *.proto
