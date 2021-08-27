package ssh

import (
	"bufio"
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

		flag int `ngx:"-"`
	}
	Facts struct {
		Hostname      string `ngx:"hostname"`
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
			Password: node.Password, Timeout: time.Second * 3,
		},
	}
	if node.Proxy != "" {
		config.Proxy = &sshConfig{
			User: node.ProxyNode.User, Server: node.ProxyNode.Host, Port: node.ProxyNode.Port,
			KeyPath: node.ProxyNode.PrivateKey, Passphrase: node.ProxyNode.Passphrase,
			Password: node.ProxyNode.Password, Timeout: time.Second * 3,
		}
	}
	return config
}

func (node *Node) GatheringFacts() error {
	node.Logger("gathering facts")

	if hostname, err := node.Sudo().HideLog().CmdString("hostname -f"); err != nil {
		return err
	} else {
		idx := strings.Index(hostname, ".")
		if idx == -1 {
			node.Hostname = hostname
		} else {
			node.Hostname = hostname[0:idx]
		}
		node.Facts.Hostname = hostname
	}

	envMaps := make(map[string]string)
	if envs, err := node.Sudo().HideLog().CmdBytes("cat /etc/os-release"); err != nil {
		return err
	} else {
		lineReader := bufio.NewReader(envs)
		for line, _, err := lineReader.ReadLine(); err == nil; line, _, err = lineReader.ReadLine() {
			key, value := utils.Split2(string(line), "=")
			if unquoteValue, err := strconv.Unquote(value); err == nil {
				envMaps[key] = unquoteValue
			} else {
				envMaps[key] = value
			}
		}
	}

	node.Facts.ReleaseName = envMaps["ID"]
	node.Facts.MajorVersion = envMaps["VERSION_ID"]
	distribution, err := node.Sudo().HideLog().CmdBytes("uname -r")
	if err != nil {
		return err
	}
	node.Facts.KernelVersion = strings.Split(distribution.String(), "-")[0]

	node.Logger("facts: hostname: %s, release: %s, major: %s, kernel: %s",
		node.Facts.Hostname, node.Facts.ReleaseName, node.Facts.MajorVersion, node.Facts.KernelVersion)

	utils.Assert(node.Facts.Hostname != "", "gathering facts hostname")
	utils.Assert(node.Facts.ReleaseName != "", "gathering facts release name")
	utils.Assert(node.Facts.MajorVersion != "", "gathering facts major version")
	utils.Assert(node.Facts.KernelVersion != "", "gathering facts kernel version")
	return nil
}

func (node *Node) IsUbuntu() bool {
	return strings.ToLower(node.Facts.ReleaseName) == "ubuntu"
}

func (node *Node) IsCentOS() bool {
	return strings.ToLower(node.Facts.ReleaseName) == "centos"
}

func (node *Node) IsRoot() bool {
	return node.User == "root"
}

func (node *Node) HomeDir(join ...string) string {
	if node.IsRoot() {
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
