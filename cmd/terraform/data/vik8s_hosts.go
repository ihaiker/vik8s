package data

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/configure"
	"github.com/ihaiker/vik8s/install/hosts"
)

func Vik8sHosts() *schema.Resource {
	inputs := NodeSchema(false, false)
	inputs["nodes"] = &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: NodeSchema(true, true),
		},
		Computed: true,
	}
	return &schema.Resource{
		ReadWithoutTimeout: hostsReadContext,
		Schema:             inputs,
	}
}

func hostsReadContext(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	storage := i.(*configure.MemStorage)

	node, err := NodeFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(NodeId(node))

	nodes, err := hosts.ParseAddr(*node, node.Host)
	if err != nil {
		return diag.FromErr(err)
	}

	outputs := make([]map[string]interface{}, 0)
	for _, node = range nodes {
		if node.Proxy != "" {
			if proxyNode, has := storage.Hosts[node.Proxy]; !has {
				return diag.FromErr(fmt.Errorf("proxy node not found: %s", node.Proxy))
			} else {
				node.ProxyNode = proxyNode
			}
		}
		if err = node.GatheringFacts(); err != nil {
			return diag.FromErr(err)
		}
		storage.Hosts[NodeId(node)] = node
		outputs = append(outputs, ToResourceData(node))
	}
	if err = data.Set("nodes", outputs); err != nil {
		return diag.FromErr(err)
	}
	return diags
}
