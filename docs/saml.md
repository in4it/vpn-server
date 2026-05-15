# SAML

## Introduction

SAML support enables users to log in using an identity provider, rather than requiring the `admin` user to create them manually.

## OneLogin Setup

* Create a new application in OneLogin
* Search for `SCIM Provisioner with SAML (SCIM v2 Core)` or a generic SAML application if you don't want SCIM (provisioning) support
* Go to `More Actions`, right-click `SAML Metadata`, then click `Copy link address`
* Use this information to create the SAML configuration in the VPN Server:
    * Metadata URL: the URL you just copied
    * Allow Missing Attributes: needs to be enabled for the SCIM Provisioner, as it doesn't pass the necessary SAML attributes
* Once the configuration is created, copy the `ACS URL` and the `Audience URL` and add them to the `Configuration` tab in OneLogin:
    * SAML Audience URL: the `Audience URL` from the VPN Server
    * SAML Consumer URL: the `ACS URL` from the VPN Server 

## Unsupported Feature
Currently, there is no login button for SAML, unlike OpenID Connect. The SAML connection can be initiated from the identity provider.
