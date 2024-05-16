# Upgrade

# Automated upgrade

An automatic upgrade procedure within the UI is available from v1.0.36 and onwards. If a new version is available, a banner will show on the status page with a link to perform the upgrade.

# Manual upgrade
Execute the following commands using SSH to upgrade the VPN Server

```
VPN_SERVER_VERSION=$(curl -s https://in4it-vpn-server.s3.amazonaws.com/assets/binaries/latest)
cd /vpn
rm rest-server reset-admin-password configmanager
curl -o rest-server https://in4it-vpn-server.s3.amazonaws.com/assets/binaries/${VPN_SERVER_VERSION}/restserver-linux-amd64
curl -o reset-admin-password https://in4it-vpn-server.s3.amazonaws.com/assets/binaries/${VPN_SERVER_VERSION}/reset-admin-password-linux-amd64
curl -o configmanager https://in4it-vpn-server.s3.amazonaws.com/assets/binaries/${VPN_SERVER_VERSION}/configmanager-linux-amd64
chown vpn:vpn rest-server reset-admin-password configmanager
chmod 700 rest-server reset-admin-password configmanager
setcap 'cap_net_bind_service=+ep' rest-server
systemctl restart vpn-configmanager
systemctl restart vpn-rest-server
```

# Revert to older version

Run the same commands as above, but specify a version instead:
```
VPN_SERVER_VERSION=v1.x.x
```