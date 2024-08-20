# VPN Server

## Usage
You can launch the WireGuard® based VPN server for production use from the [AWS Marketplace](https://aws.amazon.com/marketplace/pp/prodview-dymnyb6a2pq72), the [Azure Marketplace](
https://azuremarketplace.microsoft.com/en-us/marketplace/apps/in4it.vpn-server), the [DigitalOcean Marketplace](https://marketplace.digitalocean.com/apps/vpn-server), the [GCP Marketplace](https://console.cloud.google.com/marketplace/product/in4it-public/vpn-server) or install the VPN manually. Personal use is allowed under BSL (Business Source License).

## Features
* Easy to use admin UI
* SAML, OpenID Connect, SCIM support
* WireGuard® as VPN technology, a fast and modern VPN Solution

## Bugs or Issues
Use the GitHub Issues to report any bugs or issues. We are monitoring new issues and will respond in a timely matter.

## Manual install
You can install the VPN Server manually if you can't use one of the cloud marketplace options. Make sure you have an Ubuntu 24.04 Linux instance running with open ports 80/tcp, 443/tcp, and 51820/udp. Clone this repository, then run:

```
make
mv restserver-linux-amd64 /tmp           # you can change amd64 to arm64 if you are on arm64
mv reset-admin-password-linux-amd64 /tmp
mv configmanager-linux-amd64 /tmp
provisioning/scripts/install_vpn.sh
```

You can now start the VPN server using the following commands:
```
systemctl enable vpn-configmanager
systemctl enable vpn-rest-server
```

The VPN Server admin frontend should be available at `http://<ip of instance>`