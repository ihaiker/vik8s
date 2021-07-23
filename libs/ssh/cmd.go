package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"io"
	"strings"
)

func (node *Node) MustCmd(cmd string, hide ...bool) {
	_ = node.MustCmd2String(cmd, hide...)
}

func (node *Node) MustCmd2String(cmd string, hide ...bool) string {
	out, err := node.cmd(cmd, len(hide) == 0)
	utils.Panic(err, "exec cmd %s", cmd)
	return string(out)
}

func (node *Node) Cmd2String(cmd string, hide ...bool) (string, error) {
	out, err := node.Cmd(cmd, hide...)
	return string(out), err
}

func (node *Node) Cmd(cmd string, hide ...bool) ([]byte, error) {
	return node.cmd(cmd, len(hide) == 0 || !hide[0])
}

//cmd Running command cmd and show logger
func (node *Node) cmd(cmd string, show bool) ([]byte, error) {
	if show {
		node.Logger("run command: %s", cmd)
	}
	return node.easyssh().Run(cmd)
}

func (node *Node) CmdChannel(cmd string, handler StreamWatcher, hide ...bool) error {
	if len(hide) == 0 || !hide[0] {
		node.Logger("run command: %s", cmd)
	}
	return node.easyssh().Stream(cmd, handler)
}

func (node *Node) CmdStd(cmd string, std io.Writer, hide ...bool) error {
	if len(hide) == 0 || !hide[0] {
		node.Logger("run command: %s", cmd)
	}
	return node.easyssh().Stream(cmd, func(stdout io.Reader) {
		_, _ = io.Copy(std, stdout)
	})
}

func (node *Node) MustCmdStd(cmd string, std io.Writer, hideCmd ...bool) {
	err := node.CmdStd(cmd, std, hideCmd...)
	utils.Panic(err, cmd)
}

func (node *Node) Mkdir(path ...string) error {
	_, err := node.cmd(fmt.Sprintf("mkdir -p %s", strings.Join(path, " ")), false)
	return err
}

func (node *Node) Md5Sum(file string) string {
	out, err := node.cmd(fmt.Sprintf("md5sum %s | awk '{printf $1}'", file), false)
	if err != nil {
		return ""
	}
	return strings.ToUpper(string(out))
}
