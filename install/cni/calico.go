package cni

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

type calico struct {
	ipip    bool
	mtu     int
	version string
	repo    string
	typha   struct {
		Enable     bool
		Prometheus bool
		Replicas   int
	}

	etcd struct {
		Enable       bool
		TLS          bool
		Endpoints    []string
		Ca           string
		CaBase64     string `json:"-"`
		Key          string
		KeyBase64    string `json:"-"`
		Cert         string
		CertBase64   string `json:"-"`
		EndpointsUrl string
	}
}

func (f *calico) Name() string {
	return "calico"
}

func (f *calico) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.version, flags(f, "version"), "3.14.0", "")
	cmd.Flags().BoolVar(&f.ipip, flags(f, "ipip"), true, "Enable IPIP")
	cmd.Flags().IntVar(&f.mtu, flags(f, "mtu"), 1440, `Configure the MTU to use for workload interfaces and the tunnels.  
For IPIP, set to your network MTU - 20; for VXLAN set to your network MTU - 50.`)

	cmd.Flags().StringVar(&f.repo, flags(f, "repo"), "", fmt.Sprintf("Choose a container registry to pull control plane images from"))

	cmd.Flags().BoolVar(&f.typha.Enable, flags(f, "typha"), false, "Enable Typha, When install Calico with Kubernetes API datastore, more than 50 nodes")
	cmd.Flags().BoolVar(&f.typha.Prometheus, flags(f, "typha-prometheus"), false,
		"enable prometheus metrics.  Since Typha is host-networked, this opens a port on the host, which may need to be secured.")
	cmd.Flags().IntVar(&f.typha.Replicas, flags(f, "typha-replicas"), 1,
		"replicas number, see Deployment 'calico-typha' at https://docs.projectcalico.org/manifests/calico-typha.yaml")

	cmd.Flags().BoolVar(&f.etcd.Enable, flags(f, "etcd"), false,
		fmt.Sprintf(`install calico with etcd datastore. 
If you enable etcd to store data and not specify the -%s parameter, 
the system will look for the etcd cluster from the following two points.
  1. Get the etcd cluster provided by '--etcd'
  2. Obtain the etcd cluster that comes with kubernates.`,
			flags(f, "etcd-endpoints")))

	cmd.Flags().BoolVar(&f.etcd.TLS, flags(f, "etcd-tls"), true, "TLS enabled etcd")
	cmd.Flags().StringSliceVar(&f.etcd.Endpoints, flags(f, "etcd-endpoints"), []string{}, "the location of your etcd cluster. for example: 172.16.100.10:2379")
	cmd.Flags().StringVar(&f.etcd.Ca, flags(f, "etcd-ca"), "", "the etcd ca file path")
	cmd.Flags().StringVar(&f.etcd.Key, flags(f, "etcd-key"), "", "the etcd key file path")
	cmd.Flags().StringVar(&f.etcd.Cert, flags(f, "etcd-cert"), "", "the etcd cert file path\n")

}

func (f *calico) applyVik8sETCDServer(node *ssh.Node, vip string) {
	name := "yaml/cni/calico-vik8s-etcd.conf"
	reduce.MustApplyAssert(node, name, paths.Json{
		"VIP": vip,
	})
}

func (f *calico) Apply(master *ssh.Node) {
	if f.repo != "" && strings.HasSuffix(f.repo, "/") {
		f.repo = f.repo + "/"
	}
	data := paths.Json{
		"Version": "v" + f.version, "Repo": f.repo,
		"IPIP": f.ipip, "MTU": f.mtu,
		"CIDR": config.K8S().PodCIDR, "Interface": config.K8S().Interface,
		"Typha": f.typha,
	}

	local := "yaml/cni/calico.conf"
	if f.etcd.Enable {
		if len(f.etcd.Endpoints) == 0 {
			f.etcd.TLS = true
			//第一步, 使用 vik8s etcd init 安装了etcd集群
			if config.ExternalETCD() {
				f.etcd.Endpoints = config.Etcd().Nodes
				dir := etcd.CertsDir()
				f.etcd.Ca = filepath.Join(dir, "ca.crt")
				f.etcd.Key = filepath.Join(dir, "apiserver-etcd-client.crt")
				f.etcd.Cert = filepath.Join(dir, "apiserver-etcd-client.key")
			} else {
				vip := tools.GetVip(config.K8S().SvcCIDR, tools.Vik8sCalicoETCD)
				f.applyVik8sETCDServer(master, vip)
				certsDir := paths.Join("kube/pki/etcd")
				certPath, keyPath := etcdcerts.CreateCalicoETCDPKIAssert(certsDir, config.K8S().CertsValidity)
				f.etcd.Endpoints = []string{vip + ":2379"}
				f.etcd.Ca = paths.Join("kube/pki/etcd/ca.crt")
				f.etcd.Key = keyPath
				f.etcd.Cert = certPath
			}
		}

		//base64 cert files
		if f.etcd.TLS {
			f.etcd.CaBase64 = utils.Base64File(f.etcd.Ca)
			f.etcd.KeyBase64 = utils.Base64File(f.etcd.Key)
			f.etcd.CertBase64 = utils.Base64File(f.etcd.Cert)
		}

		f.etcd.EndpointsUrl = ""
		for i, endpoint := range f.etcd.Endpoints {
			if i != 0 {
				f.etcd.EndpointsUrl += ","
			}
			if f.etcd.TLS {
				f.etcd.EndpointsUrl += "https://" + endpoint
			} else {
				f.etcd.EndpointsUrl += "http://" + endpoint
			}
		}

		data["Etcd"] = f.etcd
		local = "yaml/cni/calico-etcd.conf"
	} else if f.typha.Enable {
		local = "yaml/cni/calico-typha.conf"
	}

	reduce.MustApplyAssert(master, local, data)

	/*bs, _ := json.Marshal(f.etcd)
	k8s.Config.CNI.Params = map[string]string{
		"Version": f.version, "Repo": f.repo,
		"IPIP": strconv.FormatBool(f.ipip), "MTU": strconv.Itoa(f.mtu),
		"TyphaEnable":     strconv.FormatBool(f.typha.Enable),
		"TyphaPrometheus": strconv.FormatBool(f.typha.Prometheus),
		"TyphaReplicas":   strconv.Itoa(f.typha.Replicas),
		"Etcd":            string(bs),
	}*/
}

func (f *calico) Clean(node *ssh.Node) {
	_, _ = node.Cmd("rm -rf /var/lib/calico")
	_, _ = node.Cmd("rm -f /etc/NetworkManager/conf.d/calico.conf")
	_, _ = node.Cmd("modprobe -r ipip")
	_, _ = node.Cmd("ip link delete tunl0@NONE")
	_, _ = node.Cmd("ifconfig tunl0 down")
}
