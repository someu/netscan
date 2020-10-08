#!/bin/bash

#go run cmd/* scan -i 113.160.0.0/24 -p 80

rm build/* -rf

cp -r sources build
go build -o build/fakescan cmd/*
go build -o build/fakescan_server server/*