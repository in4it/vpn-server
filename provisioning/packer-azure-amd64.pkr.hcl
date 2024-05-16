packer {
  required_plugins {
    azure = {
      source  = "github.com/hashicorp/azure"
      version = "~> 1"
    }
  }
}

variable "image_version" {
  type    = string
}

source "azure-arm" "vpn-server" {
  image_offer     = "ubuntu-24_04-lts"
  image_publisher = "Canonical"
  image_sku       = "server"
  location        = "East US"
  os_type         = "linux"
  shared_image_gallery_destination {
    resource_group = "vpn-server"
    gallery_name   = "vpnserver"
    image_name     = "in4it-vpn-server"
    image_version  = replace(var.image_version, "v", "")
  }
  use_azure_cli_auth = true
  vm_size            = "Standard_DC1s_v3"
}

build {
  sources = ["source.azure-arm.vpn-server"]

  provisioner "file" {
    destination = "/tmp/configmanager-linux-amd64"
    source      = "../configmanager-linux-amd64"
  }

  provisioner "file" {
    destination = "/tmp/reset-admin-password-linux-amd64"
    source      = "../reset-admin-password-linux-amd64"
  }

  provisioner "file" {
    destination = "/tmp/restserver-linux-amd64"
    source      = "../restserver-linux-amd64"
  }

  provisioner "file" {
    destination = "/tmp/vpn-configmanager.service"
    source      = "systemd/vpn-configmanager.service"
  }

  provisioner "file" {
    destination = "/tmp/vpn-rest-server.service"
    source      = "systemd/vpn-rest-server.service"
  }

  provisioner "shell" {
    execute_command = "{{ .Vars }} sudo -E sh '{{ .Path }}'"
    pause_before    = "10s"
    scripts         = ["scripts/install_vpn.sh"]
  }

  provisioner "shell" {
    inline = ["sudo -s -- sh -c 'rm /home/packer/.ssh/authorized_keys; waagent -deprovision+user -force'"]
  }

}
