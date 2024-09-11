# Release Notes

## Version v1.1.4
* Improved setup flow for AWS & DigitalOcean

## Version v1.1.3
* New Feature: Log packets traversing the VPN Server. This release supports logging TCP / DNS / HTTP / HTTPS packets and inspecting the destination of http/https packets.

## Version v1.1.2
* UI: fixes in user creation

## Version v1.1.0
* UI: change VPN configuration within the admin UI
* UI: ability to reload WireGuard® configuration 
* UI: modify client/server WireGuard® configuration files using templates

Note: after upgrading, make sure to close any old browser tab to ensure the new UI version is loaded.

## Version v1.0.41
* UI: axios version bump
* UI: disable https forwarding when request is served over http
* UI: general improvements

## Version v1.0.40
* GCP marketplace release

## Version v1.0.39
* DigitalOcean marketplace release

## Version v1.0.38
* General bugfixes

## Version v1.0.37
* SAML Support for authentication
* SCIM Support for provisioning

## Version v1.0.36
* An administrator will now be alerted when there is a new version of the VPN Server available. An upgrade procedure to the latest version can be started from the admin web UI. 
* Minor bugfixes

Upgrade instructions can be found [here](upgrade.md)

Once upgraded to this release, new upgrades can be done through the UI.

## Version v1.0.35
* Fix IP address management issue where in some cases the same IP is handed out

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
