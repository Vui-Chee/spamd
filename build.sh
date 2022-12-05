#!/usr/bin/env sh

env GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o spamd_darwin_arm64
env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o spamd_darwin_amd64
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o spamd_linux_amd64
env GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o spamd_linux_arm
env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o spamd_windows_amd64.exe
