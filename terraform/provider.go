package terraform

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	hs "github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/logs"
	cri_docker "github.com/ihaiker/vik8s/terraform/data/cri/docker"
	"github.com/ihaiker/vik8s/terraform/data/external_etcd"
	hostsData "github.com/ihaiker/vik8s/terraform/data/hosts"
	"github.com/ihaiker/vik8s/terraform/resource/etcd"
	"log"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"china": {
				Type:        schema.TypeBool,
				Description: "use it in china",
				DefaultFunc: schema.EnvDefaultFunc("VIK8S_CHINA", true),
				Optional:    true,
			},
			"cloud": {
				Type:     schema.TypeString,
				Default:  "terraform",
				Optional: true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"vik8s_hosts":         hostsData.Vik8sHosts(),
			"vik8s_cri_docker":    cri_docker.DockerCRI(),
			"vik8s_external_etcd": external_etcd.ExternalEtcd(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"vik8s_etcd": etcd.Vik8sETCD(),
		},
		ConfigureContextFunc: configureProvider,
	}
	return provider
}

func configureProvider(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	logs.SetOutput(log.Writer())

	paths.China = fmt.Sprint(data.Get("china")) == "true"
	paths.Cloud = data.Get("cloud").(string)
	paths.IsTerraform = true

	logs.Infof("use config dir: %s/%s", paths.ConfigDir, paths.Cloud)

	var configure *config.Configuration
	var err error

	if configure, err = config.Load(paths.Vik8sConfiguration()); err != nil {
		return nil, diag.FromErr(err)
	} else if configure.Hosts, err = hs.New(paths.HostsConfiguration(), hs.Option{Port: 22, User: "root", PrivateKey: "$HOME/.ssh/id_rsa"}); err != nil {
		return nil, diag.FromErr(err)
	} else {
		return configure, nil
	}
}
