package config

type K8SConfiguration struct {
	Version       string `ngx:"version"`
	KubeadmConfig string `ngx:"kubeadm-config"`

	ApiServer              string   `json:"api-server"`
	ApiServerCertExtraSans []string `json:"api-server-cert-extra-sans"`
	ApiServerVIP           string   `json:"api-server-vip"`

	Repo string `json:"repo,omitempty"`

	Interface string `json:"interface"`
	PodCIDR   string `json:"pod-cidr"`
	SvcCIDR   string `json:"svc-cidr"`
}
