package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"path/filepath"
	"strings"
)

type (
	Node struct {
		Host       string `ngx:"host"`
		Port       string `ngx:"port"`
		User       string `ngx:"user"`
		Password   string `ngx:"password"`
		PrivateKey string `ngx:"private-key"`
		Passphrase string `ngx:"passphrase"`

		Hostname string `ngx:"hostname"`

		Proxy     string `ngx:"proxy"`
		ProxyNode *Node  `ngx:"-"`

		Facts Facts `ngx:"-"`
	}
	Facts struct {
		ReleaseName   string `ngx:"releaseName"`
		MajorVersion  string `ngx:"majorVersion"`
		KernelVersion string `ngx:"kernelVersion"`
	}
	Nodes []*Node
)

func (node *Node) easyssh() *easySSHConfig {
	config := &easySSHConfig{
		sshConfig: sshConfig{
			User: node.User, Server: node.Host, Port: node.Port,
			KeyPath: node.PrivateKey, Passphrase: node.Passphrase,
			Password: node.Password,
		},
	}
	if node.Proxy != "" {
		config.Proxy = &sshConfig{
			User: node.ProxyNode.User, Server: node.ProxyNode.Host, Port: node.ProxyNode.Port,
			KeyPath: node.ProxyNode.PrivateKey, Passphrase: node.ProxyNode.Passphrase,
			Password: node.ProxyNode.Password,
		}
	}
	return config
}

func (node *Node) GatheringFacts() error {
	if hostname, err := node.cmd("hostname -s", false); err != nil {
		return err
	} else {
		node.Hostname = string(hostname)
	}
	if distribution, err := node.cmd("uname -r", false); err != nil {
		return err
	} else {
		if strings.Index(string(distribution), "el7") != -1 {
			node.Facts.MajorVersion = "7"
		} else if strings.Index(string(distribution), "el8") != -1 {
			node.Facts.MajorVersion = "8"
		}
		node.Facts.KernelVersion = strings.Split(string(distribution), "-")[0]
	}
	if releaseName, err := node.cmd("cat /etc/redhat-release  | awk '{printf $1}'", false); err != nil {
		return err
	} else {
		node.Facts.ReleaseName = string(releaseName)
	}
	return nil
}

func (node *Node) Address() string {
	return fmt.Sprintf("%s:%s", node.Host, node.Port)
}

func (node *Node) HomeDir(join ...string) string {
	if node.User == "root" {
		return filepath.Join(append([]string{"/root"}, join...)...)
	} else {
		return filepath.Join(append([]string{"/home", node.User}, join...)...)
	}
}

func (node *Node) Vik8s(join ...string) string {
	return node.HomeDir(append([]string{".vik8s"}, join...)...)
}

func (node *Node) Logger(format string, params ...interface{}) {
	fmt.Printf("[%s,%s] ", color.RedString(node.Hostname), color.HiGreenString(node.Address()))
	fmt.Printf(format, params...)
	fmt.Println()
}

func (nodes *Nodes) TryGet(hostnameOrIP string) *Node {
	for _, node := range *nodes {
		if node.Hostname == hostnameOrIP || node.Host == hostnameOrIP {
			return node
		}
	}
	return nil
}

func (nodes *Nodes) Get(hostnameOrIP string) *Node {
	node := nodes.TryGet(hostnameOrIP)
	utils.Assert(node != nil, "not found node %s", hostnameOrIP)
	return node
}
