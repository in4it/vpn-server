# Quick Start

# Login

The first user created is the `admin` user. You can login using the "admin" username and the "admin" password. If you forgot the password, login to the server using SSH and execute the command `sudo /vpn/reset-admin-password`.

# Create first user

You can create a first user on the user page. The `admin` user cannot create VPN connections.

* Create a new user on the user page
* Log-out and Log-in using the new credentials
* Create a new connection on the Connections page
* Download the configuration
* Import the configuration in a WireGuard® Client. See [https://www.wireguard.com/install/](https://www.wireguard.com/install/) for WireGuard® clients.

# Access for existing users in Identity Providers
To allow access for users created in Active Directory, Okta, Onelogin, or other Identity Providers, navigate to the [OIDC](oidc.md), [SAML](saml.md), or [Provisioning (SCIM)](scim.md) pages to setup an IdP connection.