# SAML

## Introduction

SAML support enables users to log in using an identity provider, rather than having the users manually created by the `admin` user.

## Onelogin Setup

* Create a new Application in Onelogin
* Search for `SCIM Provisioner with SAML (SCIM v2 Core)` or a generic SAML application if you don't want SCIM (provisioning) support
* Go to `More Actions`, right click on `SAML Metadata`, click `Copy link address`
* Use this information to create the SAML Configuration in the VPN Server:
    * Metadata URL: the URL you just copied
    * Allow Missing Attributes: needs to be enabled for the SCIM Provisioner, as it doesn't pass the necessary SAML attributes
* Once the configuration is created, you can copy the `ACS URL` and the `Audience URL` and fill it out in the `Configuration` tab in Onelogin:
    * SAML Audience URL: the `Adience URL` from the VPN Server
    * SAML Consumer URL: the `ACS URL` from the VPN Server 

## Unsupported feature
Currently there's no login button for SAML (unlike for OpenID Connect). The SAML connection can be initiated from the Identity Provider.