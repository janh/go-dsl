#!/bin/bash
set -e

echo "- Creating temporary build directory"
mkdir /tmp/build
cd /tmp/build

echo "- Set up multiarch"
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
  libgdk-pixbuf2.0-dev (=2.30.0),
  libpango1.0-dev (=1.40.5),
  libsoup2.4-dev (=2.40),
  libicu-dev
EOF
equivs-build dummy
dpkg -i dummy_1.0_all.deb

echo "- Install build tools"
apt install --no-install-recommends -y \
  gcc g++ \
  gcc-7-multilib g++-7-multilib \
  gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf \
  gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
  curl ca-certificates \
  git
git config --system --add safe.directory '*'

echo "- Install libraries"
apt install --no-install-recommends -y \
  libgtk-3-dev libwebkit2gtk-4.0-dev \
  libgtk-3-dev:i386 libwebkit2gtk-4.0-dev:i386 \
  libgtk-3-dev:armhf libwebkit2gtk-4.0-dev:armhf \
  libgtk-3-dev:arm64 libwebkit2gtk-4.0-dev:arm64

echo "- Manually install conflicting packages"
apt install --no-install-recommends -y \
  libharfbuzz-dev libharfbuzz-dev:i386 libharfbuzz-dev:armhf libharfbuzz-dev:arm64 \
  libxml2-dev libxml2-dev:i386 libxml2-dev:armhf libxml2-dev:arm64
apt download \
  libgdk-pixbuf2.0-dev libgdk-pixbuf2.0-dev:i386 libgdk-pixbuf2.0-dev:armhf libgdk-pixbuf2.0-dev:arm64 \
  libpango1.0-dev libpango1.0-dev:i386 libpango1.0-dev:armhf libpango1.0-dev:arm64 \
  libsoup2.4-dev libsoup2.4-dev:i386 libsoup2.4-dev:armhf libsoup2.4-dev:arm64 \
  libicu-dev libicu-dev:i386 libicu-dev:armhf libicu-dev:arm64
find . -name "*.deb" -print | xargs -i dpkg -x {} /

echo "- Cleaning up"
rm -r /tmp/build
apt clean
