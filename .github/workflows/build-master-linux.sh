#!/bin/bash
set -e

echo "- Install Go"
apt update
apt install --no-install-recommends -y curl ca-certificates
curl -o go.tar.gz "https://dl.google.com/go/$(curl https://go.dev/VERSION?m=text | head -n1).linux-amd64.tar.gz"
tar -xzf "go.tar.gz"
export PATH="$PATH:$(realpath ./go/bin)"
go version

echo "- Install dependencies"
cat << EOF > /etc/apt/sources.list
deb [arch=amd64,i386] http://archive.ubuntu.com/ubuntu/ bionic main universe
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ bionic main universe
deb [arch=amd64,i386] http://archive.ubuntu.com/ubuntu/ bionic-updates main universe
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ bionic-updates main universe
deb [arch=amd64,i386] http://security.ubuntu.com/ubuntu/ bionic-security main universe
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ bionic-security main universe
EOF
dpkg --add-architecture i386
dpkg --add-architecture armhf
dpkg --add-architecture arm64
apt update
apt install --no-install-recommends -y \
  ca-certificates zip \
  build-essential \
  gcc-7-multilib g++-7-multilib \
  gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf \
  gcc-aarch64-linux-gnu g++-aarch64-linux-gnu

echo "- Build CLI"
CGO_ENABLED=0 GOARCH=amd64 go build -o ./build/x86-64/dsl ./cmd
CGO_ENABLED=0 GOARCH=386 go build -o ./build/x86/dsl ./cmd
CGO_ENABLED=0 GOARCH=arm GOARM=6 go build -o ./build/arm/dsl ./cmd
CGO_ENABLED=0 GOARCH=arm64 go build -o ./build/arm64/dsl ./cmd

echo "- Build GUI (x86-64)"
apt install --no-install-recommends -y libgtk-3-dev libwebkit2gtk-4.0-dev
CGO_ENABLED=1 \
GOARCH=amd64 \
  go build -tags gui -o ./build/x86-64/dsl-gui ./cmd
apt remove -y libgtk-3-dev libwebkit2gtk-4.0-dev


echo "- Build GUI (x86)"
apt install --no-install-recommends -y libgtk-3-dev:i386 libwebkit2gtk-4.0-dev:i386
PKG_CONFIG_PATH=/usr/lib/i386-linux-gnu/pkgconfig \
CGO_ENABLED=1 \
GOARCH=386 \
  go build -tags gui -o ./build/x86/dsl-gui ./cmd
apt remove -y libgtk-3-dev:i386 libwebkit2gtk-4.0-dev:i386

echo "- Build GUI (arm)"
apt install --no-install-recommends -y libgtk-3-dev:armhf libwebkit2gtk-4.0-dev:armhf
CC=arm-linux-gnueabihf-gcc \
CXX=arm-linux-gnueabihf-g++ \
PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabihf/pkgconfig \
CGO_ENABLED=1 \
GOARCH=arm \
GOARM=6 \
  go build -tags gui -o ./build/arm/dsl-gui ./cmd
apt remove -y libgtk-3-dev:armhf libwebkit2gtk-4.0-dev:armhf

echo "- Build GUI (arm64)"
apt install --no-install-recommends -y libgtk-3-dev:arm64 libwebkit2gtk-4.0-dev:arm64
CC=aarch64-linux-gnu-gcc \
CXX=aarch64-linux-gnu-g++ \
PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig \
CGO_ENABLED=1 \
GOARCH=arm64 \
  go build -tags gui -o ./build/arm64/dsl-gui ./cmd
apt remove -y libgtk-3-dev:arm64 libwebkit2gtk-4.0-dev:arm64
