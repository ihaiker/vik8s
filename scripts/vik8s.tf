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
  address     = "192.168.11.45"
}

data "vik8s_host" "master" {
  username = "root"
  address  = "10.24.1.4"
  password = "jianchi"
}

data "vik8s_hosts" "slaves" {
  username = "root"
  address  = "10.24.1.12-14"
  password = "jianchi"
}
