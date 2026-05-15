# Quick Start

## Log In

The first user created is the `admin` user. You can log in using the `admin` username and the `admin` password. If you forget the password, log in to the server using SSH and run `sudo /vpn/reset-admin-password`.

## Create the First User

You can create a first user on the user page. The `admin` user cannot create VPN connections.

* Create a new user on the user page
* Log out and log in using the new credentials
* Create a new connection on the Connections page
* Download the configuration
* Import the configuration into a WireGuard® client. See [https://www.wireguard.com/install/](https://www.wireguard.com/install/) for WireGuard® clients.

## Access for Existing Users in Identity Providers

To allow access for users created in Active Directory, Okta, OneLogin, or other identity providers, navigate to the [OIDC](oidc.md), [SAML](saml.md), or [Provisioning (SCIM)](scim.md) pages to set up an IdP connection.
