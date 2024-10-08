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

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "amazon-ebs" "autogenerated_1" {
  ami_name      = "in4it-vpn-server ${local.timestamp}"
  instance_type = "t3.micro"
  profile       = "${var.aws_profile}"
  region        = "us-east-1"
  source_ami    = "ami-0f2a1bb3c242fe285"
  ssh_username  = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.autogenerated_1"]

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
