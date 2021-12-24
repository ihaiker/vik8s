package ssh

import (
	"bytes"
	"github.com/fatih/color"
	"github.com/ihaiker/vik8s/libs/utils"
	"io"
	"os"
)

func (node *Node) Cmd(command string) error {
	return node.CmdWatcher(command, func(stdout io.Reader) error { return nil })
}

func (node *Node) CmdWatcher(command string, watcher StreamWatcher) (err error) {
	defer node.reset()
	if node.isSudo() {
		command = "sudo " + command
	}
	if node.isShowLogger() {
		node.Logger("run command: %s", command)
	}
	retries := node.retries
	if retries == 0 {
		retries = 1
	}
	for i := 0; i < retries; i++ {
		if err = node.stream(command, watcher); err == nil {
			return
		} else if retries != 1 {
			node.Logger(color.New(color.FgHiRed).Sprintf("executor: [%s], error: %s", command, err.Error()))
		}
	}
	return
}
func (node *Node) CmdOutput(command string, output io.Writer) error {
	return node.CmdWatcher(command, func(stdout io.Reader) error {
		_, err := io.Copy(output, stdout)
		return err
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

func (node *Node) Retries(retries int) *Node {
	node.retries = retries
	return node
}

func (node *Node) isSudo() bool {
	return node.flag&Sudo == Sudo && node.User != "root"
}

func (node *Node) isShowLogger() bool {
	return node.flag&Log != Log
}
func (node *Node) reset() *Node {
	node.flag = None
	node.retries = 1
	return node
}
