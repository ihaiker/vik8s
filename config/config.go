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

//Load 加载vik8s.conf配置，如果配置文件不存在，直接返回空配置
func Load(filename string) (cfg *Configuration, err error) {
	cfg = new(Configuration)
	cfg.filename = filename
	if !utils.Exists(filename) {
		cfg.K8S = DefaultK8SConfiguration()
		cfg.Docker = DefaultDockerConfiguration()
		return
	}

	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = ngx.Unmarshal(data, cfg); err != nil {
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

func (cfg *Configuration) IsExternalETCD() bool {
	return cfg.ETCD != nil && len(cfg.ETCD.Nodes) > 0
}
