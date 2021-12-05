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

#data vik8s_hosts bastion {
#  host    = "velcyr.eicp.net"
#  port    = "21687"
#}

data vik8s_hosts master {
  host    = "192.168.10.176"
  ssh_key = "${local.path}/master01/vmware_esxi/private_key"
  #  bastion = data.vik8s_hosts.bastion.nodes.0
}

data vik8s_hosts slave20 {
  host    = "192.168.11.160"
  ssh_key = "${local.path}/slave20/vmware_esxi/private_key"
  #  bastion = data.vik8s_hosts.bastion.nodes.0
}

data vik8s_hosts slave21 {
  address = "192.168.11.152"
  ssh_key = "${local.path}/slave21/vmware_esxi/private_key"
  #  bastion = data.vik8s_hosts.bastion.nodes.0
}

data "vik8s_cri_docker" "docker" {}

resource "vik8s_etcd" "etcd" {
  nodes = concat(data.vik8s_hosts.master.nodes, data.vik8s_hosts.slave20.nodes, data.vik8s_hosts.slave21.nodes)
}
