package tools

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"strings"
)

func SearchLabelNode(master *ssh.Node, labels map[string]string) []string {
	utils.Assert(len(labels) > 0, "label is empty")
	str := utils.Join(labels, ",", "=")

	out, err := master.CmdString(fmt.Sprintf(`kubectl get nodes -l '%s' -o=jsonpath="{.items[*].status.addresses[1].address}"`, str))
	utils.Panic(err, "get nodes error")

	hasname := strings.TrimSpace(out)
	if hasname == "" {
		return []string{}
	}
	return strings.Split(hasname, " ")
}

func AddNodeLabel(master *ssh.Node, labels map[string]string, nodes ...string) {
	for _, node := range nodes {
		for label, value := range labels {
			err := master.CmdStdout(fmt.Sprintf("kubectl label nodes %s %s=%s", node, label, value))
			utils.Panic(err, "add node %s label %s=%s", node, label, value)
		}
	}
}

func RemoveNodeLabel(master *ssh.Node, label string, nodes ...string) {
	for _, node := range nodes {
		err := master.CmdStdout(fmt.Sprintf("kubectl label nodes %s %s-", node, label))
		utils.Panic(err, "remove node %s label %s", node, label)
	}
}

//根据提供的IP或者hostname选择节点的hostname
func SelectNodeNames(nodes []*ssh.Node, selectNodes []string) []string {
	selectNodes = utils.ParseIPS(selectNodes)
	selectNodeNames := make([]string, 0)
NEXT:
	for _, selectNode := range selectNodes {
		for _, node := range nodes {
			if node.Hostname == selectNode || node.Host == selectNode {
				if utils.Search(selectNodeNames, node.Hostname) == -1 {
					selectNodeNames = append(selectNodeNames, node.Hostname)
				}
				continue NEXT
			}
		}
		utils.Panic(os.ErrNotExist, "node %s", selectNode)
	}
	return selectNodeNames
}

func AutoLabelNodes(master *ssh.Node, labels map[string]string, nodeNames ...string) {
	labeledNodes := SearchLabelNode(master, labels)

	//check label
	for _, labeledNode := range labeledNodes {
		utils.Assert(utils.Search(nodeNames, labeledNode) >= 0,
			color.RedString("node %s include label %s "), labeledNode, utils.Join(labels, ",", "="))
	}

	//add label
	for _, selectNode := range nodeNames {
		if utils.Search(labeledNodes, selectNode) == -1 {
			AddNodeLabel(master, labels, selectNode)
		}
	}
}
