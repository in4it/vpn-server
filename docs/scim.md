# SCIM

## Introduction

SCIM (Cross-domain Identity Management) can copy users from an identity provider to the VPN Server. **The password is not copied**. SCIM only syncs users with your identity provider; it does not provide authentication. For authentication, configure SAML or OpenID Connect (OIDC).

Once SCIM is enabled, users that are deleted or suspended will be deleted or suspended in the VPN Server. This is not the case when only using SAML or OIDC for authentication.

## OneLogin Setup

* In the VPN Server, go to `Authentication & Provisioning`, click on the `Provisioning` tab and click on the checkbox `Enable SCIM v2 endpoint` to enable the SCIM endpoint
    * Copy the `Bearer token` and the `Base URL`
* Create a new application in OneLogin
* Search for `SCIM Provisioner with SAML (SCIM v2 Core)`. Even if you don't intend to use SAML, you can use this application for SCIM only
* Go to `Configuration` and paste the `Base URL` you copied from the VPN Server. Do the same for the `SCIM Bearer Token`
* Click `Enable` on the same `Configuration` page to enable the API connection
* Go to `Provisioning` and ensure `Enable provisioning` is enabled. Uncheck the three approval checkboxes unless you want to approve every change: `Create User`, `Delete User`, `Update User`
* Go to `Access` to link a OneLogin role to this application
* Once users are assigned to the application, you can initiate a manual sync by going to `Users`, clicking `More actions`, and then clicking `Sync logins`. If all goes well, you'll see the users provisioned with a green checkmark

## Unsupported Feature
Currently, there is no login button for SAML, unlike OpenID Connect. The SAML connection can be initiated from the identity provider.
