package hosts_test

import (
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type TestHostsSuite struct {
	suite.Suite
	cfg *hosts.Option
}

func (t *TestHostsSuite) SetupTest() {
	paths.ConfigDir = "./_testdata"
	t.Nil(os.MkdirAll(filepath.Dir(paths.HostsConfiguration()), os.ModePerm))
	idRsaPath := paths.Join("id_rsa")
	t.generatorIdRsa(idRsaPath)
	t.cfg = &hosts.Option{
		User: "root", PrivateKey: idRsaPath, Port: 22,
	}
}

func (t TestHostsSuite) TearDownTest() {
	_ = os.RemoveAll(paths.ConfigDir)
}

func (t TestHostsSuite) generatorIdRsa(path string) {
	privateKey := certs.NewPrivateKey()
	content := certs.EncodePrivateKeyPEM(privateKey)
	err := ioutil.WriteFile(path, content, 0655)
	t.Nil(err, "write id_rsa error ", path)
}

func (t TestHostsSuite) assert(node *ssh.Node, user, password, pk, host, port string) {
	t.Equal(user, node.User)
	t.Equal(password, node.Password)
	t.Equal(pk, node.PrivateKey)
	t.Equal(host, node.Host)
	t.Equal(port, node.Port)
}

func (t TestHostsSuite) TestAdd() {
	t.T().Log(paths.HostsConfiguration())
	manager, err := hosts.New(paths.HostsConfiguration(), t.cfg, false)
	nodes := manager.All()
	t.Nil(err)
	t.Len(nodes, 0)

	nodes, err = manager.Add("172.16.100.1")
	t.Len(nodes, 1)

	t.assert(nodes[0], t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.1", "22")
	node := nodes.Get("172.16.100.1")
	t.assert(node, t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.1", "22")

	t.Len(manager.All(), 1)

	nodes, err = manager.Add("172.16.100.2:2222")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.2", "2222")

	t.Len(manager.All(), 2)

	nodes, err = manager.Add("haiker@172.16.100.3:1234")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], "haiker", t.cfg.Password, t.cfg.PrivateKey, "172.16.100.3", "1234")

	t.Len(manager.All(), 3)

	nodes, err = manager.Add("haiker@172.16.100.4")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], "haiker", t.cfg.Password, t.cfg.PrivateKey, "172.16.100.4", "22")

	nodes, err = manager.Add("haiker:1234qwer@172.16.100.5")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], "haiker", "1234qwer", "", "172.16.100.5", "22")

	nodes, err = manager.Add("giir-12312:bbb@172.16.100.10-172.16.100.15:234")
	t.Nil(err)
	t.Len(nodes, 6)
	for i := 0; i < 6; i++ {
		t.assert(nodes[i], "giir-12312", "bbb", "", "172.16.100."+strconv.Itoa(i+10), "234")
	}

	nodes, err = manager.Add("root:$HOME/.ssh/id_rsa_not_found@172.16.100.16-172.16.100.20:22")
	t.Nil(err)
	t.Len(nodes, 5)
	for i := 0; i < 5; i++ {
		t.assert(nodes[i], "root", "$HOME/.ssh/id_rsa_not_found", "", "172.16.100."+strconv.Itoa(i+16), "22")
	}

	nodes, err = manager.Add("172.16.100.21-172.16.100.22:221")
	t.Nil(err)
	t.Len(nodes, 2)
	for i := 0; i < 2; i++ {
		t.assert(nodes[i], "root", t.cfg.Password, t.cfg.PrivateKey, "172.16.100."+strconv.Itoa(i+21), "221")
	}
}

func (t TestHostsSuite) TestOverwrite() {
	manager, err := hosts.New(paths.HostsConfiguration(), t.cfg, false)
	t.Nil(err)

	nodes := manager.All()
	t.Len(nodes, 0)

	nodes, err = manager.Add("172.16.100.1")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.1", "22")

	node := nodes.Get("172.16.100.1")
	t.assert(node, t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.1", "22")

	//overwrite
	nodes, err = manager.Add("haiker:123456@172.16.100.1")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], "haiker", "123456", "", "172.16.100.1", "22")

	node = nodes.Get("172.16.100.1")
	t.assert(node, "haiker", "123456", "", "172.16.100.1", "22")
}

func (t TestHostsSuite) TestProxy() {
	cfg := paths.HostsConfiguration()
	manager, err := hosts.New(cfg, t.cfg, false)
	t.Nil(err)

	nodes := manager.All()
	t.Len(nodes, 0)

	nodes, err = manager.Add("172.16.100.1")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.1", "22")

	t.cfg.Proxy = "172.16.100.1"
	nodes, err = manager.Add("172.16.100.2")
	t.Nil(err)
	t.Len(nodes, 1)
	t.assert(nodes[0], t.cfg.User, t.cfg.Password, t.cfg.PrivateKey, "172.16.100.2", "22")
	t.Equal(t.cfg.Proxy, nodes[0].Proxy)
}

func TestHosts(t *testing.T) {
	suite.Run(t, new(TestHostsSuite))
}
