# SCIM

## Introduction

SCIM (Cross-domain Identity Management) can copy users from an Identity Provider to the VPN Server. **The password is not copied**. SCIM is only to sync the users with your identity provider, not to provide authentication. For Authentication, configure SAML or OpenID Connect (OIDC).

Once SCIM is enabled, users that are deleted or suspended will be deleted or suspended in the VPN Server. This is not the case when only using SAML or OIDC for authentication.

## Onelogin Setup

* In the VPN Server, go to `Authentication & Provisioning`, click on the `Provisioning` tab and click on the checkbox `Enable SCIM v2 endpoint` to enable the SCIM endpoint
    * Copy the `Bearer token` and the `Base URL`
* Create a new Application in Onelogin
* Search for `SCIM Provisioner with SAML (SCIM v2 Core)`. Even if you don't intend to use SAML, you can use this application for SCIM only
* Go to `Configuration` and paste the `Base URL` which you copied from the VPN Server. Do the same for the `SCIM Bearer Token`
* Click on `Enable` on the same `Configuration` page to enable the API Connection
* Go to `Provisioning` and ensure `Enable provisioning` is enabled. Uncheck the 3 approval checkboxes unless you want to give approval for every change: `Create User`, `Delete User`, `Update User`
* Go to `Access` to link a Onelogin role to this application
* Once users are assigned to the application, you can initiate a manual sync by going to `Users`, click `More actions` and then `Sync logins`. If all goes well you'll see the users being provisioned with a green checkmark

## Unsupported feature
Currently there's no login button for SAML (unlike for OpenID Connect). The SAML connection can be initiated from the Identity Provider.