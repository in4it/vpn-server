# Release Notes

## Version v1.1.21
* Google Cloud Marketplace support, and detection of the public IP address on GCP, Azure and DigitalOcean
* UI: confirmation before deleting users, connections, OIDC/SAML providers and MFA factors, and before regenerating the SCIM token
* UI: error messages for failed actions, and better handling of expired sessions
* UI: changing your own password now requires your current password
* VPN: a user can create up to 25 VPN connections
* VPN: VPN settings and templates are validated before they are saved, and the VPN keeps running when a new configuration can't be applied
* VPN: the server template no longer supports custom `PreUp`/`PostUp`/`PreDown`/`PostDown` commands. See the [FAQ](faq.md#the-server-template-contains-unsupported-settings) if your server template contains custom commands
* Authentication: improvements to OIDC, SAML and MFA (TOTP) logins
* SCIM: paging follows RFC 7644, renaming a user to an existing username returns an error, and the SCIM token keeps working after a restart
* Upgrade: more reliable upgrades and service restarts
* Packet logs: logging resumes automatically after a low disk space event
* `reset-admin-password`: the new password needs to be entered twice and needs to be at least 8 characters
* Dependency updates (Go 1.26.9 and npm packages)

Note: after upgrading, make sure to close any old browser tabs to ensure the new UI version is loaded.

## Version v1.1.5 - v1.1.10
* Maintenance release (dependency updates)

## Version v1.1.4
* Improved setup flow for AWS & DigitalOcean

## Version v1.1.3
* New feature: Log packets traversing the VPN Server. This release supports logging TCP / DNS / HTTP / HTTPS packets and inspecting the destination of HTTP/HTTPS packets.

## Version v1.1.2
* UI: fixes in user creation

## Version v1.1.0
* UI: change VPN configuration within the admin UI
* UI: ability to reload WireGuard® configuration 
* UI: modify client/server WireGuard® configuration files using templates

Note: after upgrading, make sure to close any old browser tabs to ensure the new UI version is loaded.

## Version v1.0.41
* UI: axios version bump
* UI: disable HTTPS forwarding when the request is served over HTTP
* UI: general improvements

## Version v1.0.40
* GCP marketplace release

## Version v1.0.39
* DigitalOcean marketplace release

## Version v1.0.38
* General bug fixes

## Version v1.0.37
* SAML Support for authentication
* SCIM Support for provisioning

## Version v1.0.36
* An administrator will now be alerted when there is a new version of the VPN Server available. An upgrade procedure to the latest version can be started from the admin web UI. 
* Minor bug fixes

Upgrade instructions can be found [here](upgrade.md).

Once upgraded to this release, new upgrades can be done through the UI.

## Version v1.0.35
* Fix an IP address management issue where the same IP address is handed out in some cases

## Version v1.0.34
* Fix config parsing issue in client config for Windows clients

## Version v1.0.33

* Profile page with password change and MFA support (Google Authenticator)

## Version v1.0.32
* Initial release

## Version v1.0.31

* Fixes to get to initial release

## Version v1.0.30

* Local Users Support
* OIDC Support
* WireGuard® for VPN Connections
