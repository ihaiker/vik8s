package hosts

import (
	"fmt"
	"github.com/ihaiker/ngx/v2"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

//解析用户输入地址 root:[paswword|privatekey]@ip-ip:port
var (
	userAndPwd = `([^:]+)(:(\S+))?`
	ip         = `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
	port       = `\d{2,5}`
	pattern    = regexp.MustCompile(fmt.Sprintf(`(%s@)?((%s)-)?(%s)(:(%s))?`, userAndPwd, ip, ip, port))
)

type Option struct {
	Port       int    `ngx:"port" help:"ssh port" def:"22"`
	User       string `ngx:"user" short:"u" help:"ssh user"`
	Password   string `ngx:"password" short:"P" help:"ssh password"`
	PrivateKey string `ngx:"private-key" short:"i" help:"ssh private key" def:"$HOME/.ssh/id_rsa"`
	Passphrase string `ngx:"passphrase" flag:"passphrase" help:"private key passphrase"`
	Proxy      string `ngx:"proxy" flag:"proxy" help:"ssh proxy server"`
}

type Manager interface {
	All() ssh.Nodes
	Add(args ...string) (ssh.Nodes, error)
	GetAdd(args ...string) (ssh.Nodes, error)
	Get(hostnameOrIp string) *ssh.Node
}

func New(path string, opts *Option, gatherFacts bool) (manager *defaultManager, err error) {
	manager = &defaultManager{
		path: path, opts: opts, gatherFacts: gatherFacts,
	}
	err = manager.load()
	return
}

type defaultManager struct {
	path        string
	opts        *Option
	Nodes       ssh.Nodes `ngx:"node"`
	gatherFacts bool
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
		if node.Proxy != "" {
			node.ProxyNode = this.Get(node.Proxy)
		}
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

func (this *defaultManager) addOrGet(onlyAdd bool, args ...string) (ns ssh.Nodes, err error) {

	if onlyAdd { //如果是添加，检查参数的有效性
		for _, arg := range args {
			if !pattern.MatchString(arg) {
				return nil, utils.Error("invalid arg %s", arg)
			}
		}
	}

	defer this.down()

	ns = ssh.Nodes{}
	for _, arg := range args {

		if !onlyAdd {
			node := this.Nodes.TryGet(arg)
			if node == nil && !pattern.MatchString(arg) { //不是给定的IP，且查找不到报错
				err = utils.Error("not found node %s", arg)
				return
			}
			if node != nil { //查找到了，直接进入下一个
				ns = append(ns, node)
				continue
			}
		}

		groups := pattern.FindStringSubmatch(arg)

		user := groups[2]
		if user == "" {
			user = this.opts.User
		}
		//1、使用用户提供的密码\秘钥
		password := groups[4]
		privateKey := ""
		//2、如果用户秘钥未提供，使用全局密码，
		if password == "" {
			password = this.opts.Password
		}
		//3、如果全局密码未设置、使用全局秘钥
		if password == "" {
			password = this.opts.PrivateKey
		}
		//判断是不是秘钥文件，如果是秘钥文件就替换
		if utils.Exists(os.ExpandEnv(password)) { //秘钥文件
			privateKey, _ = filepath.Abs(os.ExpandEnv(password))
			password = ""
		}
		endIp := groups[7]
		startIp := groups[6]
		if startIp == "" {
			startIp = endIp
		}

		from := net.ParseIP(startIp).To4()
		to := net.ParseIP(endIp).To4()

		port := groups[9]
		if port == "" {
			port = strconv.Itoa(this.opts.Port)
		}

		proxy := this.opts.Proxy

		for utils.ComposeIP(from, to) <= 0 {
			node := this.Nodes.TryGet(from.String())
			if !onlyAdd && node != nil {
				ns = append(ns, node)
				continue
			}
			isAdd := node == nil
			if node == nil {
				node = &ssh.Node{}
			}
			node.User = user
			node.Password = password
			node.PrivateKey = privateKey
			node.Host = from.String()
			node.Port = port
			node.Passphrase = this.opts.Passphrase
			node.Proxy = proxy

			if proxy != "" {
				if proxyNode := this.Get(proxy); proxyNode == nil {
					err = utils.Error("the ssh proxy %s not found", proxy)
					return
				} else {
					node.ProxyNode = proxyNode
				}
			}
			if this.gatherFacts {
				utils.Panic(node.GatheringFacts(), "connect host error")
			}

			if isAdd {
				this.Nodes = append(this.Nodes, node)
				fmt.Printf("add host %s: %s \n", node.Hostname, node.Host)
			} else {
				fmt.Printf("overwrite host %s: %s \n", node.Hostname, node.Host)
			}

			ns = append(ns, node)
			from = utils.NextIP(from)
		}
	}
	return
}

func (this *defaultManager) Add(args ...string) (ssh.Nodes, error) {
	return this.addOrGet(true, args...)
}

func (this *defaultManager) GetAdd(args ...string) (ssh.Nodes, error) {
	return this.addOrGet(false, args...)
}

func (this *defaultManager) Get(args string) *ssh.Node {
	return this.Nodes.TryGet(args)
}
