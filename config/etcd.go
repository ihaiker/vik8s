package config

import (
	"time"
)

type ETCD struct {
	Nodes               []string `flag:"-"`
	Version             string   `def:"v3.4.13" help:"etcd version"`
	ServerCertExtraSans []string `help:"optional extra Subject Alternative Names for the etcd server signing cert, can be multiple comma separated DNS names or IPs"`

	CertsValidity time.Duration `def:"4y" help:"Certificate validity time"`
	CertsDir      string        `def:"/etc/etcd/pki" help:"certificates directory"`
	Data          string        `def:"/var/lib/etcd" help:"etcd data dir"`

	Snapshot       string `help:"Etcd v3 snapshot (local disk) file used to initialize member"`
	RemoteSnapshot string `help:"Etcd v3 snapshot (remote disk at first node) file used to initialize member"`
}
