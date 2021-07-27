package hosts

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

var _manager Manager

func Load(path string, config *Option, gatherFacts bool) {
	manager, err := New(path, config, gatherFacts)
	utils.Panic(err, "load host configuration %s", path)
	_manager = manager
}

func Nodes() ssh.Nodes {
	return _manager.All()
}

func Add(args ...string) ssh.Nodes {
	ns, err := _manager.GetAdd(args...)
	utils.Panic(err, "add or get nodes")
	return ns
}

func Get(arg string) *ssh.Node {
	node := _manager.Get(arg)
	utils.Assert(node != nil, "not found %s", arg)
	return node
}

func Gets(args []string) (ns ssh.Nodes) {
	for _, arg := range args {
		ns = append(ns, Get(arg))
	}
	return
}

func MustGatheringFacts(ns ...*ssh.Node) {
	for _, node := range ns {
		utils.Panic(node.GatheringFacts(), "gathering facts %s", node.Host)
	}
}
