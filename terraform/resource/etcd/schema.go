package etcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"io/ioutil"
	"path/filepath"
	"time"
)

type etcdSchemaConfig struct {
	*config.ETCDConfiguration
	certs map[string]string
}

func etcdSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"nodes": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"version": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},

		"server_cert_extra_sans": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			ForceNew: true,
		},
		"certs_validity": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"certs_dir": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"data_root": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"snapshot": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"remote_snapshot": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"repo": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},

		"token": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},

		"certs": {
			Type:     schema.TypeMap,
			Computed: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func expendEtcd(data *schema.ResourceData) (cfg *etcdSchemaConfig, err error) {
	in := tools.Data(data)

	cfg = &etcdSchemaConfig{
		ETCDConfiguration: config.DefaultETCDConfiguration(),
		certs:             map[string]string{},
	}

	cfg.Version = in.String("version", cfg.Version)
	cfg.Token = in.String("token", cfg.Token)
	cfg.ServerCertExtraSans = in.Set("server_cert_extra_sans", cfg.ServerCertExtraSans)
	if cfg.CertsValidity, err =
		time.ParseDuration(in.String("", cfg.CertsValidity.String())); err != nil {
		return
	}
	cfg.CertsDir = in.String("certs_dir", cfg.CertsDir)
	cfg.DataRoot = in.String("data_root", cfg.DataRoot)
	cfg.Snapshot = in.String("snapshot", cfg.Snapshot)
	cfg.RemoteSnapshot = in.String("remote_snapshot", cfg.RemoteSnapshot)
	cfg.Repo = in.String("repo", cfg.Repo)
	cfg.Nodes = in.Set("nodes", []string{})

	if v := data.Get("certs"); v != nil {
		certs := v.(map[string]interface{})
		for p, raw := range certs {
			logs.Info("read etcd cert: ", p)
			filename := paths.Join(p)
			if err = utils.Mkdir(filepath.Dir(filename)); err != nil {
				err = utils.Wrap(err, "write certificates file : %s", p)
				return
			}
			if err = ioutil.WriteFile(filename, []byte(raw.(string)), 0666); err != nil {
				err = utils.Wrap(err, "write certificates file : %s", p)
				return
			}
			cfg.certs[p] = raw.(string)
		}
	}
	return
}

func flattenEtcd(etcd *etcdSchemaConfig) interface{} {
	if etcd == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"token":                  etcd.Token,
		"version":                etcd.Version,
		"server_cert_extra_sans": tools.SetValue(etcd.ServerCertExtraSans),
		"certs_validity":         etcd.CertsValidity.String(),
		"certs_dir":              etcd.CertsDir,
		"data_root":              etcd.DataRoot,
		"snapshot":               etcd.Snapshot,
		"remote_snapshot":        etcd.RemoteSnapshot,
		"repo":                   etcd.Repo,
		"nodes":                  etcd.Nodes,
		"certs":                  etcd.certs,
	}}
}
