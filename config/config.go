package config

import (
	"github.com/ihaiker/ngx/v2"
	"github.com/ihaiker/vik8s/libs/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Configuration struct {
	filename   string
	Docker     *DockerConfiguration     `ngx:"docker"`
	Containerd *ContainerdConfiguration `ngx:"containerd"`
	K8S        *K8SConfiguration        `ngx:"k8s"`
	ETCD       *ETCD                    `ngx:"etcd"`
}

var (
	Config = &Configuration{}
)

func Docker() *DockerConfiguration {
	return Config.Docker
}
func Containerd() *ContainerdConfiguration {
	return Config.Containerd
}
func K8S() *K8SConfiguration {
	return Config.K8S
}
func Etcd() *ETCD {
	return Config.ETCD
}

//加载vik8s.conf配置，如果配置文件不存在，直接返回空配置
func Load(filename string) (err error) {
	Config.filename = filename
	if !utils.Exists(filename) {
		return
	}
	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = ngx.Unmarshal(data, Config); err != nil {
		return
	}
	return
}

func (cfg *Configuration) Write() error {
	data, _ := ngx.Marshal(cfg)
	if err := os.MkdirAll(filepath.Dir(cfg.filename), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(cfg.filename, data, 0666)
}

func (cfg *Configuration) IsDockerCri() bool {
	return cfg.Docker != nil
}
