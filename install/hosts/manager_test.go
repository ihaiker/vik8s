package hosts_test

import (
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type TestHostsSuite struct {
	suite.Suite
	cfg *hosts.Option
}

func (t *TestHostsSuite) SetupTest() {
	paths.ConfigDir = "./_testdata"
	t.TearDownTest() //remove config folder
}

func (t TestHostsSuite) TearDownTest() {
	_ = os.RemoveAll(paths.ConfigDir)
}

func (t TestHostsSuite) TestGetOrAdd() {
	cfg := paths.HostsConfiguration()
	manager, err := hosts.New(cfg)
	t.Nil(err, "初始化 hosts.conf错误")

	node := &ssh.Node{
		Host:       "10.24.0.10",
		Port:       "22",
		User:       "root",
		Password:   "",
		PrivateKey: "",
		Passphrase: "",
		Hostname:   "master",
		Proxy:      "",
		ProxyNode:  nil,
		Facts:      ssh.Facts{},
	}
	_ = manager.Add(node)

	node = manager.Get("10.24.0.10")
	t.NotNil(node)

	node = manager.Get("master")
	t.NotNil(node)
}

func TestHosts(t *testing.T) {
	suite.Run(t, new(TestHostsSuite))
}
