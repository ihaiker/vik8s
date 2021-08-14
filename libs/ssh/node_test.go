package ssh

import (
	"github.com/stretchr/testify/assert"
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

func TestFlag(t *testing.T) {
	fnode := new(Node)
	assert.False(t, fnode.isSudo())
	assert.True(t, fnode.isShowLogger())

	fnode.Sudo()
	assert.True(t, fnode.isSudo())
	assert.True(t, fnode.isShowLogger())

	fnode.reset()
	assert.False(t, fnode.isSudo())
	assert.True(t, fnode.isShowLogger())

	fnode.HideLog()
	assert.False(t, fnode.isSudo())
	assert.False(t, fnode.isShowLogger())

	fnode.Sudo().HideLog()
	assert.True(t, fnode.isSudo())
	assert.False(t, fnode.isShowLogger())
}
