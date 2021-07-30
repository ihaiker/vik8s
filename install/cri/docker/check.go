package docker

import (
	"encoding/json"
	dockercerts "github.com/ihaiker/vik8s/certs/docker"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"strings"
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

	if idx := utils.Search(cfg.Hosts, "fd://"); idx == -1 {
		cfg.Hosts = append(cfg.Hosts, "fd://")
	}
	for i, host := range cfg.Hosts { //设置本地IP
		if strings.Contains(host, "{IP}") {
			cfg.Hosts[i] = strings.Replace(host, "{IP}", node.Host, 1)
		}
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
		err := node.SudoScp(cfg.TLS.CaCertPath, "/etc/docker/certs.d/ca.pem")
		utils.Panic(err, "upload cert file error: %s", cfg.TLS.CaCertPath)

		if cfg.TLS.ServerKeyPath != "" {
			err = node.SudoScp(cfg.TLS.ServerKeyPath, "/etc/docker/certs.d/key.pem")
			utils.Panic(err, "upload cert file error: %s", cfg.TLS.ServerKeyPath)

			err = node.SudoScp(cfg.TLS.ServerCertPath, "/etc/docker/certs.d/cert.pem")
			utils.Panic(err, "upload cert file error: %s", cfg.TLS.ServerCertPath)

		} else {
			serverCertPath, serverKeyPath, err := dockercerts.GenerateServerCertificates(node, cfg.TLS)
			utils.Panic(err, "generate server certificates")

			err = node.SudoScp(serverKeyPath, "/etc/docker/certs.d/key.pem")
			utils.Panic(err, "upload cert file error: %s", serverKeyPath)

			err = node.SudoScp(serverCertPath, "/etc/docker/certs.d/cert.pem")
			utils.Panic(err, "upload cert file error: %s", serverCertPath)
		}

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

	daemonJsonPath := "/etc/docker/daemon.json"
	daemonChange, serviceChange := false, false
	if cfg.DaemonJson != "" {
		if daemonChange = !node.Equal(cfg.DaemonJson, daemonJsonPath); daemonChange {
			err = node.SudoScp(cfg.DaemonJson, daemonJsonPath)
			utils.Panic(err, "scp daemon.json")
		}
	} else {
		bs := daemonJson(node, cfg)
		if daemonChange = !node.Equal(bs, daemonJsonPath); daemonChange {
			err = node.SudoScpContent(bs, daemonJsonPath)
			utils.Panic(err, "scp daemon.json")
		}
	}

	serviceConfig := `
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service containerd.service
Wants=network-online.target
Requires=docker.socket containerd.service
[Service]
Type=notify
ExecStart=/usr/bin/dockerd --containerd=/run/containerd/containerd.sock
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always
StartLimitBurst=3
StartLimitInterval=60s
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
Delegate=yes
KillMode=process
OOMScoreAdjust=-500

[Install]
WantedBy=multi-user.target
`
	dockerServicePath := "/usr/lib/systemd/system/docker.service"
	if serviceChange = !node.Equal([]byte(serviceConfig), dockerServicePath); serviceChange {
		err = node.SudoScpContent([]byte(serviceConfig), dockerServicePath)
		utils.Panic(err, "scp systemctl append file ")

		err = node.SudoCmd("systemctl daemon-reload")
		utils.Panic(err, "reload daemon")
	}

	bases.EnableAndStartService("docker", daemonChange || serviceChange, node)

	//BUGFIX 如果 Node 上安装的 Docker 版本大于 1.12，那么 Docker 会把默认的 iptables FORWARD 策略改为 DROP。
	//转发丢弃, 这会引发 Pod 网络访问的问题
	utils.Panic(node.SudoCmd("iptables -P FORWARD ACCEPT"), "open iptables role")

}

func installUbuntu(cfg *config.DockerConfiguration, node *ssh.Node, china bool) {

}
