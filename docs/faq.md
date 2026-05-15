# FAQ

## How can I report a bug or get help?
You can ask for help or report an issue at [https://github.com/in4it/vpn-server](https://github.com/in4it/vpn-server).


## How can I check whether the VPN works?
You can ping the VPN server using the VPN IP address. In a terminal, enter the command:

```
ping 10.189.184.1
```

## How can I route specific subnets?
Use the ClientRoutes setting on the VPN settings page to specify the default route the client should use.

## What is the default IP range used for the VPN?
The default IP range is 10.189.184.0/21. The VPN Server will always use the first non-network IP address in the range, which is 10.189.184.1. If you want to change this IP range, you can edit the configuration files directly. The IP range is defined in `/vpn/config/vpn-config.json`. We aim to make all configuration parameters available as options in the admin UI, but this is not the case yet.

When making changes to the client configuration files, make sure to restart the VPN using `systemctl restart vpn-configmanager` and `systemctl restart vpn-rest-server`.

When making changes to the server configuration, use `wg-quick down vpn` to first bring down the existing VPN, then run `systemctl restart vpn-configmanager` and `systemctl restart vpn-rest-server`.


## Where can I make changes to the VPN Server or Client configuration file?
You can find the client and server configuration file templates in `/vpn/config/templates/`. After editing the files, make sure to restart the VPN using `systemctl restart vpn-configmanager` and `systemctl restart vpn-rest-server`.

## The Copy button does not work on the Authentication & Provisioning page
The copy feature only works if the VPN Server is using HTTPS, as browsers only allow clipboard access in a secure context. OpenID Connect (OIDC) callback URLs also often have to use HTTPS. If you intend to use the Authentication & Provisioning features, enable TLS (HTTPS) on the VPN setup page.
