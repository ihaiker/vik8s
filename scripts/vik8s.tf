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

data "vik8s_host" "proxy" {
  username    = "root"
  private_key = "/Users/haiker/workbench/self/esxi/admin/.vagrant/machines/admin/vmware_esxi/private_key"
  address     = "velcyr.eicp.net"
  port        = "21687"
}

data "vik8s_host" "master" {
  username    = "root"
  address     = "192.168.10.176"
  private_key = "/Users/haiker/workbench/self/go/vik8s/.vagrant/machines/master01/vmware_esxi/private_key"
  proxy       = data.vik8s_host.proxy.id
}
