package cluster

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"log"
)

func crateEtcd(configure *config.Configuration, nodes nodesConfig, data *schema.ResourceData) (err error) {

	if configure.ExternalETCD != nil {
		log.Println("use external etcd, it will ignore etcd node")
		return
	}

	etcdNodes := nodes.roleNode("etcd")
	if len(etcdNodes) == 0 && configure.ETCD == nil {
		log.Println("use internal etcd")
		return
	}
	if len(etcdNodes) == 0 {
		return utils.Error("not found etcd nodes")
	}

	log.Println("create external etcd cluster")
	if configure.ETCD == nil {
		configure.ETCD = config.DefaultETCDConfiguration()
	}

	master := etcdNodes[0]
	if err = initCluster(configure, master); err != nil {
		return
	}
	configure.ETCD.Nodes = append(configure.ETCD.Nodes, master.Host)
	/*certs := []string{
		"etcd/pki/apiserver-etcd-client.crt", "etcd/pki/apiserver-etcd-client.key",
		"etcd/pki/ca.crt", "etcd/pki/ca.key",
		"etcd/pki/etcdctl-etcd-client.crt", "etcd/pki/etcdctl-etcd-client.key",
		"etcd/pki/healthcheck-client.crt", "etcd/pki/healthcheck-client.key",
	}

	var fileContext []byte
	for _, cert := range certs {
		if fileContext, err = ioutil.ReadFile(paths.Join(cert)); err != nil {
			return
		} else {
			configure.Certificates[cert] = string(fileContext)
		}
	}*/

	for _, node := range etcdNodes[1:] {
		if err = joinCluster(configure, node); err != nil {
			return
		}
		configure.ETCD.Nodes = append(configure.ETCD.Nodes, node.Host)
	}
	return
}

func updateEtcd(configure *config.Configuration, nodes nodesConfig, data *schema.ResourceData) (err error) {
	if !data.HasChange("nodes") {
		log.Println("no node changed")
		return
	}
	oldNodes, _ := data.GetChange("nodes")
	var oldNodesConfig nodesConfig
	if oldNodesConfig, err = expendNodes(oldNodes); err != nil {
		_ = data.Set("nodes", oldNodesConfig)
		return
	}

	return
}

func initCluster(configure *config.Configuration, _node *ssh.Node) (err error) {
	defer utils.Catch(func(e error) {
		log.Println(utils.Stack())
		err = utils.Wrap(e, "init etcd %s", _node.Host)
	})
	log.Println("init etcd ", _node.Host)
	etcd.InitCluster(configure, _node)
	return
}

func joinCluster(configure *config.Configuration, node *ssh.Node) (err error) {
	defer utils.Catch(func(e error) {
		log.Println(utils.Stack())
		err = utils.Wrap(e, "join etcd %s", node.Host)
	})
	log.Println("join etcd ", node.Host)
	etcd.JoinCluster(configure, node)
	return
}
