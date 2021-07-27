package ssh

import (
	"testing"
)

func TestNode_GatheringFacts(t *testing.T) {
	conf := getConfig("SSH_FACTS_")
	if conf == nil {
		t.Log("skip gathering facts")
		return
	}
	node := &Node{
		Host:       conf.Server,
		Port:       conf.Port,
		User:       conf.User,
		Password:   conf.Password,
		PrivateKey: conf.KeyPath,
		Passphrase: conf.Passphrase,
		Facts:      Facts{},
	}
	if err := node.GatheringFacts(); err != nil {
		t.Fatal(err)
	}
	t.Log(node.Facts)
}
