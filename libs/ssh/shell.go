package ssh

import (
	"fmt"
	"math/rand"
	"time"
)

func (node *Node) Shell(shell string, watch StreamWatcher) error {
	sudo := node.isSudo()
	file := fmt.Sprintf("/tmp/vik8s-%s-%d.sh", time.Now().Format("2006.01.02"), rand.Int63())
	if err := node.ScpContent([]byte(shell), file); err != nil {
		return err
	}
	if sudo {
		node = node.Sudo()
	}
	return node.CmdWatcher(fmt.Sprintf("sh -c 'chmod +x %s && sh -c %s'", file, file), watch)
}
