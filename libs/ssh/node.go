package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"golang.org/x/crypto/ssh"
	"gopkg.in/fatih/color.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	AuthType string
	Node     struct {
		Host     string   `json:"host" yaml:"host" toml:"host"`
		Port     string   `json:"port" yaml:"port" toml:"port"`
		User     string   `json:"user" yaml:"user" toml:"user"`
		Type     AuthType `json:"type" yaml:"type" toml:"type"`
		Password string   `json:"password" yaml:"password" toml:"password"`
		Key      string   `json:"key" yaml:"key" toml:"key"`

		Hostname      string `json:"hostname"`
		ReleaseName   string `json:"releaseName"`
		MajorVersion  string `json:"majorVersion"`
		KernelVersion string `json:"kernelVersion"`
	}
	Nodes []*Node
)

const (
	Password   AuthType = "password"
	PrivateKey AuthType = "key"
)

func publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
	keyPath := os.ExpandEnv(kPath)
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, utils.Wrap(err, fmt.Sprintf("ssh key file read failed. path: %s ", keyPath))
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, utils.Wrap(err, "ssh key signer failed")
	}
	return ssh.PublicKeys(signer), nil
}

func (node *Node) connect(r func(client *ssh.Client) error) error {
	config := &ssh.ClientConfig{
		Config: ssh.Config{
			Ciphers: []string{
				"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
				"arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc",
			},
		},
		Timeout:         time.Second * 3,
		User:            node.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}

	if node.Type == Password {
		config.Auth = []ssh.AuthMethod{ssh.Password(node.Password)}
	} else if auth, err := publicKeyAuthFunc(node.Key); err != nil {
		return err
	} else {
		config.Auth = []ssh.AuthMethod{auth}
	}

	if sshClient, err := ssh.Dial("tcp", node.Address(), config); err != nil {
		return utils.Wrap(err, "connect %s", node.Address())
	} else {
		defer sshClient.Close()
		return r(sshClient)
	}
}

func (node *Node) Info() error {
	if hostname, err := node.cmd("hostname -s", false); err != nil {
		return err
	} else {
		node.Hostname = string(hostname)
	}
	if distribution, err := node.cmd("uname -r", false); err != nil {
		return err
	} else {
		if strings.Index(string(distribution), "el7") != -1 {
			node.MajorVersion = "7"
		} else if strings.Index(string(distribution), "el8") != -1 {
			node.MajorVersion = "8"
		}
		node.KernelVersion = strings.Split(string(distribution), "-")[0]
	}
	if releaseName, err := node.cmd("cat /etc/redhat-release  | awk '{printf $1}'", false); err != nil {
		return err
	} else {
		node.ReleaseName = string(releaseName)
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
