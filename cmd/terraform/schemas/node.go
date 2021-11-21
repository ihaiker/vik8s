package schemas

import (
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"path/filepath"
)

func NodeId(node *ssh.Node) string {
	id := fmt.Sprintf("%v:%v:%v:%v@%v:%v#%v",
		node.User, node.Password, node.PrivateKey, node.Passphrase, node.Host, node.Port, node.Proxy)
	return fmt.Sprintf("host-%x", sha256.Sum224([]byte(id)))
}

func NodeFromResourceData(data *schema.ResourceData) (*ssh.Node, error) {
	username := data.Get("username").(string)
	password := data.Get("password").(string)
	privateKey := data.Get("private_key").(string)
	passphrase := data.Get("passphrase").(string)
	address := data.Get("address").(string)
	port := data.Get("port").(int)
	proxy := data.Get("proxy").(string)

	return &ssh.Node{
		User:       username,
		Password:   password,
		Host:       address,
		Port:       port,
		PrivateKey: privateKey,
		Passphrase: passphrase,
		Proxy:      proxy,
	}, nil
}

func ToResourceData(node *ssh.Node) map[string]interface{} {
	data := make(map[string]interface{}, 0)
	data["username"] = node.User
	data["password"] = node.Password
	data["private_key"] = node.PrivateKey
	data["passphrase"] = node.Passphrase
	data["address"] = node.Host
	data["port"] = node.Port
	data["proxy"] = node.Proxy
	data["id"] = NodeId(node)
	data["facts"] = []map[string]interface{}{{
		"hostname":       node.Facts.Hostname,
		"release_name":   node.Facts.ReleaseName,
		"major_version":  node.Facts.MajorVersion,
		"kernel_version": node.Facts.KernelVersion,
	}}
	return data
}

func Node(id, facts bool) map[string]*schema.Schema {
	node := map[string]*schema.Schema{
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
		},
		"address": {
			Type:        schema.TypeString,
			Description: "ssh address, your can use ip range. example: 192.168.0.10-24, 192.168.0.10-1.25",
			Required:    true,
		},
		"private_key": {
			Type:        schema.TypeString,
			Description: "ssh private key",
			Optional:    true,
			Sensitive:   true,
			DefaultFunc: func() (interface{}, error) {
				if home, err := os.UserHomeDir(); err != nil {
					return "", err
				} else {
					privateKey := filepath.Join(home, ".ssh/id_rsa")
					if utils.Exists(privateKey) {
						return privateKey, nil
					} else {
						return "", err
					}
				}
			},
			ValidateFunc: func(i interface{}, path string) (waring []string, err []error) {
				if i != "" && utils.NotExists(i.(string)) {
					err = []error{utils.Error("private key not found: %v", i)}
				}
				return
			},
		},
		"passphrase": {
			Type:        schema.TypeString,
			Description: "private key passphrase",
			Optional:    true,
		},
		"port": {
			Type:        schema.TypeInt,
			Description: "ssh port",
			Optional:    true,
			Default:     22,
		},
		"proxy": {
			Type:        schema.TypeString,
			Description: "ssh proxy server",
			Optional:    true,
		},
	}
	if id {
		node["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	if facts {
		node["facts"] = &schema.Schema{
			Description: "node facts",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: factsSchema(),
			},
		}
	}
	return node
}

func factsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"hostname": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "node hostname",
		},
		"release_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "os release name. example `centos`,`ubuntu`",
		},
		"major_version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "os majorVersion",
		},
		"kernel_version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "os kernelVersion",
		},
	}
}
