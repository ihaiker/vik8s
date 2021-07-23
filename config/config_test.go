package config_test

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestConfigSuite struct {
	suite.Suite
}

func (p *TestConfigSuite) TestDocker() {
	err := config.Load("./_testdata/docker.conf")
	p.Nil(err)
	p.Equal("v20.1.10", config.Config.Docker.Version)
	p.Equal("/data", config.Config.Docker.DataRoot)
	p.Equal("{}", config.Config.Docker.DaemonJson)
	p.True(config.Config.Docker.StraitVersion)
}

func (p *TestConfigSuite) TestContainerd() {
	err := config.Load("./_testdata/containerd.conf")
	p.Nil(err)
	p.Equal("v1.21.2", config.Config.Containerd.Version)
	p.Equal("/data", config.Config.Containerd.DataRoot)
	p.True(config.Config.Containerd.StraitVersion)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}
