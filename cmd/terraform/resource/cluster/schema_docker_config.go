package cluster

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/tools"
	"github.com/ihaiker/vik8s/config"
)

func dockerSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"version": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "docker version",
				},
				"strait_version": {
					Type:        schema.TypeBool,
					Optional:    true,
					Computed:    true,
					Description: "Strict check DOCKER version if inconsistent will upgrade",
				},
				"skip_if_exists": {
					Type:        schema.TypeBool,
					Optional:    true,
					Computed:    true,
					Description: "skip install and change anything if exists docker.",
				},
				"data_root": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"hosts": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"daemon_json": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "docker cfg file, if set this option, other option will ignore.",
				},
				"insecure_registries": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"registry_mirrors": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},

				"storage": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"driver": {
								Type:     schema.TypeString,
								Required: true,
							},
							"opt": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
				"dns": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"list": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"opt": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"search": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"tls": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ca": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Trust certs signed only by this CA",
							},
							"key": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Path to TLS certificate file",
							},
							"cert": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Path to TLS key file",
							},
						},
					},
				},
			},
		},
	}
}

func expendDockerConfiguration(p interface{}) (cfg *config.DockerConfiguration, err error) {
	cfg = config.DefaultDockerConfiguration()
	in := tools.ListWrapper(p)
	cfg.Version = in.String("version", cfg.Version)
	cfg.StraitVersion = in.Bool("strait_version")
	cfg.DataRoot = in.String("data_root", cfg.DataRoot)
	cfg.Hosts = in.Set("hosts", cfg.Hosts)
	cfg.DaemonJson = in.String("daemon_json", cfg.DaemonJson)
	cfg.InsecureRegistries = in.Set("insecure_registries", cfg.InsecureRegistries)
	cfg.RegistryMirrors = in.Set("registry_mirrors", cfg.RegistryMirrors)
	if cfg.Storage, err = expendDockerStorageConfiguration(in.Get("storage")); err != nil {
		return
	}
	if cfg.DNS, err = expendDockerDNSConfiguration(in.Get("dns")); err != nil {
		return
	}
	if cfg.TLS, err = expendDockerTLSConfiguration(in.Get("tls")); err != nil {
		return
	}
	return
}

func expendDockerStorageConfiguration(p interface{}) (cfg *config.DockerStorageConfiguration, err error) {
	in := tools.ListWrapper(p)
	if in == nil {
		return
	}
	cfg = new(config.DockerStorageConfiguration)
	cfg.Driver = in.String("driver", cfg.Driver)
	cfg.Opt = in.Set("opt", cfg.Opt)
	return
}

func expendDockerDNSConfiguration(p interface{}) (cfg *config.DockerDNSConfiguration, err error) {
	in := tools.ListWrapper(p)
	if in == nil {
		return
	}
	cfg = new(config.DockerDNSConfiguration)
	cfg.List = in.Set("list", []string{})
	cfg.Opt = in.Set("opt", []string{})
	cfg.Search = in.Set("search", []string{})
	return
}

func expendDockerTLSConfiguration(p interface{}) (cfg *config.DockerCertsConfiguration, err error) {
	in := tools.ListWrapper(p)
	if in == nil {
		return
	}
	cfg = new(config.DockerCertsConfiguration)
	cfg.CaCertPath = in.String("ca", "")
	cfg.ServerCertPath = in.String("cert", "")
	cfg.ServerKeyPath = in.String("key", "")
	return
}

func flattenDockerConfiguration(cfg *config.DockerConfiguration) interface{} {
	data := map[string]interface{}{
		"version":             cfg.Version,
		"strait_version":      cfg.StraitVersion,
		"data_root":           cfg.DataRoot,
		"hosts":               cfg.Hosts,
		"daemon_json":         cfg.DaemonJson,
		"insecure_registries": cfg.InsecureRegistries,
		"registry_mirrors":    cfg.RegistryMirrors,
	}

	if storage := flattenDockerStorageConfiguration(cfg.Storage); storage != nil {
		data["storage"] = storage
	}
	if dns := flattenDockerDNSConfiguration(cfg.DNS); dns != nil {
		data["dns"] = dns
	}
	if tls := flattenDockerTLSConfiguration(cfg.TLS); tls != nil {
		data["tls"] = tls
	}
	return []interface{}{data}
}

func flattenDockerStorageConfiguration(cfg *config.DockerStorageConfiguration) interface{} {
	if cfg == nil {
		return []interface{}{}
	}
	return []interface{}{map[string]interface{}{
		"driver": cfg.Driver,
		"opt":    tools.SetValue(cfg.Opt),
	}}
}

func flattenDockerDNSConfiguration(cfg *config.DockerDNSConfiguration) interface{} {
	if cfg == nil {
		return []interface{}{}
	}
	return []interface{}{map[string]interface{}{
		"list":   cfg.List,
		"opt":    cfg.Opt,
		"search": cfg.Search,
	}}
}

func flattenDockerTLSConfiguration(cfg *config.DockerCertsConfiguration) interface{} {
	if cfg == nil {
		return []interface{}{}
	}
	return []interface{}{map[string]interface{}{
		"ca":   cfg.CaCertPath,
		"cert": cfg.ServerCertPath,
		"key":  cfg.ServerKeyPath,
	}}
}
