#!/bin/bash

cd frontend || exit
bun install
bun run build

cd ..
go build -o build/proxy-manager main.go
