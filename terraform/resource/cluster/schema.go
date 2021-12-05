package cluster

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
			ForceNew: true,
		},
		"nodes":            nodeSchema(),
		"config":           k8SConfigSchema(),
		"container_config": vik8sClusterContainerSchema(),
		"etcd":             etcdSchema(),
		"external_etcd":    externalEtcdSchema(),

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
	}
}

func expendVIK8SCluster(data *schema.ResourceData, configure *config.Configuration) (nodes nodesConfig, err error) {
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
	if configure.ETCD, err = expendEtcd(data.Get("etcd")); err != nil {
		return
	}
	if configure.ExternalETCD, err = expendExternalEtcd(data.Get("external_etcd")); err != nil {
		return
	}

	/*if v := data.Get("certificates"); v != nil {
		certs := v.(map[string]interface{})
		for p, raw := range certs {
			configure.Certificates[p] = raw.(string)
			filename := paths.Join(p)
			if err = utils.Mkdir(filepath.Dir(filename)); err != nil {
				err = utils.Wrap(err, "write certificates file : %s", p)
				return
			}
			if err = ioutil.WriteFile(filename, []byte(raw.(string)), 0666); err != nil {
				err = utils.Wrap(err, "write certificates file : %s", p)
				return
			}
		}
	}*/
	return
}

func flattenVIK8SCluster(configure *config.Configuration, nodes []*nodeConfig, data *schema.ResourceData) diag.Diagnostics {
	if err := data.Set("config", flattenK8SConfiguration(configure.K8S)); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("container_config", flattenContainerConfig(configure.Docker)); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("etcd", flattenEtcd(configure.ETCD)); err != nil {
		return diag.FromErr(err)
	}
	/*	if err := data.Set("api_server_url", configure.ApiServerUrl); err != nil {
			return diag.FromErr(err)
		}
		if err := data.Set("cluster_yaml", configure.ClusterYaml); err != nil {
			return diag.FromErr(err)
		}
		if err := data.Set("certificates", configure.Certificates); err != nil {
			return diag.FromErr(err)
		}*/
	return nil
}
