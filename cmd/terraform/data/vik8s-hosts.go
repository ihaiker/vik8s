package data

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/schemas"
	"github.com/ihaiker/vik8s/install/hosts"
	"strconv"
)

func Vik8sHosts() *schema.Resource {

	outputs := schemas.Node(false)
	outputs["nodes"] = &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: schemas.Node(true),
		},
		Computed: true,
	}
	return &schema.Resource{
		ReadWithoutTimeout: hostsReadContext,
		Schema:             outputs,
	}
}

func hostsReadContext(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	node, err := schemas.NodeFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(schemas.NodeId(node))

	opt := hosts.Option{
		User: node.User, Password: node.Password,
		PrivateKey: node.PrivateKey, Passphrase: node.Passphrase, Proxy: node.Proxy,
	}
	opt.Port, _ = strconv.Atoi(node.Port)
	nodes, err := hosts.ParseAddr(opt, node.Host)
	if err != nil {
		return diag.FromErr(err)
	}

	outputs := make([]map[string]interface{}, 0)
	for _, node := range nodes {
		if err = node.GatheringFacts(); err != nil {
			return diag.FromErr(err)
		}
		port, _ := strconv.Atoi(node.Port)
		id := fmt.Sprintf("%v:%v:%v:%v@%v:%v#%v", node.User, node.Password,
			node.PrivateKey, node.Passphrase, node.Host, port, node.Proxy)
		output := map[string]interface{}{
			"id":       fmt.Sprintf("%x", sha256.Sum256([]byte(id))),
			"username": node.User, "password": node.Password, "address": node.Host,
			"private_key": node.PrivateKey, "passphrase": node.Passphrase, "port": port,
			"proxy": node.Proxy,
		}
		outputs = append(outputs, output)
	}
	if err := data.Set("nodes", outputs); err != nil {
		return diag.FromErr(err)
	}
	return diags
}
