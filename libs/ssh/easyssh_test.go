package ssh

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getConfig(prefix string) *sshConfig {
	host := os.Getenv(prefix + "HOST")
	if host == "" {
		return nil
	}
	port := os.Getenv(prefix + "PORT")
	user := os.Getenv(prefix + "USER")
	password := os.Getenv(prefix + "PASSWORD")
	return &sshConfig{
		User:     user,
		Server:   host,
		Port:     port,
		Password: password,
	}
}

func TestSSH(t *testing.T) {
	conf := getConfig("SSH_TEST_PROXY_")
	if conf == nil {
		t.Log("skip ssh test. no config")
		return
	}
	con := &easySSHConfig{sshConfig: *conf}
	stdout, err := con.Run("hostname")
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	t.Log(stdout)
}

func TestProxySSH(t *testing.T) {
	conf := getConfig("SSH_TEST_")
	if conf == nil {
		t.Log("skip ssh test. no config")
		return
	}
	con := &easySSHConfig{sshConfig: *conf, Proxy: getConfig("SSH_TEST_PROXY_")}

	stdout, err := con.Run("hostname")
	assert.Nil(t, err)
	assert.NotEmpty(t, string(stdout))
	t.Log(stdout)
}
