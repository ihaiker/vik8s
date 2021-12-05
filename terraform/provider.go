package terraform

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/install/paths"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"china": {
				Type:        schema.TypeBool,
				Description: "use it in china",
				DefaultFunc: schema.EnvDefaultFunc("VIK8S_CHINA", true),
				Required:    false,
				Optional:    true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vik8s_cluster": Vik8sCluster(),
		},
		ConfigureContextFunc: configureProvider,
	}
	return provider
}

func configureProvider(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	paths.China = fmt.Sprint(data.Get("china")) == "true"
	paths.Cloud = "terraform"
	return nil, nil
}
