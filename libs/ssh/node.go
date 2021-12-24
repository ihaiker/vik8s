package ssh

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/utils"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type (
	Node struct {
		Host          string `ngx:"host" flag:"-"`
		Port          int    `ngx:"port" short:"p" help:"ssh port" def:"22"`
		User          string `ngx:"user" short:"u" help:"ssh user" def:"root"`
		Password      string `ngx:"password" short:"P" help:"ssh password"`
		PrivateKey    string `ngx:"private-key" short:"i" help:"ssh private key" def:"$HOME/.ssh/id_rsa"`
		PrivateKeyRaw string `ngx:"private-key-raw" help:"ssh private key"`
		Passphrase    string `ngx:"passphrase" flag:"passphrase" help:"private key passphrase"`

		Hostname string `ngx:"hostname" flag:"-"`

		Proxy     string `ngx:"proxy"`
		ProxyNode *Node  `ngx:"-" flag:"-"`

		Facts Facts `ngx:"-" flag:"-"`

		flag    int `ngx:"-" flag:"-"`
		retries int `ngx:"-" flag:"-"`

		Timeout           time.Duration `ngx:"-" flag:"-"`
		Ciphers           []string      `ngx:"-" flag:"-"`
		KeyExchanges      []string      `ngx:"-" flag:"-"`
		Fingerprint       string        `ngx:"-" flag:"-"`
		UseInsecureCipher bool          `ngx:"-" flag:"-"`
	}
	Facts struct {
		Hostname      string `ngx:"hostname"`
		ReleaseName   string `ngx:"releaseName"`
		MajorVersion  string `ngx:"majorVersion"`
		KernelVersion string `ngx:"kernelVersion"`
	}
	Nodes []*Node
)

func (node *Node) GatheringFacts() error {
	if node.Facts.ReleaseName != "" && node.Facts.MajorVersion != "" && node.Facts.KernelVersion != "" {
		return nil
	}

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
	addr := fmt.Sprintf("%s:%d", node.Host, node.Port)
	if node.Hostname != "" {
		addr = node.Hostname
	}
	return fmt.Sprintf("[%s@%s]", color.RedString(node.User), color.HiGreenString(addr))
}

func (node *Node) Logger(format string, params ...interface{}) {
	logs.Info(node.Prefix(), " ", fmt.Sprintf(format, params...))
}

func (nodes *Nodes) Get(hostnameOrIP string) *Node {
	for _, node := range *nodes {
		if node.Hostname == hostnameOrIP || node.Host == hostnameOrIP {
			return node
		}
	}
	return nil
}

func (nodes Nodes) Hosts() []string {
	nodeHosts := make([]string, len(nodes))
	for i, node := range nodes {
		nodeHosts[i] = node.Host
	}
	return nodeHosts
}
