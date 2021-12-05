package cluster

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/terraform/tools"
	"time"
)

func k8SConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"version": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"api_server": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"kubeadm_config": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"api_server_cert_extra_sans": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					ForceNew: true,
				},
				"repo": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Computed: true,
				},
				"network_interface": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"pod_cidr": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"svc_cidr": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"certs_validity": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"timezone": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"ntp_services": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func expendK8SConfiguration(p interface{}) (cfg *config.K8SConfiguration, err error) {
	cfg = config.DefaultK8SConfiguration()
	var in *tools.ResourceDataWrapper
	if in = tools.ListWrapper(p); in == nil {
		return
	}
	cfg.Version = in.String("version", cfg.Version)
	cfg.ApiServer = in.String("api_server", cfg.ApiServer)
	cfg.KubeadmConfig = in.String("kubeadm_config", cfg.KubeadmConfig)
	cfg.ApiServerCertExtraSans = in.Set("api_server_cert_extra_sans", cfg.ApiServerCertExtraSans)
	cfg.Repo = in.String("repo", cfg.Repo)
	cfg.Interface = in.String("network_interface", cfg.Interface)
	cfg.PodCIDR = in.String("pod_cidr", cfg.PodCIDR)
	cfg.SvcCIDR = in.String("svc_cidr", cfg.SvcCIDR)
	if v := in.String("certs_validity", cfg.CertsValidity.String()); v != "" {
		if cfg.CertsValidity, err = time.ParseDuration(v); err != nil {
			return
		}
	}
	cfg.Timezone = in.String("timezone", cfg.Timezone)
	cfg.NTPServices = in.Set("ntp_services", cfg.NTPServices)
	return
}

func flattenK8SConfiguration(cfg *config.K8SConfiguration) []map[string]interface{} {
	out := make(map[string]interface{}, 0)
	out["version"] = cfg.Version
	out["api_server"] = cfg.ApiServer
	out["kubeadm_config"] = cfg.KubeadmConfig
	out["api_server_cert_extra_sans"] = tools.SetValue(cfg.ApiServerCertExtraSans)
	out["repo"] = cfg.Repo
	out["network_interface"] = cfg.Interface
	out["pod_cidr"] = cfg.PodCIDR
	out["svc_cidr"] = cfg.SvcCIDR
	out["certs_validity"] = cfg.CertsValidity.String()
	out["timezone"] = cfg.Timezone
	out["ntp_services"] = tools.SetValue(cfg.NTPServices)
	return []map[string]interface{}{out}
}
