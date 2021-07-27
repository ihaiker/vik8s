package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"path/filepath"
	"strconv"
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
	if hostname, err := node.SudoCmdString("hostname -s"); err != nil {
		return err
	} else {
		node.Hostname = hostname
	}

	envMaps := make(map[string]string)
	if envs, err := node.SudoCmdString("cat /etc/os-release"); err != nil {
		return err
	} else {
		envLines := strings.Split(envs, "\n")
		for _, envLine := range envLines {
			keyAndVal := strings.Split(envLine, "=")
			if unquoteValue, err := strconv.Unquote(keyAndVal[1]); err == nil {
				envMaps[keyAndVal[0]] = unquoteValue
			} else {
				envMaps[keyAndVal[0]] = keyAndVal[1]
			}
		}
	}
	node.Facts.ReleaseName = envMaps["ID"]
	node.Facts.MajorVersion = envMaps["VERSION_ID"]
	distribution, err := node.SudoCmdBytes("uname -r")
	if err != nil {
		return err
	}
	node.Facts.KernelVersion = strings.Split(distribution.String(), "-")[0]
	return nil
}

func (node *Node) IsUbuntu() bool {
	return strings.ToLower(node.Facts.ReleaseName) == "ubuntu"
}

func (node *Node) IsCentOS() bool {
	return strings.ToLower(node.Facts.ReleaseName) == "centos"
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

func (node *Node) Prefix() string {
	addr := fmt.Sprintf("%s:%s", node.Host, node.Port)
	if node.Hostname != "" {
		addr = node.Hostname
	}
	return fmt.Sprintf("[%s@%s]", color.RedString(node.User), color.HiGreenString(addr))
}

func (node *Node) Logger(format string, params ...interface{}) {
	fmt.Print(node.Prefix(), " ")
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
