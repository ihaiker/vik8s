package hosts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/terraform/tools"
)

func Vik8sHosts() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: tools.Logger(" vik8s_hosts read ", hostsReadContext),
		Schema:             hostsSchema(),
	}
}

func hostsReadContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) (err error) {
	var nodes ssh.Nodes
	if nodes, err = expendNodes(data); err != nil {
		return
	}
	if err = tools.GatheringFacts(nodes, configure.Hosts); err != nil {
		return
	}
	_ = data.Set("nodes", nodes.Hosts())
	data.SetId(tools.Id("hosts", nodes))
	return
}
