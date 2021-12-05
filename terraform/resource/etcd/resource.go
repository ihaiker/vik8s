package etcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"io/ioutil"
	"strings"
)

func Vik8sETCD() *schema.Resource {
	return &schema.Resource{
		Schema:               etcdSchema(),
		CreateWithoutTimeout: tools.Logger("etcd create", createEtcdContext),
		ReadWithoutTimeout:   tools.Logger("etcd read", readEtcdContext),
		UpdateWithoutTimeout: tools.Logger("etcd update", updateEtcdContext),
		DeleteWithoutTimeout: tools.Logger("etcd delete", deleteEtcdContext),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func createEtcdContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) (err error) {
	var etcdConfig *etcdSchemaConfig
	if etcdConfig, err = expendEtcd(data); err != nil {
		return
	}

	logs.Info("create etcd cluster")

	cfg.ETCD = etcdConfig.ETCDConfiguration
	sshNodes := cfg.Hosts.MustGets(cfg.ETCD.Nodes)
	cfg.ETCD.Nodes = make([]string, 0)

	for i, node := range sshNodes {
		if i == 0 {
			logs.Info("init etcd cluster")
			etcd.InitCluster(cfg, node)
		} else {
			logs.Info("join etcd node: ", node.Host)
			etcd.JoinCluster(cfg, node)
		}
		cfg.ETCD.Nodes = append(cfg.ETCD.Nodes, node.Host)
	}

	certs := []string{
		"etcd/pki/ca.crt", "etcd/pki/ca.key",
		"etcd/pki/etcdctl-etcd-client.crt", "etcd/pki/etcdctl-etcd-client.key",
		"etcd/pki/apiserver-etcd-client.crt", "etcd/pki/apiserver-etcd-client.key",
		"etcd/pki/healthcheck-client.crt", "etcd/pki/healthcheck-client.key",
	}
	logs.Info("out etcd cluster certs")
	var fs []byte
	for _, cert := range certs {
		if fs, err = ioutil.ReadFile(paths.Join(cert)); err != nil {
			return
		} else {
			etcdConfig.certs[cert] = string(fs)
		}
	}

	data.SetId(tools.Id("etcd", etcdConfig))
	return tools.SetState(flattenEtcd(etcdConfig), data)
}

func deleteEtcdContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) (err error) {
	var etcdConfig *etcdSchemaConfig
	if etcdConfig, err = expendEtcd(data); err != nil {
		return
	}
	cfg.ETCD = etcdConfig.ETCDConfiguration

	logs.Info("remove nodes: ", strings.Join(cfg.ETCD.Nodes, ","))
	sshNodes := cfg.Hosts.MustGets(cfg.ETCD.Nodes)
	for _, node := range sshNodes {
		etcd.ResetCluster(cfg, node)
		cfg.ETCD.RemoveNode(node.Host)
	}
	return
}

func changeNodes(data *schema.ResourceData) (add, keep, remove []string) {
	if !data.HasChange("nodes") {
		keep = tools.GetDataSetString(data.Get("nodes"))
		return
	}

	o, n := data.GetChange("nodes")
	oldNode := o.(*schema.Set)
	newNodes := n.(*schema.Set)

	removeSet := oldNode.Difference(newNodes)
	addSet := newNodes.Difference(oldNode)
	keepSet := oldNode.Intersection(newNodes)

	remove = tools.GetDataSetString(removeSet)
	add = tools.GetDataSetString(addSet)
	keep = tools.GetDataSetString(keepSet)
	return
}

func updateEtcdContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) (err error) {
	var etcdConfig *etcdSchemaConfig
	if etcdConfig, err = expendEtcd(data); err != nil {
		return
	}
	adds, keep, removes := changeNodes(data)

	cfg.ETCD = etcdConfig.ETCDConfiguration
	cfg.ETCD.Nodes = append(keep, removes...)

	removeNodes := getNodes(cfg.Hosts, removes)
	for _, node := range removeNodes {
		etcd.ResetCluster(cfg, node)
		cfg.ETCD.RemoveNode(node.Host)
	}

	if data.HasChange("version") {
		logs.Info("update new version: ", cfg.ETCD.Version)
	}

	addNodes := cfg.Hosts.MustGets(adds)
	for _, node := range addNodes {
		etcd.JoinCluster(cfg, node)
	}

	if data.HasChanges("snapshot", "remote_snapshot") {
		logs.Info("snapshot update")
	}

	cfg.ETCD = etcdConfig.ETCDConfiguration
	return tools.SetState(flattenEtcd(etcdConfig), data)
}

func getNodes(manager *hosts.Manager, itemes []string) ssh.Nodes {
	nodes := make([]*ssh.Node, 0)
	for _, item := range itemes {
		_ = utils.Safe(func() {
			nodes = append(nodes, manager.MustGet(item))
		})
	}
	return nodes
}

func readEtcdContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) error {
	etcdConfig, err := expendEtcd(data)
	if err != nil {
		return err
	}
	cfg.ETCD = etcdConfig.ETCDConfiguration
	return tools.SetState(flattenEtcd(etcdConfig), data)
}
