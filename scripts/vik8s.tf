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

data vik8s_hosts master {
  host = "192.168.10.176"
}

data vik8s_hosts adders {
  hosts      = ["192.168.11.160","192.168.11.152"]
  depends_on = [
    data.vik8s_hosts.master
  ]
}

resource "vik8s_cluster" "cluster" {
  masters    = concat(data.vik8s_hosts.master.nodes, data.vik8s_hosts.adders.nodes)
  depends_on = [
    data.vik8s_hosts.master,
    data.vik8s_hosts.adders
  ]
}
