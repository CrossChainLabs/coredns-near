#!/bin/bash
set -e

SRCDIR=`pwd`
BUILDDIR=`pwd`/build

mkdir -p ${BUILDDIR} 2>/dev/null
cd ${BUILDDIR}
echo "Cloning coredns repo..."
export GOPRIVATE=github.com/CrossChainLabs/coredns-simple
git clone https://github.com/coredns/coredns.git

cd coredns
git checkout v1.8.3

echo "Building..."
make

echo "Patching plugin config..."
ed plugin.cfg <<EOED
/rewrite:rewrite
a
near:github.com/CrossChainLabs/coredns-simple
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
rm -r ${BUILDDIR}
