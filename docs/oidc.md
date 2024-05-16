# OIDC

## Introduction

OIDC support enables users to log in using an identity provider, rather than having the users manually created by the `admin` user.

## Onelogin Setup

* Create a new Application in Onelogin
* Search for the OIDC type
* In the SSO tab, Client ID and Client Secret can be copied
* Click on the `Well-known Configuration` link. Copy the URL from the URL bar and use this as the discovery URI
* Ensure `Authentication Method` is `POST` (not `Basic`)
* Use this information to create the OIDC configuration in the VPN Server
    * Note: When adding the OIDC configuration in the VPN Server, make sure to remove `offline_access` from the scopes.
* Once the configuration is created, you can copy the `redirect URI` and add this in the `redirect URI's` textbox in the `Configuration` tab
* A new login button will appear when trying to log in to the VPN. If you also want to initiate a login from the Onelogin portal, also copy the `Login URL` and fill in out in `Login Url` textbox in `Configuration` in Onelogin  

## Azure OIDC Setup

* Go to Microsoft AD / Microsoft Entra ID
* Click on `manage`, then `app registrations`
* Click on `New registration`
* Give it a name. If you only want organization users to login, use the `Single Tenant` option
* Redirect URI can be filled out later, when we completed the OIDC configuration in the VPN Server
* Once the `registration` is created, you can copy the Client ID, create a new Client Secret
* The Discovery URI can be found by clicking on `Endpoints`. The correct URL is under `OpenID Connect metadata document`
* Use this information to create the OIDC Connection in the VPN Server
* Once the VPN server shows you the `redirect URI`, copy this link, browse to the `Authentication` page in the Azure portal under the same `App registration`, and enter it under the `Web Redirect URIs`