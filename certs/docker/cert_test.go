package dockercert

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func TestGenerateCertificate(t *testing.T) {
	tmpDir := "_testdata"
	err := os.MkdirAll(tmpDir, os.ModePerm)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	var cfg *config.DockerCertsConfiguration
	if cfg, err = GenerateBootstrapCertificates(tmpDir); err != nil {
		t.Fatal(err)
	}
	node := &ssh.Node{
		Host: "10.24.0.10",
	}
	if _, _, err := GenerateServerCertificates(node, cfg); err != nil {
		t.Fatal(err)
	}
}
