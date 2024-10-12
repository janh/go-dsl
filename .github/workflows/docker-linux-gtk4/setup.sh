#!/bin/bash
set -e

echo "- Creating temporary build directory"
mkdir /tmp/build
cd /tmp/build

echo "- Set up multiarch"
cat << EOF > /etc/apt/sources.list
deb [arch=amd64,i386] http://archive.ubuntu.com/ubuntu/ jammy main universe
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy main universe
deb [arch=amd64,i386] http://archive.ubuntu.com/ubuntu/ jammy-updates main universe
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy-updates main universe
deb [arch=amd64,i386] http://security.ubuntu.com/ubuntu/ jammy-security main universe
deb [arch=armhf,arm64] http://ports.ubuntu.com/ubuntu-ports/ jammy-security main universe
EOF
dpkg --add-architecture i386
dpkg --add-architecture armhf
dpkg --add-architecture arm64
apt update

echo "- Install prerequisites"
apt install --no-install-recommends -y \
  equivs

echo "- Set up hack for multi-arch package conflicts"
cat << EOF > dummy
Package: dummy
Version: 1.0
Architecture: all
Multi-Arch: foreign
Provides:
  gobject-introspection,
  python3-mako,
  python3-markdown,
  build-essential,
  libgraphene-1.0-dev (=1.9.1),
  libnghttp2-dev,
  libpolkit-gobject-1-dev (=0.105)
EOF
equivs-build dummy
dpkg -i dummy_1.0_all.deb

echo "- Install build tools"
apt install --no-install-recommends -y \
  gcc g++ \
  gcc-11-multilib g++-11-multilib \
  gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf \
  gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
  curl ca-certificates \
  git
git config --system --add safe.directory '*'

echo "- Install libraries"
apt install --no-install-recommends -y \
  libgtk-4-dev libwebkitgtk-6.0-dev \
  libgtk-4-dev:i386 libwebkitgtk-6.0-dev:i386 \
  libgtk-4-dev:armhf libwebkitgtk-6.0-dev:armhf \
  libgtk-4-dev:arm64 libwebkitgtk-6.0-dev:arm64

echo "- Manually install conflicting packages"
apt download \
  libgraphene-1.0-dev libgraphene-1.0-dev:i386 libgraphene-1.0-dev:armhf libgraphene-1.0-dev:arm64 \
  libnghttp2-dev libnghttp2-dev:i386 libnghttp2-dev:armhf libnghttp2-dev:arm64 \
  libpolkit-gobject-1-dev libpolkit-gobject-1-dev:i386 libpolkit-gobject-1-dev:armhf libpolkit-gobject-1-dev:arm64
find . -name "*.deb" -print | xargs -i dpkg -x {} /

echo "- Cleaning up"
rm -r /tmp/build
apt clean
