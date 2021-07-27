package etcd

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"os"
)

func JoinCluster(node *ssh.Node) {

	if Config.Exists(node.Host) {
		fmt.Printf("%s already in the cluster\n", color.RedString(node.Host))
		return
	}
	master := hosts.Get(Config.Nodes[0])

	bases.Check(node)
	checkEtcdadm(node)
	makeAndPushCerts(node)
	etcdadmJoin(master, node)

	Config.Join(node.Host)
}

func etcdadmJoin(master *ssh.Node, node *ssh.Node) {
	utils.Line("etcdadm join")
	cmd := "etcdadm join --name " + node.Hostname +
		" --install-dir /usr/local/bin " +
		" --certs-dir " + Config.CertsDir +
		" --version " + Config.Version
	/*
		for _, san := range Config.ServerCertExtraSans {
			cmd += " --server-cert-extra-sans " + san
		}
	*/
	cmd += " https://" + master.Host + ":2379"
	err := node.CmdStd(cmd, os.Stdout)
	utils.Panic(err, "etcdadm join")
}
