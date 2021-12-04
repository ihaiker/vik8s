package config

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"time"
)

type ETCD struct {
	Token               string   `ngx:"token" flag:"token" help:"cluster token"`
	Nodes               []string `ngx:"nodes" flag:"-"`
	Version             string   `ngx:"version" help:"etcd version"`
	ServerCertExtraSans []string `ngx:"server-cert-extra-sans" help:"optional extra Subject Alternative Names for the etcd server signing cert, can be multiple comma separated DNS names or IPs"`

	CertsValidity time.Duration `ngx:"certs-validity" help:"Certificate validity time"`
	CertsDir      string        `ngx:"certs-dir" help:"certificates directory"`
	Data          string        `ngx:"data" help:"etcd data dir"`

	Snapshot       string `ngx:"snapshot" help:"Etcd v3 snapshot (local disk) file used to initialize member"`
	RemoteSnapshot string `ngx:"remote-snapshot" help:"Etcd v3 snapshot (remote disk at first node) file used to initialize member"`

	Repo string `ngx:"repo" flag:"repo" help:"the repo url"`
}

func (this *ETCD) RemoveNode(node string) bool {
	idx := utils.Search(this.Nodes, node)
	if idx != -1 {
		this.Nodes = append(this.Nodes[0:idx], this.Nodes[idx+1:]...)
	}
	return idx != -1
}

func DefaultETCDConfiguration() *ETCD {
	return &ETCD{
		Version:       "v3.4.13",
		CertsValidity: time.Now().AddDate(100, 0, 0).Sub(time.Now()),
		CertsDir:      "/etc/etcd/pki",
		Data:          "/var/lib/etcd",
	}
}
