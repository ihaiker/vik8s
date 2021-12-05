package terraform

import (
	"context"
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/terraform/tools"
	"log"
	"time"
)

func Vik8sCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: tools.Safe("vik8s_cluster create", createClusterNodeContext),
		ReadWithoutTimeout:   tools.Safe("vik8s_cluster read", readClusterNodeContext),
		UpdateWithoutTimeout: tools.Safe("vik8s_cluster update", updateClusterNodeContext),
		DeleteWithoutTimeout: tools.Safe("vik8s_cluster delete", deleteClusterNodeContext),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: vik8sClusterSchema(),
	}
}

func readClusterNodeContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) diag.Diagnostics {
	nodes, err := expendVIK8SCluster(data, configure)
	if err != nil {
		return diag.FromErr(err)
	}
	return flattenVIK8SCluster(configure, nodes, data)
}

func createClusterNodeContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) diag.Diagnostics {
	nodes, err := expendVIK8SCluster(data, configure)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(base64.StdEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339))))
	_ = data.Set("update_time", time.Now().Format(time.RFC3339))

	if err = checkAllNodes(nodes); err != nil {
		return diag.FromErr(err)
	}

	return flattenVIK8SCluster(configure, nodes, data)
}

func updateClusterNodeContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) (dd diag.Diagnostics) {
	log.Println("update cluster node: ", data.Id())
	return readClusterNodeContext(ctx, data, configure)
}

func deleteClusterNodeContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) (dd diag.Diagnostics) {
	log.Println("delete cluster node: ", data.Id())
	return
}
