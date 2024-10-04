packer {
  required_plugins {
    amazon = {
      source  = "github.com/hashicorp/amazon"
      version = "~> 1"
    }
  }
}

variable "aws_profile" {
  type    = string
  default = "${env("AWS_PROFILE")}"
}

variable "ami_users" {
  type    = list(string)
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

data "amazon-ami" "ubuntu" {
    filters = {
        virtualization-type = "hvm"
        name = "ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"
        root-device-type = "ebs"
    }
    owners = ["099720109477"]
    most_recent = true
}
// BYOL
source "amazon-ebs" "vpn-server-byol" {
  ami_name      = "in4it-vpn-server-byol-${local.timestamp}"
  ami_users     = var.ami_users
  instance_type = "m7a.medium"
  launch_block_device_mappings {
    delete_on_termination = true
    device_name           = "/dev/sda1"
    volume_size           = 30
    volume_type           = "gp3"
  }
  profile      = "${var.aws_profile}"
  region       = "us-east-1"
  ami_regions  = ["eu-west-1"]
  source_ami   = data.amazon-ami.ubuntu.id
  ssh_username = "ubuntu"
}

// License included
source "amazon-ebs" "vpn-server-licensed" {
  ami_name      = "in4it-vpn-server-licensed-${local.timestamp}"
  ami_users     = var.ami_users
  instance_type = "m7a.medium"
  launch_block_device_mappings {
    delete_on_termination = true
    device_name           = "/dev/sda1"
    volume_size           = 30
    volume_type           = "gp3"
  }
  profile      = "${var.aws_profile}"
  region       = "us-east-1"
  ami_regions  = ["eu-west-1"]
  source_ami   = data.amazon-ami.ubuntu.id
  ssh_username = "ubuntu"
}

build {
  sources = [
    "source.amazon-ebs.vpn-server-byol",
    "source.amazon-ebs.vpn-server-licensed"
  ]

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
    inline = ["rm /home/ubuntu/.ssh/authorized_keys"]
  }

  provisioner "shell" {
    inline = ["sudo rm /root/.ssh/authorized_keys"]
  }

}
