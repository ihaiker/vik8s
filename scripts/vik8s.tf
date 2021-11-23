terraform {
  required_providers {
    vik8s = {
      source  = "xhaiker/vik8s"
      version = "0.5.2"
    }
  }
}

provider "vik8s" {
}

locals {
  path = "${path.module}/../.vagrant/machines"
}

/*data "vik8s_host" "proxy" {
  username = "root"
  address  = "velcyr.eicp.net"
  port     = "21687"
}

data "vik8s_host" "master" {
  username    = "root"
  address     = "192.168.10.176"
  private_key = "${local.path}/master01/vmware_esxi/private_key"
  proxy       = data.vik8s_host.proxy.id
}*/

/*
data "vik8s_host" "slave20" {
  username    = "root"
  address     = "192.168.11.160"
  private_key = "${local.path}/slave20/vmware_esxi/private_key"
  proxy       = data.vik8s_host.proxy.id
}

data "vik8s_host" "slave21" {
  username    = "root"
  address     = "192.168.11.152"
  private_key = "${local.path}/slave21/vmware_esxi/private_key"
  proxy       = data.vik8s_host.proxy.id
}
*/
/*

resource "vik8s_cluster_node" "master" {
  host_id = "name"
  master  = false
}
*/

resource "vik8s_cluster_node" "slave" {
  host_id = "name"
  master  = false
  config {
    ntp_services = ["t1", "t2"]
  }

}
