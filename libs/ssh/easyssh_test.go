package ssh

import (
	"github.com/ihaiker/ngx/v2"
	"github.com/ihaiker/ngx/v2/query"
	"github.com/stretchr/testify/suite"
	"testing"
)

type easysshSuite struct {
	suite.Suite
	config *ngx.Configuration
}

func (t *easysshSuite) SetupTest() {
	config, err := ngx.Parse("../../bin/default/hosts.conf")
	t.Nil(err, "load config error")
	t.config = config
}

func TestEasySsh(t *testing.T) {
	suite.Run(t, new(easysshSuite))
}

func (t *easysshSuite) getConfig(prefix string) *sshConfig {
	nodes, err := query.Selects(t.config, ".node")
	t.Nil(err, "select config node: .node")
	for _, node := range nodes {
		hostname := node.Body.Get("hostname").Args[0]
		if hostname != prefix {
			continue
		}
		host := node.Body.Get("host").Args[0]
		port := node.Body.Get("port").Args[0]
		user := node.Body.Get("user").Args[0]
		privateKey := node.Body.Get("private-key").Args[0]
		return &sshConfig{
			User: user, Server: host, Port: port,
			KeyPath: privateKey,
		}
	}
	t.Failf("not found prefix node: %s", prefix)
	return nil
}

func (t *easysshSuite) TestSSH() {
	conf := t.getConfig("master01")
	con := &easySSHConfig{sshConfig: *conf}
	stdout, err := con.Run("cat /etc/os-release")
	t.Nil(err, "run command error")
	t.NotEmpty(stdout, "run command stdout is empty")
	t.T().Log(string(stdout), err)
}
