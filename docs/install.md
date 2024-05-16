# Installation

## AWS 

Installation can be started using the AWS Marketplace. Once the provisioning of the EC2 instance is finished, point your browser to http://ip (not https - yet), to start the setup.

You'll be asked to provide a secret to go to the next step. Log in using SSH (or AWS SSM if you use AWS SSM). The login is `ubuntu`, use the SSH key you configured when setting up the EC2. Once logged in you can use the "cat" command to display the secret. Alternatively, you can use `sudo /vpn/reset-admin-password` to set an admin password securely over SSH.