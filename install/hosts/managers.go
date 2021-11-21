package hosts

import (
	"github.com/ihaiker/ngx/v2"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Option = ssh.Node

type Manager interface {
	All() ssh.Nodes
	Add(node *ssh.Node) error
	Get(hostnameOrIp string) *ssh.Node
}

func New(path string) (manager *defaultManager, err error) {
	manager = &defaultManager{path: path}
	err = manager.load()
	return
}

type defaultManager struct {
	path  string
	Nodes ssh.Nodes `ngx:"node"`
}

func (this *defaultManager) load() error {
	if utils.Exists(this.path) {
		if bodys, err := ioutil.ReadFile(this.path); err != nil {
			return err
		} else {
			if err = ngx.Unmarshal(bodys, this); err != nil {
				return err
			}
		}
	}
	for _, node := range this.Nodes {
		_ = getProxy(this, node)
	}
	return nil
}

func (this *defaultManager) down() error {
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

func (this *defaultManager) All() ssh.Nodes {
	return this.Nodes
}

func getProxy(this Manager, node *ssh.Node) error {
	if node.Proxy != "" {
		if proxyNode := this.Get(node.Proxy); proxyNode != nil {
			node.ProxyNode = proxyNode
		} else {
			return utils.Error("the ssh proxy %s not found", node.Proxy)
		}
	}
	return nil
}

func (this *defaultManager) Add(node *ssh.Node) error {
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

func (this *defaultManager) Get(args string) *ssh.Node {
	return this.Nodes.Get(args)
}
