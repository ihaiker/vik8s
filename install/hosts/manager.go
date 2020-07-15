package hosts

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type SSH struct {
	Password string `json:"password,omitempty"`
	PkFile   string `json:"pk"`
	Port     int    `json:"port"`
}

/*
	解析用户输入地址 root:paswword@ip-ip:port
*/
var pattern = regexp.MustCompile(`((\S+):(\S+)@)?((\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})-)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(:(\d{1,3}))?`)

func Add(cfg SSH, args ...string) (ns ssh.Nodes) {
	readHosts()
	defer writeHosts()

	ns = ssh.Nodes{}
	for _, arg := range args {

		if !pattern.MatchString(arg) {
			ns = append(ns, nodes.Get(arg))
			continue
		}

		groups := pattern.FindStringSubmatch(arg)
		//user := groups[2] 暂时不支持非root用户，但是预留功能

		pwdType := ssh.Password
		password := groups[3]
		if password == "" {
			password = cfg.Password
		}
		if password == "" {
			password = cfg.PkFile
		}
		pkFile := ""
		if utils.Exists(os.ExpandEnv(password)) { //秘钥文件
			pwdType = ssh.PrivateKey
			pkFile = password
			password = ""
		}
		endIp := groups[6]
		startIp := groups[5]
		if startIp == "" {
			startIp = endIp
		}

		port := groups[8]
		if port == "" {
			port = strconv.Itoa(cfg.Port)
		}

		from := net.ParseIP(startIp).To4()
		to := net.ParseIP(endIp).To4()

		for utils.ComposeIP(from, to) <= 0 {
			node := nodes.TryGet(from.String())
			if node == nil {
				node = &ssh.Node{
					User: "root", Password: password, Key: pkFile,
					Type: pwdType,
					Host: from.String(), Port: port,
				}
				utils.Panic(node.Info(), "connection node: %s", node.Host)
				fmt.Printf("add host %s: %s \n", node.Hostname, node.Host)
				ns = append(ns, node)
				nodes = append(nodes, node)
			} else {
				ns = append(ns, node)
			}
			from = utils.NextIP(from)
		}
	}
	return
}

func Get(nameOrIP string) *ssh.Node {
	readHosts()
	return nodes.Get(nameOrIP)
}

func Remove(nameOrIP string) {
	readHosts()
	defer writeHosts()
	for i, node := range nodes {
		if node.Hostname == nameOrIP || node.Host == nameOrIP {
			nodes = append(nodes[0:i], nodes[i+1:]...)
			break
		}
	}
}

func Gets(args ...[]string) ssh.Nodes {
	nodes := ssh.Nodes{}
	for _, arg := range args {
		for _, s := range arg {
			nodes = append(nodes, Get(s))
		}
	}
	return nodes
}

var hasRead = false

func readHosts() {
	if hasRead {
		return
	}
	hasRead = true
	name := tools.Join("hosts.json")
	if utils.Exists(name) {
		bodys, err := ioutil.ReadFile(name)
		utils.Panic(err, "read file")
		err = json.Unmarshal(bodys, &nodes)
		utils.Panic(err, "read file")
	}
}

func writeHosts() {
	name := tools.Join("hosts.json")
	dir := filepath.Dir(name)
	utils.Panic(os.MkdirAll(dir, os.ModePerm), "mkdir %s", dir)
	out, err := json.MarshalIndent(nodes, "", "    ")
	utils.Panic(err, "write hosts")
	utils.Panic(ioutil.WriteFile(name, out, 0644), "write hosts")
}

func Nodes() ssh.Nodes {
	readHosts()
	return nodes
}

var nodes = ssh.Nodes{}
