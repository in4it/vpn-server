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

# Marketplace license from the Producer Portal (Deployment package section),
# e.g. projects/<project>/global/licenses/cloud-marketplace-<id>
variable "image_license" {
  type      = string
  default   = "${env("GCP_MARKETPLACE_LICENSE")}"
  sensitive = true
}

variable "zone" {
  type    = string
  default = "us-east4-c"
}

# version from ./latest, e.g. v1.1.20 (dots are not allowed in image names)
variable "image_version" {
  type    = string
  default = "dev"
}

source "googlecompute" "vpn-server" {
  project_id          = var.project_id
  source_image_family = "ubuntu-2404-lts-amd64"
  zone                = var.zone
  machine_type        = "e2-standard-2"
  ssh_username        = "ubuntu"
  image_name          = "in4it-vpn-server-ubuntu2404-x86-64-${replace(lower(var.image_version), ".", "-")}-${local.timestamp}"
  image_family        = "in4it-vpn-server"
  image_description   = "in4it VPN Server ${var.image_version}"
  image_labels        = { version = replace(lower(var.image_version), ".", "-") }
  image_licenses      = [var.image_license]
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

  post-processor "manifest" {
    output     = "packer-gcp-manifest.json"
    strip_path = true
  }
}
