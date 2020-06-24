package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"os"
	"time"
)

type config struct {
	SSH hosts.SSH `json:"ssh"`

	Masters []string `json:"masters,omitempty"`
	Nodes   []string `json:"nodes,omitempty"`

	ETCD struct {
		External          bool     `json:"external"` //是否使用外部etcd集群
		Nodes             []string `json:"nodes,omitempty"`
		CAFile            string   `json:"ca,omitempty"`
		ApiServerKeyFile  string   `json:"apiserver-key-file,omitempty"`
		ApiServerCertFile string   `json:"apiserver-cert-file,omitempty"`
	} `json:"etcd"`

	Docker struct {
		Version      string   `json:"version"`
		DaemonJson   string   `json:"daemonJson,omitempty"`
		Registry     []string `json:"registry,omitempty"`
		CheckVersion bool     `json:"checkVersion"`
	} `json:"docker"`

	Kubernetes struct {
		Version       string `json:"k8s-version"`
		KubeadmConfig string `json:"kubeadm-config,omitempty"`

		ApiServer              string   `json:"apiserver"`
		ApiServerCertExtraSans []string `json:"apiServerCertExtraSans"`
		ApiServerVIP           string   `json:"apiserver-vip"`

		Repo string `json:"repo,omitempty"`

		Interface string `json:"interface"`
		PodCIDR   string `json:"pod-cidr"`
		SvcCIDR   string `json:"svc-cidr"`
	} `json:"kubernetes"`

	CNI struct {
		Name   string            `json:"name"`
		Params map[string]string `json:"params,omitempty"`
	} `json:"cni,omitempty"`

	CertsValidity time.Duration `json:"certsValidity"`

	Timezone    string   `json:"timezone"`
	NTPServices []string `json:"ntpServices"`
}

var Config = new(config)

func (cfg *config) Master() *ssh.Node {
	return hosts.Get(cfg.Masters[0])
}

func (cfg *config) LoadCmd(cmd *cobra.Command, args []string) {
	cfg.Load()
}

func (cfg *config) Load() {
	config := tools.Join("config.json")
	bs, err := ioutil.ReadFile(config)
	utils.Panic(err, "Is your system uninitialized?")
	utils.Panic(json.Unmarshal(bs, cfg), "parse config file %s ", config)
	cfg.Parse()
}

func (cfg *config) Write() {
	config := tools.Join("config.json")

	if len(cfg.Nodes) == 0 && len(cfg.Masters) == 0 {
		_ = os.Remove(config)
		return
	}

	bs, _ := json.MarshalIndent(cfg, "", "    ")
	defer utils.Catch(func(err error) {
		fmt.Println("write config error ", err)
		fmt.Println("Be sure to save the following content to " + config + ", very important! very important! very important! 重要！重要！重要！")
		fmt.Println(string(bs))
	})

	utils.Panic(os.MkdirAll(tools.Join(), os.ModePerm), "mkdir config file dir")
	utils.Panic(ioutil.WriteFile(config, bs, 0666), "write config file")
}

func (cfg *config) readInstallETCDCluster() {
	//用户自己架设的外部etcd集群
	if len(cfg.ETCD.Nodes) == 0 {
		if err := etcd.Config.Read(); err == nil {
			if len(etcd.Config.Nodes) > 0 {
				cfg.ETCD.External = true
				cfg.ETCD.Nodes = utils.Append(utils.ParseIPS(etcd.Config.Nodes), ":2379")
				cfg.ETCD.CAFile = tools.Join("etcd/pki/ca.crt")
				cfg.ETCD.ApiServerKeyFile = tools.Join("etcd/pki/apiserver-etcd-client.key")
				cfg.ETCD.ApiServerCertFile = tools.Join("etcd/pki/apiserver-etcd-client.crt")
			}
		}
	}
}

func (cfg *config) Parse() {
	cfg.SSH.PkFile = os.ExpandEnv(cfg.SSH.PkFile)
	cfg.readInstallETCDCluster()

	_, _, err := net.ParseCIDR(Config.Kubernetes.PodCIDR)
	utils.Panic(err, "invalid --pod-cidr %s", cfg.Kubernetes.PodCIDR)

	_, _, err = net.ParseCIDR(Config.Kubernetes.SvcCIDR)
	utils.Panic(err, "invalid --svc-cidr %s", cfg.Kubernetes.SvcCIDR)

	cfg.Kubernetes.ApiServerVIP = tools.GetVip(cfg.Kubernetes.SvcCIDR, tools.Vik8sApiServer)

	if cfg.Kubernetes.Repo == "" {
		cfg.Kubernetes.Repo = repo.KubeletImage()
	}
}

func (cfg *config) ExistsNode(ip string) (exists bool, master bool) {
	for _, node := range cfg.Masters {
		if node == ip {
			exists = true
			master = true
			return
		}
	}
	for _, node := range cfg.Nodes {
		if node == ip {
			exists = true
			master = false
			return
		}
	}
	return
}

func (cfg *config) JoinNode(master bool, ip string) {
	if master {
		cfg.Masters = append(cfg.Masters, ip)
	} else {
		cfg.Nodes = append(cfg.Nodes, ip)
	}
	cfg.Write()
}

func (cfg *config) RemoveNode(ip string) {
	for i, node := range cfg.Nodes {
		if node == ip {
			cfg.Nodes = append(cfg.Nodes[0:i], cfg.Nodes[i+1:]...)
			break
		}
	}
	for i, node := range cfg.Masters {
		if node == ip {
			cfg.Masters = append(cfg.Masters[0:i], cfg.Masters[i+1:]...)
			break
		}
	}
	cfg.Write()
}
