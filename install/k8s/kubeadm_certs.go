package k8s

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	kubecerts "github.com/ihaiker/vik8s/certs/kubernetes"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"net"
	"path/filepath"
)

func makeCerts(node *ssh.Node) {
	if Config.ETCD.External {
		scpExternalEtcdCa(node)
	} else {
		makeEtcdCerts(node)
	}
	makeKubeCerts(node)
}

func makeJoinControlPlaneConfigFiles(node *ssh.Node) {
	dir := tools.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", Config.Kubernetes.ApiServer)
	files := kubecerts.CreateJoinControlPlaneKubeConfigFiles(dir, node.Hostname, endpoint, Config.CertsValidity)
	for key, path := range files {
		remote := filepath.Join("/etc/kubernetes", fmt.Sprintf("%s.conf", key))
		utils.Panic(node.Scp(path, remote), "scp %s %s", path, remote)
	}
}

func makeWorkerConfigFiles(node *ssh.Node) {
	dir := tools.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", Config.Kubernetes.ApiServer)
	files := kubecerts.CreateWorkerKubeConfigFile(dir, node.Hostname, endpoint, Config.CertsValidity)
	for key, path := range files {
		remote := filepath.Join("/etc/kubernetes", fmt.Sprintf("%s.conf", key))
		utils.Panic(node.Scp(path, remote), "scp %s %s", path, remote)
	}
}

func makeEtcdCerts(node *ssh.Node) {
	node.Logger("make etcd certs files")

	name := node.Hostname
	dir := tools.Join("kube", "pki", "etcd")

	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host, net.IPv6loopback.String()}
	sans = append(sans, utils.ParseIPS(Config.Masters)...)
	if Config.CNI.Name == "calico" {
		sans = append(sans, tools.GetVip(Config.Kubernetes.SvcCIDR, tools.Vik8sCalicoETCD), "vik8s-calico-etcd")
	}
	vt := Config.CertsValidity

	etcdcerts.CreatePKIAssets(name, dir, sans, vt)

	certsFiles := map[string]string{
		"ca":                    "etcd/ca",
		"server-" + name:        "etcd/server",
		"peer-" + name:          "etcd/peer",
		"healthcheck-client":    "etcd/healthcheck-client",
		"apiserver-etcd-client": "apiserver-etcd-client",
	}
	scpCerts(certsFiles, node, dir)
}

func scpExternalEtcdCa(node *ssh.Node) {
	if !Config.ETCD.External {
		return
	}
	files := map[string]string{
		Config.ETCD.CAFile:            "/etc/kubernetes/pki/etcd/ca.crt",
		Config.ETCD.ApiServerKeyFile:  "/etc/kubernetes/pki/etcd/apiserver-etcd-client.key",
		Config.ETCD.ApiServerCertFile: "/etc/kubernetes/pki/etcd/apiserver-etcd-client.crt",
	}
	for local, remote := range files {
		utils.Panic(node.Scp(local, remote), "[%s] scp %s %s", node.Host, local, remote)
	}
}

func makeKubeCerts(node *ssh.Node) {
	certNode := kubecerts.Node{
		Name:                node.Hostname,
		Host:                node.Host,
		ApiServer:           Config.Kubernetes.ApiServer,
		SvcCIDR:             Config.Kubernetes.SvcCIDR,
		CertificateValidity: Config.CertsValidity,
		SANS:                Config.Kubernetes.ApiServerCertExtraSans,
	}
	//apisever = MasterIPS + VIP + CertSANS
	for _, masterIp := range Config.Masters {
		masterNode := hosts.Get(masterIp)
		certNode.SANS = append(certNode.SANS, masterNode.Hostname, masterNode.Host)
	}

	node.Logger("make kube certs files")

	dir := tools.Join("kube", "pki")
	kubecerts.CreatePKIAssets(dir, certNode)

	//sa
	{
		utils.Panic(node.Scp(filepath.Join(dir, "sa.key"), "/etc/kubernetes/pki/sa.key"), "scp sa.key")
		utils.Panic(node.Scp(filepath.Join(dir, "sa.pub"), "/etc/kubernetes/pki/sa.pub"), "scp sa")
	}
	certsFiles := map[string]string{
		//public
		"ca": "ca", "front-proxy-ca": "front-proxy-ca",
		//node
		"apiserver-" + node.Hostname:                "apiserver",
		"apiserver-kubelet-client-" + node.Hostname: "apiserver-kubelet-client",
		"front-proxy-client-" + node.Hostname:       "front-proxy-client",
	}
	scpCerts(certsFiles, node, dir)
}

func scpCerts(certsFiles map[string]string, node *ssh.Node, localDir string) {
	remoteDir := "/etc/kubernetes/pki"
	for lf, rf := range certsFiles {
		for _, exp := range []string{".key", ".crt"} {
			local := filepath.Join(localDir, lf+exp)
			remote := filepath.Join(remoteDir, rf+exp)
			utils.Panic(node.Scp(local, remote), "scp %s %s", local, remote)
		}
	}
}
