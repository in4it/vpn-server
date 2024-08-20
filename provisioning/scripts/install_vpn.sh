#!/bin/bash
ARCHITECTURE=$(uname -m)
if [ -e "/usr/bin/cloud-init" ] ; then
    cloud-init status --wait
fi
apt-get -y update
apt-get -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' full-upgrade
apt-get install -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' wireguard
echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf
sysctl net.ipv4.ip_forward=1

mkdir -p /vpn
groupadd vpn
useradd -d /vpn -s /usr/sbin/nologin -g vpn vpn
mv /tmp/restserver-linux-${ARCHITECTURE} /vpn/rest-server
mv /tmp/reset-admin-password-linux-${ARCHITECTURE} /vpn/reset-admin-password
mv /tmp/configmanager-linux-${ARCHITECTURE} /vpn/configmanager

chown -R vpn:vpn /vpn
chmod 700 /vpn
chmod 700 /vpn/rest-server
chmod 700 /vpn/reset-admin-password
chmod 700 /vpn/configmanager

setcap 'cap_net_bind_service=+ep' /vpn/rest-server

mv /tmp/vpn-configmanager.service /etc/systemd/system/vpn-configmanager.service
mv /tmp/vpn-rest-server.service /etc/systemd/system/vpn-rest-server.service

systemctl enable vpn-configmanager
systemctl enable vpn-rest-server

apt-get clean