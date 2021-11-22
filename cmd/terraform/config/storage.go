package config

import "github.com/ihaiker/vik8s/libs/ssh"

type MemStorage struct {
	Hosts map[string]*ssh.Node
}
