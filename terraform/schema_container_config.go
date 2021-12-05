package terraform

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/terraform/tools"
)

func vik8sClusterContainerSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"docker": dockerSchema(),
			},
		},
	}
}

func expendContainerConfig(p interface{}) (cfg *config.DockerConfiguration, err error) {
	cfg = config.DefaultDockerConfiguration()
	in := tools.ListWrapper(p)
	if in == nil {
		return
	}
	if d := in.Get("docker"); d != nil {
		if cfg, err = expendDockerConfiguration(d); err != nil {
			return
		}
	}
	return
}

func flattenContainerConfig(cfg *config.DockerConfiguration) interface{} {
	out := map[string]interface{}{
		"docker": flattenDockerConfiguration(cfg),
	}
	return []interface{}{out}
}
