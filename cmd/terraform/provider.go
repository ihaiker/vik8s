package terraform

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/configure"
	"github.com/ihaiker/vik8s/cmd/terraform/data"
	"github.com/ihaiker/vik8s/cmd/terraform/resource/cluster"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"vik8s_cluster": cluster.Vik8sCluster(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"vik8s_host":  data.Vik8sHost(),
			"vik8s_hosts": data.Vik8sHosts(),
		},
		ConfigureContextFunc: func(ctx context.Context, resourceData *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return &configure.MemStorage{
				Hosts: make(map[string]*ssh.Node, 0),
			}, nil
		},
	}
	return provider
}
