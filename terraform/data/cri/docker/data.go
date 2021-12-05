package docker

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/terraform/tools"
)

func DockerCRI() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: tools.Logger(" vik8s_cri_docker read ", readDockerCRIContext),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: dockerCRISchema(),
	}
}

func readDockerCRIContext(ctx context.Context, data *schema.ResourceData, cfg *config.Configuration) (err error) {
	if cfg.Docker, err = expendDockerConfiguration(data); err != nil {
		return
	}
	data.SetId(tools.Id("cri-docker", cfg.Docker))
	if err = cfg.Write(); err != nil {
		return
	}
	return tools.SetState(flattenDockerConfiguration(cfg.Docker), data)
}
