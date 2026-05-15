# Installation

## Cloud Provider

You can start the installation from the AWS Marketplace, Azure Marketplace, or DigitalOcean Marketplace. Once the instance has finished provisioning, point your browser to `http://ip` (not HTTPS yet) to start the setup.

You'll be asked to provide a secret to continue to the next step. Log in using SSH, or AWS Systems Manager Session Manager if you use AWS SSM. The username is `ubuntu`; use the SSH key you configured when setting up the instance. Once logged in, you can use the `cat` command to display the secret. Alternatively, you can run `sudo /vpn/reset-admin-password` to set an admin password securely over SSH.
