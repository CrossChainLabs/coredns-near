#!/bin/bash
set -e

SRCDIR=`pwd`
BUILDDIR=`pwd`/build

mkdir -p ${BUILDDIR} 2>/dev/null
cd ${BUILDDIR}
rm -rf coredns 2>/dev/null

echo "Cloning coredns repo..."
git clone https://github.com/coredns/coredns.git

cd coredns
git checkout v1.8.3

go get github.com/CrossChainLabs/coredns-near

echo "Building..."
make

echo "Patching plugin config..."
ed plugin.cfg <<EOED
/rewrite:rewrite
a
near:github.com/CrossChainLabs/coredns-near
.
w
q
EOED

echo "Building with plugin..."
#make
make SHELL='sh -x' CGO_ENABLED=1 coredns

cp coredns ${SRCDIR}
#chmod -R 755 .git
cd ${SRCDIR}
