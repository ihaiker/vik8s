package hosts

import (
	"github.com/ihaiker/ngx/v2"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Option = ssh.Node

type Manager struct {
	path  string
	opt   Option
	Nodes ssh.Nodes `ngx:"node"`
}

func New(path string, opt Option) (*Manager, error) {
	manager := &Manager{
		path: path, opt: opt,
		Nodes: ssh.Nodes{},
	}
	err := manager.load()
	return manager, err
}

func (this *Manager) load() error {
	if utils.NotExists(this.path) {
		return nil
	}

	if context, err := ioutil.ReadFile(this.path); err != nil {
		return err
	} else if err = ngx.Unmarshal(context, this); err != nil {
		return err
	}

	for _, node := range this.Nodes {
		_ = this.getProxy(node)
	}
	return nil
}
func (this *Manager) down() error {
	dir := filepath.Dir(this.path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	if out, err := ngx.Marshal(this); err != nil {
		return err
	} else {
		return ioutil.WriteFile(this.path, out, 0644)
	}
}

func (this *Manager) All() ssh.Nodes {
	return this.Nodes
}

func (this *Manager) getProxy(node *ssh.Node) error {
	if node.Proxy != "" {
		if proxyNode := this.Get(node.Proxy); proxyNode != nil {
			node.ProxyNode = proxyNode
		} else {
			return utils.Error("the ssh proxy %s not found", node.Proxy)
		}
	}
	return nil
}

func (this *Manager) Add(node *ssh.Node) error {
	defer this.down()
	for i, n := range this.Nodes {
		if n.Host == node.Host {
			this.Nodes[i] = node
			return nil
		}
	}
	this.Nodes = append(this.Nodes, node)
	return nil
}

func (this *Manager) Get(args string) *ssh.Node {
	return this.Nodes.Get(args)
}

func (this *Manager) FetchNode(overwrite bool, nodes ...*ssh.Node) (ssh.Nodes, error) {
	for i, node := range nodes {
		oldNode := this.Get(node.Host)
		if oldNode == nil || overwrite { //当前节点不在列表中，或者需要覆盖
			if err := node.Sudo().HideLog().GatheringFacts(); err != nil {
				return nil, utils.Wrap(err, "gathering facts")
			}
			_ = this.Add(node)
		} else {
			nodes[i] = oldNode
		}
	}
	return nodes, nil
}

// Fetch 配置
func (this *Manager) Fetch(overwrite bool, args ...string) (ssh.Nodes, error) {
	nodes := ssh.Nodes{}
	for _, arg := range args {
		if node := this.Get(arg); node != nil {
			nodes = append(nodes, node)
		} else if ns, err := ParseAddr(this.opt, arg); err == nil {
			for _, node = range ns {
				if err = this.getProxy(node); err != nil {
					return nil, err
				}
			}
			nodes = append(nodes, ns...)
		} else {
			return nil, utils.Error("not found node: %s", arg)
		}
	}
	return this.FetchNode(overwrite, nodes...)
}

func (this *Manager) MustGet(arg string) *ssh.Node {
	node := this.Get(arg)
	utils.Assert(node != nil, "not found %s", arg)
	err := node.GatheringFacts()
	utils.Panic(err, "gathering facts: %s", node.Host)
	return node
}

func (this *Manager) MustGets(args []string) ssh.Nodes {
	nodes, err := this.Fetch(false, args...)
	utils.Panic(err, "fetch nodes: %s", strings.Join(args, ", "))
	for _, node := range nodes {
		err = node.GatheringFacts()
		utils.Panic(err, "gathering facts: %s", node.Host)
	}
	return nodes
}

func (this *Manager) Gets(args []string) (ssh.Nodes, error) {
	return this.Fetch(false, args...)
}
