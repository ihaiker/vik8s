package tools

import (
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func GatheringFacts(nodes ssh.Nodes, manager *hosts.Manager) error {
	for _, node := range nodes {
		if node.Proxy != "" {
			if node.ProxyNode = manager.Get(node.Proxy); node.ProxyNode == nil {
				return utils.Error("not found bastion node: %s", node.Proxy)
			}
		}
		if err := node.GatheringFacts(); err != nil {
			return err
		}
		if err := manager.Add(node); err != nil {
			return err
		}
	}
	return nil
}
