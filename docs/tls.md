# TLS

## Configuration

You can enable TLS (https) in the VPN Settings. TLS only works if you have a hostname configured as the "VPN Server Hostname". Make sure you have created a DNS record like vpn.yourcompany.com to the IP address of the VM instance. Once you enable the TLS setting, let's encrypt will be activated. An API call will be made to [letsencrypt.com](https://letsencrypt.org/), which will then make an HTTP call on your hostname to verify ownership. Only when this call succeeds, the TLS certificate will be issued, and the VPN Server will be accessible over https.

## http to https forward
Make sure to only enable the http to https forwarding when https is fully working. If you enabled the http to https forwarding, but can't access the VPN Server over https, you can still disable the forwarding manually.

Log in using SSH to the VPN Server and cd into the /vpn/config directory. The config.json file contains an attribute `redirectToHttps` that will be set to `true`. You can either remove the attribute and value or set the value to false. Also make sure that the attribute `protocol` is set back to `http` instead of `https`. Restart the VPN server using `systemctl restart vpn-rest-server`.

## Alternatives
On Cloud providers like AWS a Load Balancer can be created to ensure access between the client and the AWS Load Balancer is using TLS.

## VPN Traffic
VPN Traffic between client and VPN Server using WireGuard® is always encrypted. The TLS solution using Let's Encrypt is only to encrypt web traffic between the client (the browser) and the VPN Server Admin Web Interface.