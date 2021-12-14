package cluster

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/cni"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"io/ioutil"
	"path/filepath"
	"time"
)

func vik8sClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"masters": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"slaves": {
			Type:     schema.TypeSet,
			Optional: true,
			Required: false,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"config": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"version": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"api_server": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"api_server_vip": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"kubeadm_config": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"api_server_cert_extra_sans": {
						Type:     schema.TypeSet,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
						ForceNew: true,
					},
					"repo": {
						Type:     schema.TypeString,
						Optional: true,
						ForceNew: true,
						Computed: true,
					},
					"network_interface": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"pod_cidr": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"svc_cidr": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"certs_validity": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ForceNew: true,
					},
					"timezone": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"ntp_services": {
						Type:     schema.TypeSet,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"network": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"flannel": {
						Type:     schema.TypeSet,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"version": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"repo": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"limits_cpu": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  "100m",
								},
								"limits_memory": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  "50Mi",
								},
							},
						},
					},
					"calico": {
						Type:     schema.TypeSet,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"version": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"repo": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"customer": {
						Type:     schema.TypeSet,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"urls": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"files": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
		},
		"api_server_url": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"cluster_yaml": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"certificates": {
			Type:      schema.TypeMap,
			Optional:  true,
			Computed:  true,
			Sensitive: true,
			Elem:      &schema.Schema{Type: schema.TypeString},
		},
		"update_time": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
	}
}

type Network struct {
	Flannel  *cni.Flannel
	Calico   *cni.Calico
	Customer *cni.Customer
}

type Configuration struct {
	Masters      []string
	Slaves       []string
	Network      *Network
	Config       *config.K8SConfiguration
	ApiServerUrl string
	ClusterYaml  string
	Certificates map[string]string
	UpdateTime   string
}

func expendVIK8SCluster(data *schema.ResourceData) (cfg *Configuration, err error) {
	in := tools.Data(data)

	cfg = new(Configuration)
	cfg.Masters = in.Set("masters", []string{})
	cfg.Slaves = in.Set("slaves", []string{})
	cfg.ApiServerUrl = in.String("api_server_url", "")
	cfg.ClusterYaml = in.String("cluster_yaml", "")
	cfg.Certificates = in.Map("certificates", map[string]string{})

	if cfg.Config, err = expendK8SConfiguration(data.Get("config")); err != nil {
		cfg.Config.Masters = make([]string, 0)
		cfg.Config.Nodes = make([]string, 0)
		return
	}
	cfg.Config.Repo = repo.KubeletImage(cfg.Config.Repo)

	if cfg.Network, err = expendNetwork(in.Get("network")); err != nil {
		return
	}

	if v := data.Get("certificates"); v != nil {
		certs := v.(map[string]interface{})
		for p, raw := range certs {
			cfg.Certificates[p] = raw.(string)

			filename := paths.Join(p)
			if err = utils.Mkdir(filepath.Dir(filename)); err != nil {
				return nil, utils.Wrap(err, "write certificates file: %s", p)
			}
			if err = ioutil.WriteFile(filename, []byte(raw.(string)), 0666); err != nil {
				return nil, utils.Wrap(err, "write certificates file: %s", p)
			}
		}
	}

	cfg.UpdateTime = in.String("update_time", "")
	return
}

func expendNetworkFlannel(p interface{}) (f *cni.Flannel, err error) {
	in := tools.SetWrapper(p)
	if in == nil {
		return
	}
	f = cni.NewFlannelCni()
	f.Version = in.String("version", f.Version)
	f.Repo = in.String("repo", f.Repo)
	f.LimitCPU = in.String("limits_cpu", f.LimitCPU)
	f.LimitMemory = in.String("limits_memory", f.LimitMemory)
	return
}

func expendNetworkCalico(p interface{}) (c *cni.Calico, err error) {
	in := tools.SetWrapper(p)
	if in == nil {
		return
	}
	c = cni.NewCalico()
	c.Version = in.String("version", c.Version)
	c.Repo = in.String("repo", c.Repo)
	return
}

func expendNetworkCustomer(p interface{}) (c *cni.Customer, err error) {
	in := tools.SetWrapper(p)
	if in == nil {
		return
	}
	c = new(cni.Customer)
	c.Urls = in.List("urls", []string{})
	c.Files = in.List("files", []string{})
	return
}

