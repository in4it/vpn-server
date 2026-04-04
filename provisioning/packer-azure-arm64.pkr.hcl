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
  image_sku       = "server-arm64"
  location        = "East US"
  os_type         = "linux"
  shared_image_gallery_destination {
    resource_group = "vpn-server"
    gallery_name   = "vpnserver"
    image_name     = "in4it-vpn-server-arm64"
    image_version  = replace(var.image_version, "v", "")
  }
  use_azure_cli_auth = true
  vm_size            = "Standard_D2ps_v6"
  public_ip_sku      = "Standard"
}

build {
  sources = ["source.azure-arm.vpn-server"]

  provisioner "file" {
    destination = "/tmp/configmanager-linux-aarch64"
    source      = "../configmanager-linux-arm64"
  }

  provisioner "file" {
    destination = "/tmp/reset-admin-password-linux-aarch64"
    source      = "../reset-admin-password-linux-arm64"
  }

  provisioner "file" {
    destination = "/tmp/restserver-linux-aarch64"
    source      = "../restserver-linux-arm64"
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
    inline = [
      # Remove specific offending files flagged by Marketplace malware scanners
      "sudo find /usr/src -type f -name 'pismo.h' -exec rm -f {} +",
      "sudo find /usr/src -type f -path '*/drivers/mtd/maps/Kconfig' -exec rm -f {} +",

      "sudo grep -R 'pismoworld' /usr/src 2>/dev/null || echo 'No remaining references to pismoworld.org'",
    ]
  }

  provisioner "shell" {
    inline = ["sudo -s -- sh -c 'rm /home/packer/.ssh/authorized_keys; waagent -deprovision+user -force'"]
  }

}
