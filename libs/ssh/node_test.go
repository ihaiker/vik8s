package ssh

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type nodeSuite struct {
	*easysshSuite
}

func (t *nodeSuite) TestNodeFlag() {
	fnode := new(Node)
	t.False(fnode.isSudo())
	t.True(fnode.isShowLogger())

	fnode.Sudo()
	t.True(fnode.isSudo())
	t.True(fnode.isShowLogger())

	fnode.reset()
	t.False(fnode.isSudo())
	t.True(fnode.isShowLogger())

	fnode.HideLog()
	t.False(fnode.isSudo())
	t.False(fnode.isShowLogger())

	fnode.Sudo().HideLog()
	t.True(fnode.isSudo())
	t.False(fnode.isShowLogger())
}

func TestNode(t *testing.T) {
	suite.Run(t, &nodeSuite{new(easysshSuite)})
}
