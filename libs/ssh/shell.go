package ssh

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

func (node *Node) Shell(shell string) (out []byte, err error) {
	file := fmt.Sprintf("/tmp/vik8s-%s-%d.sh", time.Now().Format("2006.01.02"), rand.Int63())
	if err = node.ScpContent([]byte(shell), file); err != nil {
		return
	}
	out, err = node.Cmd("chmod +x "+file+" && sh -c "+file, true)
	return
}

func (node *Node) ShellChannel(shell string, handler func(stream io.Reader), hide ...bool) error {
	if !(len(hide) > 0 && hide[0]) {
		node.Logger("run shell <<<<<<<<<<< \n%s\n", shell)
	}
	file := fmt.Sprintf("/tmp/vik8s-%s-%d.sh", time.Now().Format("2006.01.02"), rand.Int63())
	if err := node.ScpContent([]byte(shell), file); err != nil {
		return err
	}
	return node.CmdChannel("chmod +x "+file+" && sh -c "+file, handler, true)
}
