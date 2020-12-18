#!/bin/bash

pushd scripts
node parser.js
popd

gofmt -w appscan/source.go

rm -rf build/*

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o build/netscan netscan.go
#CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o build/portscan.exe portscan/cli/cli.go

