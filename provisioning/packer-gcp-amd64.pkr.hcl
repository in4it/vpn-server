packer {
  required_plugins {
    googlecompute = {
      source  = "github.com/hashicorp/googlecompute"
      version = "~> 1"
    }
  }
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

variable "project_id" {
  type      = string
  default   = "${env("GCP_PROJECT_ID")}"
  sensitive = true
}

source "googlecompute" "vpn-server" {
  project_id          = var.project_id
  source_image_family = "ubuntu-2404-lts-amd64"
  zone                = "us-east4-b"
  machine_type        = "n1-standard-1"
  ssh_username        = "ubuntu"
  image_name          = "in4it-vpn-server-${local.timestamp}"
  image_licenses      = ["projects/in4it-public/global/licenses/cloud-marketplace-f66537e5f7276a36-df1ebeb69c0ba664"]
}

build {
  sources = ["source.googlecompute.vpn-server"]

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
    inline = ["rm /home/ubuntu/.ssh/authorized_keys"]
  }
}
