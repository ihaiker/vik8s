package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"golang.org/x/crypto/ssh"
	"gopkg.in/fatih/color.v1"
	"io"
	"strings"
)

func (node *Node) session(r func(session *ssh.Session) error) error {
	return node.connect(func(client *ssh.Client) error {
		if session, err := client.NewSession(); err != nil {
			return utils.Wrap(err, "open session")
		} else {
			defer session.Close()
			return r(session)
		}
	})
}

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

func (node *Node) cmd(cmd string, show bool) (out []byte, err error) {
	err = node.session(func(session *ssh.Session) error {
		if show {
			node.Logger("cmd [%s]", color.BlueString(cmd))
		}
		out, err = session.CombinedOutput(cmd)
		if err == nil {
			length := len(out)
			if length > 0 && out[length-1] == '\n' {
				out = out[0 : length-1]
			}
		}
		return err
	})
	return
}

func (node *Node) CmdChannel(cmd string, handler func(stream io.Reader), hide ...bool) error {
	return node.session(func(session *ssh.Session) error {
		if !(len(hide) > 0 && hide[0]) {
			node.Logger("cmd [%s]", color.BlueString(cmd))
		}
		reader, writer := io.Pipe()
		defer reader.Close()
		defer writer.Close()
		session.Stderr = writer
		session.Stdout = writer
		go handler(reader)
		if err := session.Start(cmd); err != nil {
			return err
		}
		return session.Wait()
	})
}

func (node *Node) CmdStd(cmd string, std io.Writer, hideCmd ...bool) error {
	return node.session(func(session *ssh.Session) error {
		if !(len(hideCmd) > 0 && hideCmd[0]) {
			node.Logger("cmd [%s]", color.BlueString(cmd))
		}
		session.Stderr = std
		session.Stdout = std
		if err := session.Start(cmd); err != nil {
			return err
		}
		return session.Wait()
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
