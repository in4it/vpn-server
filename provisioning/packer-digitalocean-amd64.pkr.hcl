packer {
  required_plugins {
    digitalocean = {
      version = ">= 1.0.4"
      source  = "github.com/digitalocean/digitalocean"
    }
  }
}

variable "do_api_token" {
  type      = string
  default   = "${env("DIGITALOCEAN_API_TOKEN")}"
  sensitive = true
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "digitalocean" "vpn-server" {
  snapshot_name      = "in4it-vpn-server-${local.timestamp}"
  image              = "ubuntu-24-04-x64"
  api_token          = var.do_api_token
  region             = "nyc3"
  size               = "s-1vcpu-1gb"
  ssh_username       = "root"
}

build {
  sources = ["source.digitalocean.vpn-server"]

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
    environment_vars = [
        "DEBIAN_FRONTEND=noninteractive",
        "LC_ALL=C",
        "LANG=en_US.UTF-8",
        "LC_CTYPE=en_US.UTF-8"
    ]
    execute_command = "{{ .Vars }} sudo -E sh '{{ .Path }}'"
    pause_before    = "10s"
    scripts         = ["scripts/install_vpn.sh"]
  }

  provisioner "shell" {
    inline = ["sudo rm /root/.ssh/authorized_keys"]
  }
  provisioner "shell" {
    inline = ["sudo apt-get -y purge droplet-agent"]
  }

  provisioner "shell" {
    scripts         = ["scripts/digitalocean/90-cleanup.sh", "scripts/digitalocean/99-img-check.sh"]
  }

}
