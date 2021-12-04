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
	cfg, err := config.Load("./_testdata/docker.conf")
	p.Nil(err)
	p.Equal("v20.1.10", cfg.Docker.Version)
	p.Equal("/data", cfg.Docker.DataRoot)
	p.Equal("{}", cfg.Docker.DaemonJson)
	p.True(cfg.Docker.StraitVersion)
}

func (p *TestConfigSuite) TestContainerd() {
	cfg, err := config.Load("./_testdata/containerd.conf")
	p.Nil(err)
	p.Equal("v1.21.2", cfg.Containerd.Version)
	p.Equal("/data", cfg.Containerd.DataRoot)
	p.True(cfg.Containerd.StraitVersion)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}
