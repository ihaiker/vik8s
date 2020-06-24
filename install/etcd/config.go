package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"os"
	"time"
)

type ETCD struct {
	SSH hosts.SSH `json:"ssh"`

	Nodes               []string `json:"nodes,omitempty"`
	Version             string   `json:"version,omitempty"`
	ServerCertExtraSans []string `json:"server-cert-extra-sans,omitempty"`

	CertsValidity time.Duration `json:"certsValidity"`
	CertsDir      string        `json:"certDir,omitempty"`

	Snapshot       string `json:"snapshot,omitempty"`
	RemoteSnapshot string `json:"remoteSnapshot,omitempty"`

	Source string `json:"source,omitempty"`
}

func (etcd *ETCD) Write() {
	etcdConfig := tools.Join("etcd.json")

	if len(Config.Nodes) == 0 {
		_ = os.Remove(etcdConfig)
		return
	}

	bs, _ := json.MarshalIndent(etcd, "", "    ")
	defer utils.Catch(func(err error) {
		fmt.Println("write config error ", err)
		fmt.Println("Be sure to save the following content to " + etcdConfig + ", very important! very important! very important! 重要！重要！重要！")
		fmt.Println(string(bs))
	})
	utils.Panic(os.MkdirAll(tools.Join(), os.ModePerm), "mkdir config file dir")
	utils.Panic(ioutil.WriteFile(etcdConfig, bs, 0666), "write config file")
}

func (etcd *ETCD) MustRead() {
	etcdConfigLoc := tools.Join("etcd.json")

	etcdConfigBytes, err := ioutil.ReadFile(etcdConfigLoc)
	utils.Panic(err, "read etcd config file %s", etcdConfigLoc)

	err = json.Unmarshal(etcdConfigBytes, Config)
	utils.Panic(err, "read etcd config file %s", etcdConfigLoc)
}
func (etcd *ETCD) Read() error {
	etcdConfigLoc := tools.Join("etcd.json")

	etcdConfigBytes, err := ioutil.ReadFile(etcdConfigLoc)
	if err != nil {
		return utils.Wrap(err, "read etcd config file %s", etcdConfigLoc)
	}
	err = json.Unmarshal(etcdConfigBytes, Config)
	if err != nil {
		return utils.Wrap(err, "read etcd config file %s", etcdConfigLoc)
	}
	return nil
}

func (etcd *ETCD) Exists(ip string) bool {
	for _, node := range etcd.Nodes {
		if node == ip {
			return true
		}
	}
	return false
}

func (etcd *ETCD) Join(ip string) {
	etcd.Nodes = append(etcd.Nodes, ip)
	etcd.Write()
}

func (etcd *ETCD) Remove(ip string) {
	for i, node := range etcd.Nodes {
		if node == ip {
			etcd.Nodes = append(etcd.Nodes[0:i], etcd.Nodes[i+1:]...)
			break
		}
	}
	etcd.Write()
}

var Config = new(ETCD)
