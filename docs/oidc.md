# OIDC

## Introduction

OIDC support enables users to log in using an identity provider, rather than requiring the `admin` user to create them manually.

## OneLogin Setup

* Create a new application in OneLogin
* Search for the OIDC type
* In the SSO tab, copy the Client ID and Client Secret
* Click the `Well-known Configuration` link. Copy the URL from the address bar and use it as the discovery URI
* Ensure `Authentication Method` is `POST` (not `Basic`)
* Use this information to create the OIDC configuration in the VPN Server
    * Note: When adding the OIDC configuration in the VPN Server, make sure to remove `offline_access` from the scopes.
* Once the configuration is created, copy the `redirect URI` and add it to the `Redirect URIs` text box in the `Configuration` tab
* A new login button will appear when you try to log in to the VPN. If you also want to initiate login from the OneLogin portal, copy the `Login URL` and add it to the `Login URL` text box in `Configuration` in OneLogin

## Azure OIDC Setup

* Go to Microsoft AD / Microsoft Entra ID
* Click on `manage`, then `app registrations`
* Click on `New registration`
* Give it a name. If you only want organization users to log in, use the `Single Tenant` option
* You can fill out the Redirect URI later, after completing the OIDC configuration in the VPN Server
* Once the `registration` is created, copy the Client ID and create a new Client Secret
* The Discovery URI can be found by clicking on `Endpoints`. The correct URL is under `OpenID Connect metadata document`
* Use this information to create the OIDC connection in the VPN Server
* Once the VPN server shows you the `redirect URI`, copy this link, browse to the `Authentication` page in the Azure portal under the same `App registration`, and enter it under the `Web Redirect URIs`
