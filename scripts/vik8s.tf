terraform {
  required_providers {
    vik8s = {
      source  = "xhaiker/vik8s"
      version = "0.6.0"
    }
  }
}

provider "vik8s" {
  china = true
}

locals {
  path = "${path.module}/../.vagrant/machines"
}

resource "vik8s_cluster" "cluster" {
  nodes {
    address = "velcyr.eicp.net"
    port    = "21687"
    ssh_key = "$HOME/.ssh/id_rsa"
    role    = ["bastion"]
  }
  nodes {
    address = "192.168.10.176"
    ssh_key = "${local.path}/master01/vmware_esxi/private_key"
    role    = ["control_plane"]
    labels  = {
      name = "name"
    }
    bastion = "velcyr.eicp.net"
  }
  nodes {
    address = "192.168.11.160"
    ssh_key = "${local.path}/slave20/vmware_esxi/private_key"
    bastion = "velcyr.eicp.net"
    role    = ["etcd"]
  }
  nodes {
    username = "vagrant"
    address  = "192.168.11.152"
    ssh_key  = "${local.path}/slave21/vmware_esxi/private_key"
    bastion  = "velcyr.eicp.net"
  }
  config {
    repo = "vik8s.io"
  }
}
