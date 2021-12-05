package terraform

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/repo"
)

func vik8sClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"update_time": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"nodes":            nodeSchema(),
		"config":           k8SConfigSchema(),
		"container_config": vik8sClusterContainerSchema(),
	}
}

func expendVIK8SCluster(data *schema.ResourceData, configure *config.Configuration) (nodes nodes, err error) {
	if nodes, err = expendNodes(data.Get("nodes")); err != nil {
		return
	}
	if configure.K8S, err = expendK8SConfiguration(data.Get("config")); err != nil {
		return
	}
	configure.K8S.Repo = repo.KubeletImage(configure.K8S.Repo)
	if configure.Docker, err = expendContainerConfig(data.Get("container_config")); err != nil {
		return
	}
	return
}

func flattenVIK8SCluster(configure *config.Configuration, nodes []*node, data *schema.ResourceData) diag.Diagnostics {
	if err := data.Set("config", flattenK8SConfiguration(configure.K8S)); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("container_config", flattenContainerConfig(configure.Docker)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
