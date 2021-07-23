package config

type K8SConfiguration struct {
	Version       string `json:"k8s-version"`
	KubeadmConfig string `json:"kubeadm-cfg,omitempty"`

	ApiServer              string   `json:"apiserver"`
	ApiServerCertExtraSans []string `json:"apiServerCertExtraSans"`
	ApiServerVIP           string   `json:"apiserver-vip"`

	Repo string `json:"repo,omitempty"`

	Interface string `json:"interface"`
	PodCIDR   string `json:"pod-cidr"`
	SvcCIDR   string `json:"svc-cidr"`
}
