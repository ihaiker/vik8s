package hosts

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"strings"
)

var _manager Manager
var _opt *Option

//Load 加载节点配置
func Load(path string, opt *Option) (err error) {
	_opt = opt
	_manager, err = New(path)
	return
}

//Nodes 所有节点
func Nodes() ssh.Nodes {
	return _manager.All()
}

// Fetch 配置
func Fetch(overwrite bool, args ...string) (ssh.Nodes, error) {
	nodes := ssh.Nodes{}

	for _, arg := range args {
		if node := _manager.Get(arg); node != nil {
			nodes = append(nodes, node)
		} else if ns, err := parse_addr(*_opt, arg); err == nil {
			for _, node := range ns {
				if err := getProxy(_manager, node); err != nil {
					return nil, err
				}
			}
			nodes = append(nodes, ns...)
		} else {
			return nil, utils.Error("not found node: %s", arg)
		}
	}

	for i, node := range nodes {
		oldNode := _manager.Get(node.Host)
		//当前节点不在列表中，或者需要覆盖
		if oldNode == nil || overwrite {
			if err := node.GatheringFacts(); err != nil {
				return nil, utils.Wrap(err, "gathering facts")
			}
			_ = _manager.Add(node)
		} else {
			nodes[i] = oldNode
		}
	}
	return nodes, nil
}

func MustGet(arg string) *ssh.Node {
	node := _manager.Get(arg)
	utils.Assert(node != nil, "not found %s", arg)
	err := node.GatheringFacts()
	utils.Panic(err, "gathering facts: %s", node.Host)
	return node
}

func MustGets(args []string) ssh.Nodes {
	nodes, err := Fetch(false, args...)
	utils.Panic(err, "fetch Nodes: %s", strings.Join(args, ", "))
	for _, node := range nodes {
		err = node.GatheringFacts()
		utils.Panic(err, "gathering facts: %s", node.Host)
	}
	return nodes
}
