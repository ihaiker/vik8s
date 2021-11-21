package hosts

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//解析用户输入地址 root:[paswword|privatekey]@ip-ip:port
var (
	userAndPwd = `([^:]+)(:(\S+))?`
	ip         = `(\d{1,3}\.\d{1,3}\.\d{1,3}\.)?\d{1,3}`
	port       = `\d{2,5}`
	pattern    = regexp.MustCompile(fmt.Sprintf(`(%s@)?((%s)-)?(%s)(:(%s))?`, userAndPwd, ip, ip, port))
)

//ParseAddrs 解析地址
func ParseAddrs(opt Option, args ...string) (ssh.Nodes, error) {
	nodes := ssh.Nodes{}
	for _, arg := range args {
		if ns, err := ParseAddr(opt, arg); err != nil {
			return nil, err
		} else {
			nodes = append(nodes, ns...)
		}
	}
	return nodes, nil
}

//ParseAddr 解析地址
func ParseAddr(opt Option, arg string) (nodes ssh.Nodes, err error) {
	defer utils.Catch(func(e error) {
		err = e
	})
	groups := pattern.FindStringSubmatch(arg)

	user := groups[2]
	if user == "" {
		user = opt.User
	}

	//1、使用用户提供的密码\秘钥
	password := groups[4]
	privateKey := ""
	//2、如果用户秘钥未提供，使用全局密码，
	if password == "" {
		password = opt.Password
	}
	//3、如果全局密码未设置、使用全局秘钥
	if password == "" {
		password = opt.PrivateKey
	}
	//判断是不是秘钥文件，如果是秘钥文件就替换
	if utils.Exists(os.ExpandEnv(password)) { //秘钥文件
		privateKey, _ = filepath.Abs(os.ExpandEnv(password))
		password = ""
	}
	endIp := groups[8]
	startIp := groups[6]
	if startIp == "" { //单独一个IP
		startIp = endIp
	}
	//处理 分段式标识形式
	if endIp, err = merge_end_ip(startIp, endIp); err != nil {
		return nil, utils.Error("invalid address: %s", arg)
	}

	from := net.ParseIP(startIp).To4()
	to := net.ParseIP(endIp).To4()

	port := groups[11]
	if port == "" {
		port = opt.Port
	}
	proxy := opt.Proxy

	nodes = ssh.Nodes{}
	for utils.ComposeIP(from, to) <= 0 {
		node := &ssh.Node{}
		node.User = user
		node.Password = password
		node.PrivateKey = privateKey
		node.Host = from.String()
		node.Port = port
		node.Passphrase = opt.Passphrase
		node.Proxy = proxy
		nodes = append(nodes, node)
		from = utils.NextIP(from)
	}
	return
}

func merge_end_ip(start, end string) (string, error) {
	segments := strings.Split(end, ".")
	num := len(segments)
	if num > 4 {
		return "", utils.Error("invalid address: %s", end)
	}
	startSegments := strings.Split(start, ".")
	segments = append(startSegments[0:4-num], segments...)
	return strings.Join(segments, "."), nil
}
