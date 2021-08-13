package config

import "time"

type K8SConfiguration struct {
	Masters       []string `ngx:"masters" flag:"-"`
	Nodes         []string `ngx:"nodes" flag:"-"`
	Version       string   `ngx:"version" help:"Specify k8s version"`
	KubeadmConfig string   `ngx:"kubeadm-config" help:"Path to a kubeadm configuration file. see kubeadm --config"`

	ApiServer              string   `ngx:"api-server" def:"api-vik8s-io" help:"Specify a stable IP address or DNS name for the control plane. see kubeadm --control-plane-endpoint"`
	ApiServerVIP           string   `flag:"-" ngx:"api-server-vip"`
	ApiServerCertExtraSans []string `ngx:"api-server-cert-extra-sans" help:"see kubeadm init --apiserver-cert-extra-sans"`

	Repo string `ngx:"repo" help:"Choose a container registry to pull control plane images from. \n(default: Best choice from k8s.gcr.io and registry.aliyuncs.com/google_containers.)"`

	Interface string `ngx:"network-interface" def:"eth.*|en.*|em.*" help:"name of network interface"`
	PodCIDR   string `ngx:"pod-cidr" flag:"pod-cidr" def:"100.64.0.0/24" help:"Specify range of IP addresses for the pod network"`
	SvcCIDR   string `ngx:"svc-cidr" flag:"svc-cidr" def:"10.96.0.0/12" help:"Use alternative range of IP address for service VIPs"`

	CertsValidity time.Duration `ngx:"certs-validity" def:"876000h" help:"Certificate validity time"`
	Timezone      string        `ngx:"timezone" def:"Asia/Shanghai"`
	NTPServices   []string      `ngx:"ntp-services" flag:"ntp-services" def:"ntp1.aliyun.com,ntp2.aliyun.com,ntp3.aliyun.com" help:"time server\n"`
}

func DefaultK8SConfiguration() *K8SConfiguration {
	return &K8SConfiguration{
		Version:       "v1.21.3",
		ApiServer:     "api-vik8s-io",
		Interface:     "eth.*|en.*|em.*",
		PodCIDR:       "100.64.0.0/24",
		SvcCIDR:       "10.96.0.0/12",
		CertsValidity: 100 * 365 * 24 * time.Hour,
		Timezone:      "Asia/Shanghai",
		NTPServices:   []string{"ntp1.aliyun.com", "ntp2.aliyun.com", "ntp3.aliyun.com"},
	}
}

func (cfg *K8SConfiguration) ExistsNode(ip string) (exists bool, master bool) {
	for _, node := range cfg.Masters {
		if node == ip {
			exists = true
			master = true
			return
		}
	}
	for _, node := range cfg.Nodes {
		if node == ip {
			exists = true
			master = false
			return
		}
	}
	return
}

func (cfg *K8SConfiguration) JoinNode(master bool, ip string) {
	if master {
		cfg.Masters = append(cfg.Masters, ip)
	} else {
		cfg.Nodes = append(cfg.Nodes, ip)
	}
}

func (cfg *K8SConfiguration) RemoveNode(ip string) {
	for i, node := range cfg.Nodes {
		if node == ip {
			cfg.Nodes = append(cfg.Nodes[0:i], cfg.Nodes[i+1:]...)
			break
		}
	}
	for i, node := range cfg.Masters {
		if node == ip {
			cfg.Masters = append(cfg.Masters[0:i], cfg.Masters[i+1:]...)
			break
		}
	}
}