func expendNetwork(p interface{}) (nw *Network, err error) {
	nw = new(Network)
	in := tools.SetWrapper(p)
	if in == nil {
		nw.Flannel = cni.NewFlannelCni()
		return
	}
	if nw.Flannel, err = expendNetworkFlannel(in.Get("flannel")); err != nil || nw.Flannel != nil {
		return
	}
	if nw.Calico, err = expendNetworkCalico(in.Get("calico")); err != nil || nw.Calico != nil {
		return
	}
	if nw.Customer, err = expendNetworkCustomer(in.Get("customer")); err != nil || nw.Customer != nil {
		return
	}
	return
}

func expendK8SConfiguration(p interface{}) (cfg *config.K8SConfiguration, err error) {
	cfg = config.DefaultK8SConfiguration()
	var in *tools.ResourceDataWrapper
	if in = tools.ListWrapper(p); in == nil {
		return
	}
	cfg.Version = in.String("version", cfg.Version)
	cfg.ApiServer = in.String("api_server", cfg.ApiServer)
	cfg.ApiServerVIP = in.String("api_server_vip", cfg.ApiServerVIP)
	cfg.KubeadmConfig = in.String("kubeadm_config", cfg.KubeadmConfig)
	cfg.ApiServerCertExtraSans = in.Set("api_server_cert_extra_sans", cfg.ApiServerCertExtraSans)
	cfg.Repo = in.String("repo", cfg.Repo)
	cfg.Interface = in.String("network_interface", cfg.Interface)
	cfg.PodCIDR = in.String("pod_cidr", cfg.PodCIDR)
	cfg.SvcCIDR = in.String("svc_cidr", cfg.SvcCIDR)
	if v := in.String("certs_validity", cfg.CertsValidity.String()); v != "" {
		if cfg.CertsValidity, err = time.ParseDuration(v); err != nil {
			return
		}
	}
	cfg.Timezone = in.String("timezone", cfg.Timezone)
	cfg.NTPServices = in.Set("ntp_services", cfg.NTPServices)
	return
}

func flattenVIK8SCluster(data *schema.ResourceData, cfg *Configuration) (err error) {
	if err = data.Set("update_time", cfg.UpdateTime); err != nil {
		return
	}
	if err = data.Set("masters", cfg.Config.Masters); err != nil {
		return
	}
	if err = data.Set("slaves", cfg.Config.Nodes); err != nil {
		return
	}
	if err = data.Set("config", flattenK8SConfiguration(cfg.Config)); err != nil {
		return
	}
	if err = data.Set("network", flattenNetwork(cfg.Network)); err != nil {
		return
	}
	if err = data.Set("api_server_url", cfg.ApiServerUrl); err != nil {
		return
	}
	if err = data.Set("cluster_yaml", cfg.ClusterYaml); err != nil {
		return
	}
	if err = data.Set("certificates", cfg.Certificates); err != nil {
		return
	}
	return
}

func flattenNetwork(nw *Network) interface{} {
	data := make(map[string]interface{}, 0)
	if nw.Flannel != nil {
		data["flannel"] = []map[string]interface{}{
			{
				"version": nw.Flannel.Version, "repo": nw.Flannel.Repo,
				"limits_cpu": nw.Flannel.LimitCPU, "limits_memory": nw.Flannel.LimitMemory,
			},
		}
	} else if nw.Calico != nil {
		data["calico"] = []map[string]interface{}{
			{
				"version": nw.Calico.Version, "repo": nw.Calico.Repo,
			},
		}
	} else if nw.Customer != nil {
		data["customer"] = []map[string]interface{}{
			{
				"urls":  nw.Customer.Urls,
				"files": nw.Customer.Files,
			},
		}
	}
	return []interface{}{data}
}

func flattenK8SConfiguration(cfg *config.K8SConfiguration) []map[string]interface{} {
	out := make(map[string]interface{}, 0)
	out["version"] = cfg.Version
	out["api_server"] = cfg.ApiServer
	out["api_server_vip"] = cfg.ApiServerVIP
	out["kubeadm_config"] = cfg.KubeadmConfig
	out["api_server_cert_extra_sans"] = tools.SetValue(cfg.ApiServerCertExtraSans)
	out["repo"] = cfg.Repo
	out["network_interface"] = cfg.Interface
	out["pod_cidr"] = cfg.PodCIDR
	out["svc_cidr"] = cfg.SvcCIDR
	out["certs_validity"] = cfg.CertsValidity.String()
	out["timezone"] = cfg.Timezone
	out["ntp_services"] = tools.SetValue(cfg.NTPServices)
	return []map[string]interface{}{out}
}
