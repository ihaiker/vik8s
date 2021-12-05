package cluster

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"github.com/kr/pretty"
	"io/ioutil"
	"path/filepath"
	"time"
)

func externalEtcdSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		MaxItems:      1,
		Optional:      true,
		ConflictsWith: []string{"etcd"},
		ForceNew:      true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"endpoints": {
					Type:     schema.TypeSet,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Required: true,
				},
				"ca": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"ca_raw": {
					Type:          schema.TypeString,
					ConflictsWith: []string{"external_etcd.0.ca"},
					AtLeastOneOf:  []string{"external_etcd.0.ca"},
					Optional:      true,
					ForceNew:      true,
				},
				"cert": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"cert_raw": {
					Type:          schema.TypeString,
					ConflictsWith: []string{"external_etcd.0.cert"},
					AtLeastOneOf:  []string{"external_etcd.0.cert"},
					Optional:      true,
					ForceNew:      true,
				},
				"key": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"key_raw": {
					Type:          schema.TypeString,
					ConflictsWith: []string{"external_etcd.0.key"},
					AtLeastOneOf:  []string{"external_etcd.0.key"},
					Optional:      true,
					ForceNew:      true,
				},
			},
		},
	}
}
func etcdSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"token": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"version": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
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
					Computed: true,
					ForceNew: true,
				},
				"remote_snapshot": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"repo": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
			},
		},
	}
}

func expendEtcd(p interface{}) (cfg *config.ETCDConfiguration, err error) {
	in := tools.ListWrapper(p)
	if in == nil {
		return
	}
	cfg = config.DefaultETCDConfiguration()
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
	return
}

func expendExternalEtcd(p interface{}) (cfg *config.ExternalETCDConfiguration, err error) {
	in := tools.ListWrapper(p)
	if in == nil {
		return
	}
	dir := etcd.CertsDir()
	if err = utils.Mkdir(dir); err != nil {
		return
	}
	cfg = new(config.ExternalETCDConfiguration)
	cfg.Endpoints = in.Set("endpoints", cfg.Endpoints)

	if caRaw := in.String("ca_raw", ""); caRaw != "" {
		caFile := filepath.Join(dir, "ca.crt")
		if err = ioutil.WriteFile(caFile, []byte(caRaw), 0666); err != nil {
			return
		}
		cfg.CaFile = caFile
	} else {
		cfg.CaFile = in.String("ca", "")
	}
	if utils.NotExists(cfg.CaFile) {
		err = utils.Error("ca file not found: %s", cfg.CaFile)
		return
	}

	if certRaw := in.String("cert_raw", ""); certRaw != "" {
		caFile := filepath.Join(dir, "client.crt")
		if err = ioutil.WriteFile(caFile, []byte(certRaw), 0666); err != nil {
			return
		}
		cfg.Cert = caFile
	} else {
		cfg.Cert = in.String("cert", "")
	}
	if utils.NotExists(cfg.CaFile) {
		err = utils.Error("cert file not found: %s", cfg.CaFile)
		return
	}

	if keyRaw := in.String("key_raw", ""); keyRaw != "" {
		caFile := filepath.Join(dir, "server.key")
		if err = ioutil.WriteFile(caFile, []byte(keyRaw), 0666); err != nil {
			return
		}
		cfg.Key = caFile
	} else {
		cfg.Key = in.String("key", "")
	}
	if utils.NotExists(cfg.CaFile) {
		err = utils.Error("key file not found: %s", cfg.CaFile)
		return
	}
	pretty.Log(cfg)
	return
}

func flattenEtcd(etcd *config.ETCDConfiguration) interface{} {
	if etcd == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"token":                  etcd.Token,
			"version":                etcd.Version,
			"server_cert_extra_sans": tools.SetValue(etcd.ServerCertExtraSans),
			"certs_validity":         etcd.CertsValidity.String(),
			"certs_dir":              etcd.CertsDir,
			"data_root":              etcd.DataRoot,
			"snapshot":               etcd.Snapshot,
			"remote_snapshot":        etcd.RemoteSnapshot,
			"repo":                   etcd.Repo,
		},
	}
}
