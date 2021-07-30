package docker

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type TestDockerSuite struct {
	suite.Suite
}

func (p TestDockerSuite) SetupTest() {
	paths.ConfigDir = "./_testdata"
	paths.Cloud = "test"
	err := os.MkdirAll(paths.ConfigDir, os.ModePerm)
	p.Nil(err)
}

func (p TestDockerSuite) TearDownTest() {
	_ = os.RemoveAll(filepath.Join(paths.ConfigDir))
}

func (t *TestDockerSuite) TestEnableTLS() {
	cfg := config.DefaultDockerConfiguration()
	cfg.TLS.Enable = true
	err := Config(cfg)
	t.Nil(err)
	t.False(cfg.TLS.Custom)
	t.FileExists(cfg.TLS.CaPrivateKeyPath, "ca key not gen")
	t.FileExists(cfg.TLS.CaCertPath, "ca key not gen")
	t.FileExists(cfg.TLS.ClientKeyPath, "client key not gen")
	t.FileExists(cfg.TLS.ClientCertPath, "client cert not gen")
}

func (t *TestDockerSuite) TestEnableCustomTLS() {
	cfg := config.DefaultDockerConfiguration()
	cfg.TLS.Enable = true
	cfg.TLS.CaCertPath = paths.Join(DockerCertsPath, "ca.pem")

	err := Config(cfg)
	t.Nil(err)
	t.True(cfg.TLS.Custom, "custom is not true")
}

func (t *TestDockerSuite) TestDaemonJson() {
	path := paths.Join("daemon-test.json")
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	t.Nil(err, "mkdir dir error ", err)

	err = ioutil.WriteFile(path, []byte(`{}`), 0655)
	t.Nil(err, "write test deamon-test.json")

	cfg := config.DefaultDockerConfiguration()
	cfg.DaemonJson = path
	err = Config(cfg)
	t.Nil(err, "config daemon.json error")
	t.False(cfg.TLS.Enable)
}
func TestDockerConfig(t *testing.T) {
	suite.Run(t, new(TestDockerSuite))
}
