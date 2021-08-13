package k8s

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	kubecerts "github.com/ihaiker/vik8s/certs/kubernetes"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"net"
	"path/filepath"
)

func makeKubernetesCerts(node *ssh.Node) {
	if config.ExternalETCD() {
		scpExternalEtcdCa(node)
	} else {
		makeEtcdCerts(node)
		makeEtcdctlCommand(node)
	}
	makeKubeCerts(node)
}

func makeJoinControlPlaneConfigFiles(node *ssh.Node) {
	dir := paths.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", config.K8S().ApiServer)
	files := kubecerts.CreateJoinControlPlaneKubeConfigFiles(dir, node.Hostname, endpoint, config.K8S().CertsValidity)
	for key, path := range files {
		remote := filepath.Join("/etc/kubernetes", fmt.Sprintf("%s.conf", key))
		err := node.SudoScp(path, remote)
		utils.Panic(err, "scp %s %s", path, remote)
	}
}

func makeWorkerConfigFiles(node *ssh.Node) {
	dir := paths.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", config.K8S().ApiServer)
	files := kubecerts.CreateWorkerKubeConfigFile(dir, node.Hostname, endpoint, config.K8S().CertsValidity)
	for key, path := range files {
		remote := filepath.Join("/etc/kubernetes", fmt.Sprintf("%s.conf", key))
		utils.Panic(node.SudoScp(path, remote), "scp %s %s", path, remote)
	}
}

func makeEtcdCerts(node *ssh.Node) {
	node.Logger("make etcd certs files")

	name := node.Hostname
	dir := paths.Join("kube", "pki", "etcd")

	// local + master + apiserversans + apiserver-vip
	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host, net.IPv6loopback.String()}
	sans = append(sans, utils.ParseIPS(config.K8S().Masters)...)
	if node.Hostname != node.Facts.Hostname {
		sans = append(sans, node.Facts.Hostname)
	}
	sans = append(sans, config.K8S().ApiServer, config.K8S().ApiServerVIP)

	vt := config.K8S().CertsValidity
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
	dir := etcd.CertsDir()
	files := map[string]string{
		filepath.Join(dir, "ca.crt"):                    "/etc/kubernetes/pki/etcd/ca.crt",
		filepath.Join(dir, "apiserver-etcd-client.key"): "/etc/kubernetes/pki/etcd/apiserver-etcd-client.key",
		filepath.Join(dir, "apiserver-etcd-client.crt"): "/etc/kubernetes/pki/etcd/apiserver-etcd-client.crt",
	}
	for local, remote := range files {
		err := node.SudoScp(local, remote)
		utils.Panic(err, "scp %s %s", local, remote)
	}
}

func makeKubeCerts(node *ssh.Node) {
	certNode := kubecerts.Node{
		Name:                node.Hostname,
		Host:                node.Host,
		ApiServer:           config.K8S().ApiServer,
		SvcCIDR:             config.K8S().SvcCIDR,
		CertificateValidity: config.K8S().CertsValidity,
		SANS:                config.K8S().ApiServerCertExtraSans,
	}
	//apisever = MasterIPS + VIP + CertSANS
	for _, masterIp := range config.K8S().Masters {
		masterNode := hosts.Get(masterIp)
		certNode.SANS = append(certNode.SANS, masterNode.Hostname, masterNode.Host)
		if masterNode.Hostname != masterNode.Facts.Hostname {
			certNode.SANS = append(certNode.SANS, masterNode.Facts.Hostname)
		}
	}
	certNode.SANS = append(certNode.SANS, config.K8S().ApiServer, config.K8S().ApiServerVIP)

	node.Logger("make kube certs files")

	dir := paths.Join("kube", "pki")
	kubecerts.CreatePKIAssets(dir, certNode)

	//sa
	{
		utils.Panic(node.SudoScp(filepath.Join(dir, "sa.key"), "/etc/kubernetes/pki/sa.key"), "scp sa.key")
		utils.Panic(node.SudoScp(filepath.Join(dir, "sa.pub"), "/etc/kubernetes/pki/sa.pub"), "scp sa")
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
			utils.Panic(node.SudoScp(local, remote), "scp %s %s", local, remote)
		}
	}
}

// make etcdctl command, user friendly.
func makeEtcdctlCommand(node *ssh.Node) {
	node.Logger("create etcdctl command")
	cmdCtx := []byte("#!/bin/bash\n" +
		"kubectl -n kube-system exec etcd-$(hostname -s)" +
		" -- etcdctl" +
		" --cacert=/etc/kubernetes/pki/etcd/ca.crt" +
		" --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt" +
		"  --key=/etc/kubernetes/pki/etcd/healthcheck-client.key $@")

	err := node.SudoScpContent(cmdCtx, "/usr/local/bin/etcdctl")
	utils.Panic(err, "create etcdctl command")

	err = node.SudoCmd("chmod +x /usr/local/bin/etcdctl")
	utils.Panic(err, "change etcdctl model")
}
