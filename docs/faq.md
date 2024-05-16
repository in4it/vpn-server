# FAQ

## How can I check whether the VPN works

You can ping the VPN server using the VPN IP address. In a terminal, enter the command:

```
ping 10.189.184.1
```

## I want to route specific subnets?
Use the ClientRoutes setting on the VPN settings page to specify the default route the client should use.

## What is the default IP range used for the VPN?
The default IP range used is 10.189.184.0/21. The VPN Server will always be on the first (non-network) IP address or the range, so 10.189.184.1. If you want to change this IP range, you can edit the configuration files directly. The IP range is defined in /vpn/config/vpn-config.json. We're aiming to have all configuration parameters als options in the admin UI, but this is not the case yet.

When making changes to the client configuration files, make sure to restart the VPN using `systemctl restart vpn-configmanager` and `systemctl restart vpn-rest-server`.

When making changes to the server configuration, use `wg-quick down vpn` to first bring down the existing VPN, then run `systemctl restart vpn-configmanager` and `systemctl restart vpn-rest-server`.


## Where can I make changes to the VPN Server or Client configuration file?
You can find the client and server configuration file template in `/vpn/config/templates/`. After editing the files, make sure to restart the VPN using `systemctl restart vpn-configmanager` and `systemctl restart vpn-rest-server`.

## The Copy button does not work on the Authentication & Provisioning page
* The copy feature only works if the VPN Server is using https, as browsers only allow to copy something to the clipboard in a secure context. Also OpenID Connect (OIDC) callback URLs often have to be https only. If you intend to use the Authentication & Provisioning features, enable TLS (https) on the VPN setup page.