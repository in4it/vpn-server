#!/bin/sh
export LATEST=`cat ../latest`
for ARCH in amd64 arm64; do
	for BINARY in restserver reset-admin-password configmanager; do
		aws s3 cp ../${BINARY}-linux-${ARCH} s3://in4it-vpn-server/assets/binaries/${LATEST}/${BINARY}-linux-${ARCH}
		aws s3 cp ../${BINARY}-linux-${ARCH}.sha256 s3://in4it-vpn-server/assets/binaries/${LATEST}/${BINARY}-linux-${ARCH}.sha256
	done
done
if [ "$1" = "--release" ] ; then
	echo "=> $LATEST released."
	aws s3 cp ../latest s3://in4it-vpn-server/assets/binaries/latest
fi
