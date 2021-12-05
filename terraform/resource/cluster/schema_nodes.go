package cluster

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"log"
	"os"
	"time"
)

var (
	TaintEffectTypes = []string{"NoExecute", "NoSchedule", "PreferNoSchedule"}
)

type nodeConfig struct {
	role   []string
	labels map[string]string
	nodes  ssh.Nodes
}

type nodesConfig []*nodeConfig

func nodeSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		Optional:    true,
		Description: "node config",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address": {
					Type:        schema.TypeString,
					Description: "ssh address, your can use ip range. example: 192.168.0.10-24, 192.168.0.10-1.25",
					Required:    true,
				},
				"port": {
					Type:        schema.TypeInt,
					Description: "ssh port",
					Optional:    true,
					Default:     22,
				},
				"username": {
					Type:        schema.TypeString,
					Default:     "root",
					Optional:    true,
					Description: `ssh username`,
				},
				"password": {
					Type:        schema.TypeString,
					Description: `ssh password`,
					Optional:    true,
					Sensitive:   true,
					Default:     "",
				},
				"ssh_key": {
					Type:        schema.TypeString,
					Description: "ssh private key",
					Optional:    true,
					Default:     "",
				},
				"ssh_key_raw": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "SSH Private Key",
					Default:     "",
				},
				"passphrase": {
					Type:        schema.TypeString,
					Description: "private key passphrase",
					Optional:    true,
					Sensitive:   true,
					Default:     "",
				},
				"bastion": {
					Type:        schema.TypeString,
					Description: "bastion Host configuration server",
					Optional:    true,
					Default:     "",
				},
				"role": {
					Type:        schema.TypeSet,
					Optional:    true,
					Required:    false,
					Description: "Node roles in k8s cluster [control_plane/worker/etcd/bastion])",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"control_plane", "etcd", "worker", "bastion"}, true),
					},
				},
				"labels": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Node Labels",
					Default:     map[string]string{},
				},
				"taints": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Node taints",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key": {
								Type:     schema.TypeString,
								Required: true,
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
							"effect": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      "NoExecute",
								ValidateFunc: validation.StringInSlice(TaintEffectTypes, true),
							},
						},
					},
				},
			},
		},
	}
}

func expendNodes(p interface{}) (_nodes nodesConfig, err error) {
	in := tools.ListItemsWrapper(p)
	_nodes = make([]*nodeConfig, 0)
	for _, item := range in {
		sshKey := item.String("ssh_key", "")
		if sshKey != "" {
			sshKey = os.ExpandEnv(sshKey)
			if utils.NotExists(sshKey) {
				return nil, fmt.Errorf("ssh key %s not exists", sshKey)
			}
		}
		password := item.String("password", "")
		sshKeyRaw := item.String("ssh_key_raw", "")

		if sshKey == "" && password == "" && sshKeyRaw == "" {
			err = fmt.Errorf("at least one of `password`,`ssh_key`,`ssh_key_raw` ")
			return
		}
		bastion := item.String("bastion", "")
		_node := &nodeConfig{
			role:   item.Set("role", []string{"worker"}),
			labels: item.Map("labels", map[string]string{}),
		}
		opt := hosts.Option{
			User: item.String("username", ""),
			Host: item.String("address", ""),
			Port: item.Int("port"), Proxy: bastion,
			Password: password, PrivateKey: sshKey, PrivateKeyRaw: sshKeyRaw,
			Passphrase: item.String("passphrase", ""),
		}
		addr := item.String("address", "")
		if hosts.PATTERN.MatchString(addr) {
			if _node.nodes, err = hosts.ParseAddr(opt, addr); err != nil {
				err = utils.Wrap(err, "parse address: %s", addr)
				return
			}
		} else {
			opt.Host = addr
			_node.nodes = append(_node.nodes, &opt)
		}
		_nodes = append(_nodes, _node)
	}
	return
}

//checkAllNodes 检查所有主机是否可连接
func checkAllNodes(configure *config.Configuration, nodes nodesConfig) (err error) {
	log.Println("check all nodes")
	for _, sn := range nodes {
		for _, sshNode := range sn.nodes {
			log.Println("check host: ", sshNode.Host)
			if sshNode.Proxy != "" {
				if sshNode.ProxyNode, err = nodes.get(sshNode.Proxy); err != nil {
					return
				}
			}
			sshNode.Timeout = time.Second * 3
			if err = sshNode.GatheringFacts(); err != nil {
				return
			}
			_ = configure.Hosts.Add(sshNode)
		}
	}
	return
}

func (this nodesConfig) get(host string) (*ssh.Node, error) {
	for _, _node := range this {
		if _node.nodes == nil {
			continue
		}
		if p := _node.nodes.Get(host); p != nil {
			return p, nil
		}
	}
	return nil, fmt.Errorf("host %s not found", host)
}

func (this nodesConfig) roleNode(role string) ssh.Nodes {
	items := ssh.Nodes{}
	for _, defNode := range this {
		if utils.Search(defNode.role, role) != -1 {
			items = append(items, defNode.nodes...)
		}
	}
	return items
}
