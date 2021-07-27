package docker

import (
	"encoding/json"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

//https://docs.docker.com/config/daemon/

//Install docker server to node, cfg is configuration.
func Install(cfg *config.DockerConfiguration, node *ssh.Node, china bool) {
	if node.IsCentOS() {
		installCentOS(cfg, node, china)
	} else {
		installUbuntu(cfg, node, china)
	}
}

func daemonJson(node *ssh.Node, cfg *config.DockerConfiguration) []byte {
	daemon := map[string]interface{}{
		"exec-opts": []string{"native.cgroupdriver=systemd"},
		"data-root": cfg.DataRoot,
	}
	if cfg.Hosts != nil {
		daemon["hosts"] = cfg.Hosts
	}
	if cfg.RegistryMirrors != nil {
		daemon["registry-mirrors"] = cfg.RegistryMirrors
	}
	if cfg.InsecureRegistries != nil {
		daemon["insecure-registries"] = cfg.InsecureRegistries
	}

	if cfg.Storage != nil {
		daemon["storage-driver"] = cfg.Storage.Driver
		daemon["storage-opts"] = cfg.Storage.Opt
	}
	if cfg.DNS != nil {
		daemon["dns"] = cfg.DNS.List
		daemon["dns-opts"] = cfg.DNS.Opt
		daemon["dns-search"] = cfg.DNS.Search
	}
	if cfg.TLS != nil {
		err := node.SudoScp(cfg.TLS.CaCert, "/etc/docker/certs.d/ca.pem")
		utils.Panic(err, "upload cert file error: %s", cfg.TLS.CaCert)

		err = node.SudoScp(cfg.TLS.ServerKey, "/etc/docker/certs.d/key.pem")
		utils.Panic(err, "upload cert file error: %s", cfg.TLS.ServerKey)

		err = node.SudoScp(cfg.TLS.ServerCert, "/etc/docker/certs.d/cert.pem")
		utils.Panic(err, "upload cert file error: %s", cfg.TLS.ServerCert)

		daemon["tls"] = true
		daemon["tlscacert"] = "/etc/docker/certs.d/ca.pem"
		daemon["tlscert"] = "/etc/docker/certs.d/cert.pem"
		daemon["tlskey"] = "/etc/docker/certs.d/key.pem"
		daemon["tlsverify"] = true
	}
	content, _ := json.MarshalIndent(daemon, "", "    ")
	return content
}

func installCentOS(cfg *config.DockerConfiguration, node *ssh.Node, china bool) {
	defer func() {
		if e := recover(); e == nil {
			bases.EnableAndStartService("docker", node)
			//BUGFIX 如果 Node 上安装的 Docker 版本大于 1.12，那么 Docker 会把默认的 iptables FORWARD 策略改为 DROP。
			//转发丢弃, 这会引发 Pod 网络访问的问题
			utils.Panic(node.SudoCmd("iptables -P FORWARD ACCEPT"),
				"open iptables role")
		} else {
			panic(e) //继续向下抛
		}
	}()

	dockerVersion, err := bases.GetPackageVersion(node, "docker-ce")
	utils.Panic(err, "get docker version")
	if dockerVersion != "" && (dockerVersion == cfg.Version[1:] || !cfg.StraitVersion) {
		node.Logger("docker has installed version %s", dockerVersion)
	} else {
		//BUGFIX: 当 centos 小于 7.3.1611 systemd 必须更新
		if node.Facts.MajorVersion == "7" {
			err = node.SudoCmdWatcher("yum update -y systemd", utils.Stdout(node.Prefix()))
			utils.Panic(err, "update systemd")
		}
		node.Logger("install docker-ce %s", cfg.Version)
		bases.AddRepoFile(node, "docker", []byte(repo.Docker()))

		bases.Install("docker-ce", cfg.Version[1:], node)
		bases.Install("docker-ce-cli", cfg.Version[1:], node)
	}

	err = node.SudoCmd("mkdir -p /etc/docker")
	utils.Panic(err, "make docker configuration folder")

	if cfg.DaemonJson != "" {
		err = node.SudoScp(cfg.DaemonJson, "/etc/docker/daemon.json")
		utils.Panic(err, "scp daemon.json")
	} else {
		bs := daemonJson(node, cfg)
		err := node.SudoScpContent(bs, "/etc/docker/daemon.json")
		utils.Panic(err, "scp daemon.json")
	}

	serviceConfig := `[Service]
ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2375 --containerd=/run/containerd/containerd.sock`
	err = node.SudoScpContent([]byte(serviceConfig), "/etc/systemd/system/docker.service.d/hosts.conf")
	utils.Panic(err, "scp systemctl append file ")
}

func installUbuntu(cfg *config.DockerConfiguration, node *ssh.Node, china bool) {

}
