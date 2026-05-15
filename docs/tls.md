# TLS

## Configuration

You can enable TLS (HTTPS) in the VPN Settings. TLS only works if you have configured a hostname as the "VPN Server Hostname". Make sure you have created a DNS record, such as `vpn.yourcompany.com`, that points to the IP address of the VM instance. Once you enable the TLS setting, Let's Encrypt will be activated. An API call will be made to [letsencrypt.org](https://letsencrypt.org/), which will then make an HTTP call to your hostname to verify ownership. The TLS certificate will only be issued when this call succeeds, and the VPN Server will then be accessible over HTTPS.

## HTTP to HTTPS Forwarding
Only enable HTTP to HTTPS forwarding when HTTPS is fully working. If you enabled HTTP to HTTPS forwarding but cannot access the VPN Server over HTTPS, you can still disable the forwarding manually.

Log in to the VPN Server using SSH and change into the `/vpn/config` directory. The `config.json` file contains a `redirectToHttps` attribute that will be set to `true`. You can either remove the attribute and value or set the value to `false`. Also make sure the `protocol` attribute is set back to `http` instead of `https`. Restart the VPN server using `systemctl restart vpn-rest-server`.

## Alternatives
On cloud providers like AWS, you can create a load balancer to ensure traffic between the client and the AWS load balancer uses TLS.

## VPN Traffic
VPN traffic between the client and VPN Server using WireGuard® is always encrypted. The TLS solution using Let's Encrypt only encrypts web traffic between the client browser and the VPN Server admin web interface.
