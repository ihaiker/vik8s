package k8s

import (
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
)

func Clean(nodes []*ssh.Node, expFn ...func(node *ssh.Node)) {
	ssh.Sync(nodes, func(i int, node *ssh.Node) {
		for _, fn := range expFn {
			fn(node)
		}

		_ = node.Sudo().CmdStdout("kubeadm reset -f")
		_ = node.Sudo().CmdStdout("ipvsadm -C")
		_ = node.Sudo().CmdStdout("ifconfig cni0 down")
		_ = node.Sudo().CmdStdout("ip link delete cni0")
		_ = node.Sudo().CmdStdout("ip link delete kube-ipvs0")
		_ = node.Sudo().CmdStdout("ip link delete dummy0")

		_ = node.Sudo().CmdStdout("rm -rf /etc/cni/net.d/* ~/.kube /etc/kubernetes/*")
		_ = node.Sudo().CmdStdout("rm -rf /var/lib/etcd")
		_ = node.Sudo().CmdStdout("rm -rf /var/lib/ceph")

		_ = node.Shell(`
			iptables -P INPUT ACCEPT
			iptables -P FORWARD ACCEPT
			iptables -P OUTPUT ACCEPT
			iptables -t nat -F
			iptables -t mangle -F
			iptables -F
			iptables -X
			
			ip6tables -P INPUT ACCEPT
			ip6tables -P FORWARD ACCEPT
			ip6tables -P OUTPUT ACCEPT
			ip6tables -t nat -F
			ip6tables -t mangle -F
			ip6tables -F
			ip6tables -X
		`, utils.Stdout(""))
	})
	config := paths.Join("vik8s.conf")
	utils.Panic(os.RemoveAll(config), "remove %s", config)

	kube := paths.Join("kube")
	utils.Panic(os.RemoveAll(kube), "remove %s", kube)
}
