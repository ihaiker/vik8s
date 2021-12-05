package hosts

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/terraform/tools"
)

func hostsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"host": {
			Type:        schema.TypeString,
			Description: "ssh host. example: 192.168.0.9/192.168.0.10",
			Optional:    true,
		},
		"hosts": {
			Type:        schema.TypeSet,
			Description: "ssh hosts",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"address": {
			Type:         schema.TypeString,
			Description:  "ssh address, your can use ip range. example: 192.168.0.10-24, 192.168.0.10-1.25",
			Optional:     true,
			AtLeastOneOf: []string{"host", "hosts"},
		},
		"port": {
			Type:        schema.TypeInt,
			Description: "ssh port",
			Optional:    true,
			Default:     22,
		},
		"username": {
			Type:        schema.TypeString,
			Default:     "root",
			Optional:    true,
			Description: `ssh username`,
		},
		"password": {
			Type:        schema.TypeString,
			Description: `ssh password`,
			Optional:    true,
			Sensitive:   true,
			Default:     "",
		},
		"ssh_key": {
			Type:          schema.TypeString,
			Description:   "ssh private key",
			Optional:      true,
			Default:       "",
			ConflictsWith: []string{"password"},
		},
		"ssh_key_raw": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "SSH Private Key",
			Default:       "",
			ConflictsWith: []string{"ssh_key", "password"},
		},
		"passphrase": {
			Type:        schema.TypeString,
			Description: "private key passphrase",
			Optional:    true,
			Sensitive:   true,
			Default:     "",
		},
		"bastion": {
			Type:        schema.TypeString,
			Description: "bastion Host configuration server",
			Optional:    true,
			Default:     "",
		},

		"nodes": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func expendOption(data *schema.ResourceData, host string) *hosts.Option {
	in := tools.Data(data)
	return &hosts.Option{
		Port:          in.Int("port"),
		User:          in.String("username", "root"),
		Password:      in.String("password", ""),
		PrivateKey:    in.String("ssh_key", ""),
		PrivateKeyRaw: in.String("ssh_key_raw", ""),
		Passphrase:    in.String("passphrase", ""),
		Host:          host,
		Proxy:         in.String("bastion", ""),
	}
}

func expendNodes(data *schema.ResourceData) (ssh.Nodes, error) {
	opt := expendOption(data, "")

	in := tools.Data(data)
	nodes := make([]*ssh.Node, 0)

	if host := in.String("host", ""); host != "" {
		opt.Host = host
		nodes = append(nodes, opt)
	}

	if address := in.String("address", ""); address != "" {
		if addressNodes, err := hosts.ParseAddr(*opt, address); err != nil {
			return nil, err
		} else {
			nodes = append(nodes, addressNodes...)
		}
	}

	if _hosts := in.Set("hosts", []string{}); _hosts != nil || len(_hosts) == 0 {
		for _, host := range _hosts {
			nodes = append(nodes, expendOption(data, host))
		}
	}
	return nodes, nil
}
