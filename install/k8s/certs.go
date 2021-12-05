package k8s

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	kubecerts "github.com/ihaiker/vik8s/certs/kubernetes"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"net"
	"path/filepath"
)

func makeKubernetesCerts(configure *config.Configuration, node *ssh.Node) {
	if configure.IsExternalETCD() {
		scpExternalEtcdCa(configure, node)
	} else {
		makeEtcdCerts(configure, node)
		makeEtcdctlCommand(node)
	}
	makeKubeCerts(configure, node)
}

func makeJoinControlPlaneConfigFiles(configure *config.Configuration, node *ssh.Node) {
	dir := paths.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", configure.K8S.ApiServer)
	files := kubecerts.CreateJoinControlPlaneKubeConfigFiles(dir, node.Hostname, endpoint, configure.K8S.CertsValidity)
	for key, path := range files {
		remote := filepath.Join("/etc/kubernetes", fmt.Sprintf("%s.conf", key))
		err := node.Sudo().Scp(path, remote)
		utils.Panic(err, "scp %s %s", path, remote)
	}
}

func makeWorkerConfigFiles(configure *config.Configuration, node *ssh.Node) {
	dir := paths.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", configure.K8S.ApiServer)
	files := kubecerts.CreateWorkerKubeConfigFile(dir, node.Hostname, endpoint, configure.K8S.CertsValidity)
	for key, path := range files {
		remote := filepath.Join("/etc/kubernetes", fmt.Sprintf("%s.conf", key))
		utils.Panic(node.Sudo().Scp(path, remote), "scp %s %s", path, remote)
	}
}

func makeEtcdCerts(configure *config.Configuration, node *ssh.Node) {
	node.Logger("make etcd certs files")

	name := node.Hostname
	dir := paths.Join("kube", "pki", "etcd")

	// local + master + apiserversans + apiserver-vip
	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host, net.IPv6loopback.String()}
	sans = append(sans, utils.ParseIPS(configure.K8S.Masters)...)
	if node.Hostname != node.Facts.Hostname {
		sans = append(sans, node.Facts.Hostname)
	}
	sans = append(sans, configure.K8S.ApiServer, configure.K8S.ApiServerVIP)

	vt := configure.K8S.CertsValidity
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

func scpExternalEtcdCa(configure *config.Configuration, node *ssh.Node) {
	files := map[string]string{}
	if configure.ExternalETCD != nil {
		files[configure.ExternalETCD.CaFile] = "/etc/kubernetes/pki/etcd/ca.crt"
		files[configure.ExternalETCD.Key] = "/etc/kubernetes/pki/etcd/apiserver-etcd-client.key"
		files[configure.ExternalETCD.Cert] = "/etc/kubernetes/pki/etcd/apiserver-etcd-client.crt"
	} else {
		dir := etcd.CertsDir()
		files[filepath.Join(dir, "ca.crt")] = "/etc/kubernetes/pki/etcd/ca.crt"
		files[filepath.Join(dir, "apiserver-etcd-client.key")] = "/etc/kubernetes/pki/etcd/apiserver-etcd-client.key"
		files[filepath.Join(dir, "apiserver-etcd-client.crt")] = "/etc/kubernetes/pki/etcd/apiserver-etcd-client.crt"
	}
	for local, remote := range files {
		err := node.Sudo().Scp(local, remote)
		utils.Panic(err, "scp %s %s", local, remote)
	}
}

func makeKubeCerts(configure *config.Configuration, node *ssh.Node) {
	certNode := kubecerts.Node{
		Name:                node.Hostname,
		Host:                node.Host,
		ApiServer:           configure.K8S.ApiServer,
		SvcCIDR:             configure.K8S.SvcCIDR,
		CertificateValidity: configure.K8S.CertsValidity,
		SANS:                configure.K8S.ApiServerCertExtraSans,
	}
	//apisever = MasterIPS + VIP + CertSANS
	for _, masterIp := range configure.K8S.Masters {
		masterNode := configure.Hosts.MustGet(masterIp)
		certNode.SANS = append(certNode.SANS, masterNode.Hostname, masterNode.Host)
		if masterNode.Hostname != masterNode.Facts.Hostname {
			certNode.SANS = append(certNode.SANS, masterNode.Facts.Hostname)
		}
	}
	certNode.SANS = append(certNode.SANS, configure.K8S.ApiServer, configure.K8S.ApiServerVIP)

	node.Logger("make kube certs files")

	dir := paths.Join("kube", "pki")
	kubecerts.CreatePKIAssets(configure.K8S.ApiServerVIP, dir, certNode)

	//sa
	{
		utils.Panic(node.Sudo().Scp(filepath.Join(dir, "sa.key"), "/etc/kubernetes/pki/sa.key"), "scp sa.key")
		utils.Panic(node.Sudo().Scp(filepath.Join(dir, "sa.pub"), "/etc/kubernetes/pki/sa.pub"), "scp sa")
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
			utils.Panic(node.Sudo().Scp(local, remote), "scp %s %s", local, remote)
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

	err := node.Sudo().ScpContent(cmdCtx, "/usr/local/bin/etcdctl")
	utils.Panic(err, "create etcdctl command")

	err = node.Sudo().Cmd("chmod +x /usr/local/bin/etcdctl")
	utils.Panic(err, "change etcdctl model")
}
