package data

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/config"
	"github.com/ihaiker/vik8s/cmd/terraform/schemas"
)

func Vik8sHost() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: hostReadContext,
		Schema:             schemas.Node(false, true),
	}
}

func hostReadContext(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	node, err := schemas.NodeFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	storage := i.(*config.MemStorage)

	if node.Proxy != "" {
		if proxyNode, has := storage.Hosts[node.Proxy]; !has {
			return diag.FromErr(fmt.Errorf("proxy node not found: %s", node.Proxy))
		} else {
			node.ProxyNode = proxyNode
		}
	}
	if err := node.GatheringFacts(); err != nil {
		return diag.FromErr(err)
	}
	id := schemas.NodeId(node)
	data.SetId(schemas.NodeId(node))
	nodeData := schemas.ToResourceData(node)
	if err := data.Set("facts", nodeData["facts"]); err != nil {
		return diag.FromErr(err)
	}
	storage.Hosts[id] = node
	return diag.Diagnostics{}
}
