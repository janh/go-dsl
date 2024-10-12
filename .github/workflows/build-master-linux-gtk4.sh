#!/bin/bash
set -e

echo "- Install Go"
curl -o go.tar.gz "https://dl.google.com/go/$(curl https://go.dev/VERSION?m=text | head -n1).linux-amd64.tar.gz"
tar -xzf "go.tar.gz"
export PATH="$PATH:$(realpath ./go/bin)"
go version

echo "- Enter workspace"
cd /workspace

echo "- Build CLI"
CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o ./build/x86-64/dsl ./cmd
CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o ./build/x86/dsl ./cmd
CGO_ENABLED=0 GOARCH=arm GOARM=6 go build -ldflags "-s -w" -o ./build/arm/dsl ./cmd
CGO_ENABLED=0 GOARCH=arm64 go build -ldflags "-s -w" -o ./build/arm64/dsl ./cmd

echo "- Patch webview package to use GTK4"
git clone https://github.com/webview/webview_go.git /tmp/webview_go
cd /tmp/webview_go
git checkout 6173450d4dd61511002d897d55e4d0b6e75aeb96
git apply /workspace/.github/workflows/webview-go-gtk4.patch
cd /workspace
go mod edit -replace github.com/webview/webview_go=/tmp/webview_go
go mod tidy

echo "- Build GUI (x86-64)"
CGO_ENABLED=1 \
GOARCH=amd64 \
  go build -tags gui -ldflags "-s -w" -o ./build/x86-64/dsl-gui ./cmd

echo "- Build GUI (x86)"
PKG_CONFIG_PATH=/usr/lib/i386-linux-gnu/pkgconfig \
CGO_ENABLED=1 \
GOARCH=386 \
  go build -tags gui -ldflags "-s -w" -o ./build/x86/dsl-gui ./cmd

echo "- Build GUI (arm)"
CC=arm-linux-gnueabihf-gcc \
CXX=arm-linux-gnueabihf-g++ \
PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabihf/pkgconfig \
CGO_ENABLED=1 \
GOARCH=arm \
GOARM=6 \
  go build -tags gui -ldflags "-s -w" -o ./build/arm/dsl-gui ./cmd

echo "- Build GUI (arm64)"
CC=aarch64-linux-gnu-gcc \
CXX=aarch64-linux-gnu-g++ \
PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig \
CGO_ENABLED=1 \
GOARCH=arm64 \
  go build -tags gui -ldflags "-s -w" -o ./build/arm64/dsl-gui ./cmd
