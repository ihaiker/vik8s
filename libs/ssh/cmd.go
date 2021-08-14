package ssh

import (
	"bytes"
	"github.com/ihaiker/vik8s/libs/utils"
	"io"
	"os"
)

func (node *Node) Cmd(command string) error {
	return node.CmdWatcher(command, func(stdout io.Reader) {})
}

func (node *Node) CmdWatcher(command string, watcher StreamWatcher) error {
	defer node.reset()
	if node.isSudo() {
		command = "sudo " + command
	}
	if node.isShowLogger() {
		node.Logger("run command: %s", command)
	}
	return node.easyssh().Stream(command, watcher)
}
func (node *Node) CmdOutput(command string, output io.Writer) error {
	return node.CmdWatcher(command, func(stdout io.Reader) {
		_, _ = io.Copy(output, stdout)
	})
}

func (node *Node) CmdStdout(command string) error {
	return node.CmdOutput(command, os.Stdout)
}

func (node *Node) CmdPrefixStdout(command string) error {
	return node.CmdWatcher(command, utils.Stdout(node.Prefix()))
}

func (node *Node) CmdBytes(command string) (*bytes.Buffer, error) {
	output := bytes.NewBufferString("")
	err := node.CmdOutput(command, output)
	return output, err
}

func (node *Node) CmdString(command string) (string, error) {
	if outBytes, err := node.CmdBytes(command); err == nil {
		out := utils.Trdn(outBytes.Bytes())
		return string(out), nil
	} else {
		return "", err
	}
}

const (
	Sudo = 0b10
	Log  = 0b01
	None = 0b00
)

func (node *Node) Sudo() *Node {
	node.flag = node.flag | Sudo
	return node
}
func (node *Node) HideLog() *Node {
	node.flag = node.flag | Log
	return node
}
func (node *Node) isSudo() bool {
	return node.flag&Sudo == Sudo
}

func (node *Node) isShowLogger() bool {
	return node.flag&Log != Log
}
func (node *Node) reset() *Node {
	node.flag = None
	return node
}
