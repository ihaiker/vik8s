package terraform

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/data"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"vik8s_host":  data.Vik8sHost(),
			"vik8s_hosts": data.Vik8sHosts(),
		},
		ConfigureContextFunc: func(ctx context.Context, resourceData *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return nil, nil
		},
	}
}
