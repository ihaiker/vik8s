package external_etcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"io/ioutil"
	"path/filepath"
)

func externalEtcdSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"endpoints": {
			Type:     schema.TypeSet,
			MinItems: 1,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Required: true,
		},
		"ca": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"ca_raw": {
			Type:          schema.TypeString,
			ConflictsWith: []string{"ca"},
			AtLeastOneOf:  []string{"ca"},
			Optional:      true,
		},
		"cert": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"cert_raw": {
			Type:          schema.TypeString,
			ConflictsWith: []string{"cert"},
			AtLeastOneOf:  []string{"cert"},
			Optional:      true,
		},
		"key": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"key_raw": {
			Type:          schema.TypeString,
			ConflictsWith: []string{"key"},
			AtLeastOneOf:  []string{"key"},
			Optional:      true,
		},
	}
}

func expendExternalEtcd(data *schema.ResourceData) (cfg *config.ExternalETCDConfiguration, err error) {
	in := tools.Data(data)
	dir := etcd.CertsDir()
	if err = utils.Mkdir(dir); err != nil {
		return
	}

	cfg = new(config.ExternalETCDConfiguration)
	cfg.Endpoints = in.Set("endpoints", cfg.Endpoints)

	items := []string{"ca", "cert", "key"}
	for _, name := range items {
		var file string
		if raw := in.String(name+"_raw", ""); raw != "" {
			file = filepath.Join(dir, name+".crt")
			if err = ioutil.WriteFile(file, []byte(raw), 0666); err != nil {
				return
			}
		} else {
			file = in.String(name, "")
		}
		if utils.NotExists(file) {
			return nil, utils.Error("ca file not found: %s", file)
		}
		cfg.Set(name, file)
	}
	return
}
