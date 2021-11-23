package cluster

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
)

type vik8sConfig struct {
	master       bool
	hostId       string
	config       *config.K8SConfiguration
	dockerConfig *config.DockerConfiguration
}

func vik8sClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"master": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		},
		"host_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"update_time": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"config":           k8SConfigSchema(),
		"container_config": vik8sClusterContainerSchema(),
	}
}

func expendVIK8SCluster(data *schema.ResourceData) (cfg *vik8sConfig, err error) {
	cfg = new(vik8sConfig)
	if v, ok := data.GetOk("master"); ok {
		cfg.master = v.(bool)
	}
	cfg.hostId = data.Get("host_id").(string)
	if cfg.config, err = expendK8SConfiguration(data.Get("config")); err != nil {
		return
	}
	if cfg.dockerConfig, err = expendContainerConfig(data.Get("container_config")); err != nil {
		return
	}
	return
}

func flattenVIK8SCluster(cfg *vik8sConfig, data *schema.ResourceData) diag.Diagnostics {
	if err := data.Set("master", cfg.master); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("config", flattenK8SConfiguration(cfg.config)); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("container_config", flattenContainerConfig(cfg.dockerConfig)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
