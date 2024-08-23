#!/bin/sh
export LATEST=`cat ../latest`
aws s3 cp ../restserver-linux-amd64 s3://in4it-vpn-server/assets/binaries/${LATEST}/restserver-linux-amd64
aws s3 cp ../restserver-linux-amd64.sha256 s3://in4it-vpn-server/assets/binaries/${LATEST}/restserver-linux-amd64.sha256
aws s3 cp ../reset-admin-password-linux-amd64 s3://in4it-vpn-server/assets/binaries/${LATEST}/reset-admin-password-linux-amd64
aws s3 cp ../reset-admin-password-linux-amd64.sha256 s3://in4it-vpn-server/assets/binaries/${LATEST}/reset-admin-password-linux-amd64.sha256
aws s3 cp ../configmanager-linux-amd64 s3://in4it-vpn-server/assets/binaries/${LATEST}/configmanager-linux-amd64
aws s3 cp ../configmanager-linux-amd64.sha256 s3://in4it-vpn-server/assets/binaries/${LATEST}/configmanager-linux-amd64.sha256
if [ "$1" == "--release" ] ; then
	echo "=> $LATEST released."
	#aws s3 cp ../latest s3://in4it-vpn-server/assets/binaries/latest
fi
