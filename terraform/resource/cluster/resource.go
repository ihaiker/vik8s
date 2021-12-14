package cluster

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/terraform/tools"
	"io/ioutil"
	"time"
)

func Vik8sCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: tools.Logger("create", createClusterNodeContext),
		ReadWithoutTimeout:   tools.Logger("read", readClusterNodeContext),
		UpdateWithoutTimeout: tools.Logger("update", updateClusterNodeContext),
		DeleteWithoutTimeout: tools.Logger("delete", deleteClusterNodeContext),
		Schema:               vik8sClusterSchema(),
	}
}

func setDataNode(data *schema.ResourceData, configure *Configuration) (err error) {
	if err = data.Set("masters", configure.Config.Masters); err != nil {
		return
	}
	if err = data.Set("slaves", configure.Config.Nodes); err != nil {
		return
	}
	return
}

func output(data *schema.ResourceData, master *ssh.Node, cfg *config.Configuration, configure *Configuration) (err error) {
	certs := []string{
		"kube/pki/sa.key", "kube/pki/sa.pub",
		"kube/pki/ca.key", "kube/pki/ca.crt",
		"kube/pki/front-proxy-ca.key", "kube/pki/front-proxy-ca.crt",
	}
	if cfg.ExternalETCD == nil && cfg.ETCD == nil {
		certs = append(certs,
			"kube/pki/etcd/ca.key", "kube/pki/etcd/ca.crt",
			"kube/pki/etcd/etcdctl-etcd-client.key", "kube/pki/etcd/etcdctl-etcd-client.crt",
			"kube/pki/etcd/apiserver-etcd-client.key", "kube/pki/etcd/apiserver-etcd-client.crt",
			"kube/pki/etcd/healthcheck-client.key", "kube/pki/etcd/healthcheck-client.crt",
		)
	}
	var fs []byte
	for _, cert := range certs {
		if fs, err = ioutil.ReadFile(paths.Join(cert)); err != nil {
			return
		} else {
			configure.Certificates[cert] = string(fs)
		}
	}

	configure.ApiServerUrl = fmt.Sprintf("https://%s:6443", cfg.K8S.ApiServer)
	if configure.ClusterYaml, err = master.CmdString("cat ~/.kube/config"); err != nil {
		return
	}

	data.SetId(tools.Id("cluster", configure))
	configure.UpdateTime = time.Now().Format(time.RFC3339)

	return setDataNode(data, configure)
}

func createClusterNodeContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) error {
	configure, err := expendVIK8SCluster(data)
	if err != nil {
		return err
	}
	cfg.K8S = configure.Config

	masters := cfg.Hosts.MustGets(configure.Masters)
	slaves := cfg.Hosts.MustGets(configure.Slaves)

	master := masters[0]
	k8s.ResetNode(cfg, master)
	k8s.InitCluster(cfg, master)
	cfg.K8S.JoinNode(true, master.Host)

	kubeadm, _ := master.Sudo().HideLog().CmdBytes("cat " + master.Vik8s("apply/kubeadm.yaml"))

	if err = output(data, master, cfg, configure); err != nil {
		return err
	}

	for _, master = range masters[1:] {
		k8s.ResetNode(cfg, master)
		k8s.JoinControl(cfg, master)
		cfg.K8S.JoinNode(true, master.Host)
		if err = master.Sudo().HideLog().ScpContent(kubeadm.Bytes(), master.Vik8s("apply/kubeadm.yaml")); err != nil {
			return err
		}
	}

	if configure.Network.Flannel != nil {
		logs.Info("install network flannel")
		configure.Network.Flannel.Apply(cfg, masters[0])
	} else if configure.Network.Calico != nil {
		logs.Info("install network calico")
		configure.Network.Calico.Apply(cfg, masters[0])
	} else if configure.Network.Customer != nil {
		logs.Info("install network customer")
		configure.Network.Customer.Apply(cfg, masters[0])
	}

	for _, slave := range slaves {
		k8s.ResetNode(cfg, slave)
		k8s.JoinWorker(cfg, slave)
		cfg.K8S.JoinNode(false, slave.Host)
	}

	return flattenVIK8SCluster(data, configure)
}

func readClusterNodeContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) error {
	configure, err := expendVIK8SCluster(data)
	if err != nil {
		return err
	}
	cfg.K8S = configure.Config
	cfg.K8S.Masters = configure.Masters
	cfg.K8S.Nodes = configure.Slaves
	return flattenVIK8SCluster(data, configure)
}

func changeNodes(data *schema.ResourceData, name string) (add, keep, remove []string) {
	if !data.HasChange(name) {
		keep = tools.GetDataSetString(data.Get(name))
		return
	}

	o, n := data.GetChange(name)
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

func updateClusterNodeContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) error {
	configure, err := expendVIK8SCluster(data)
	if err != nil {
		return err
	}
	cfg.K8S = configure.Config
	addMasters, keepMasters, removeMasters := changeNodes(data, "masters")
	cfg.K8S.Masters = append(keepMasters, removeMasters...)
	master := cfg.Hosts.MustGet(cfg.K8S.Masters[0])
	kubeadm, _ := master.Sudo().HideLog().CmdBytes("cat " + master.Vik8s("apply/kubeadm.yaml"))

	for _, node := range cfg.Hosts.MustGets(addMasters) {
		_ = master.Cmd(fmt.Sprintf("kubectl delete nodes %s", node.Hostname))
		k8s.ResetNode(cfg, node)
		k8s.JoinControl(cfg, node)
		cfg.K8S.JoinNode(true, node.Host)
		if err = node.Sudo().HideLog().ScpContent(kubeadm.Bytes(), master.Vik8s("apply/kubeadm.yaml")); err != nil {
			return err
		}
	}

	for _, node := range cfg.Hosts.MustGets(removeMasters) {
		_ = master.Cmd(fmt.Sprintf("kubectl delete nodes %s", node.Hostname))
		k8s.ResetNode(cfg, node)
		cfg.K8S.RemoveNode(node.Host)
	}

	addSlaves, keepSlaves, removeSlaves := changeNodes(data, "slaves")
	cfg.K8S.Nodes = append(keepSlaves, removeSlaves...)
	for _, node := range cfg.Hosts.MustGets(addSlaves) {
		_ = master.Cmd(fmt.Sprintf("kubectl delete nodes %s", node.Hostname))
		k8s.ResetNode(cfg, node)
		k8s.JoinWorker(cfg, node)
		cfg.K8S.JoinNode(false, node.Host)
	}
	for _, node := range cfg.Hosts.MustGets(removeSlaves) {
		_ = master.Cmd(fmt.Sprintf("kubectl delete nodes %s", node.Hostname))
		k8s.ResetNode(cfg, node)
		cfg.K8S.RemoveNode(node.Host)
	}

	return flattenVIK8SCluster(data, configure)
}

func deleteClusterNodeContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) error {
	logs.Debug("delete cluster node: ", data.Id())
	configure, err := expendVIK8SCluster(data)
	if err != nil {
		return err
	}
	cfg.K8S = configure.Config
	cfg.K8S.Masters = configure.Masters
	cfg.K8S.Nodes = configure.Slaves

	slaves := cfg.Hosts.MustGets(cfg.K8S.Nodes)
	for _, slave := range slaves {
		k8s.ResetNode(cfg, slave)
	}

	masters := cfg.Hosts.MustGets(cfg.K8S.Masters)
	for _, master := range masters {
		k8s.ResetNode(cfg, master)
	}
	return nil
}
