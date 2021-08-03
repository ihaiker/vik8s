package config

import (
	"time"
)

type ETCD struct {
	Token               string   `ngx:"token" flag:"token" help:"cluster token"`
	Nodes               []string `ngx:"nodes" flag:"-"`
	Version             string   `ngx:"version" def:"v3.4.13" help:"etcd version"`
	ServerCertExtraSans []string `ngx:"server-cert-extra-sans" help:"optional extra Subject Alternative Names for the etcd server signing cert, can be multiple comma separated DNS names or IPs"`

	CertsValidity time.Duration `ngx:"certs-validity" def:"876000h" help:"Certificate validity time"`
	CertsDir      string        `ngx:"certs-dir" def:"/etc/etcd/pki" help:"certificates directory"`
	Data          string        `ngx:"data" def:"/var/lib/etcd" help:"etcd data dir"`

	Snapshot       string `ngx:"snapshot" help:"Etcd v3 snapshot (local disk) file used to initialize member"`
	RemoteSnapshot string `ngx:"remote-snapshot" help:"Etcd v3 snapshot (remote disk at first node) file used to initialize member"`

	Repo string `ngx:"repo" flag:"repo" help:"the repo url"`
}
