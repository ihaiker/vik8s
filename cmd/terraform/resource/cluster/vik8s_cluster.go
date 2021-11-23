package cluster

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/configure"
	"github.com/ihaiker/vik8s/cmd/terraform/tools"
	"log"
	"strings"
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

func readClusterNodeContext(ctx context.Context, data *schema.ResourceData, storage *configure.MemStorage) diag.Diagnostics {
	if cfg, err := expendVIK8SCluster(data); err != nil {
		return diag.FromErr(err)
	} else {
		return flattenVIK8SCluster(cfg, data)
	}
}

func createClusterNodeContext(ctx context.Context, data *schema.ResourceData, storage *configure.MemStorage) (dd diag.Diagnostics) {
	id := strings.Replace(data.Get("host_id").(string), "host", "node", 1)
	data.SetId(id)
	_ = data.Set("update_time", time.Now().Format(time.RFC3339))
	return readClusterNodeContext(ctx, data, storage)
}

func updateClusterNodeContext(ctx context.Context, data *schema.ResourceData, storage *configure.MemStorage) (dd diag.Diagnostics) {
	log.Println("update cluster node: ", data.Id())
	return readClusterNodeContext(ctx, data, storage)
}

func deleteClusterNodeContext(ctx context.Context, data *schema.ResourceData, storage *configure.MemStorage) (dd diag.Diagnostics) {
	log.Println("delete cluster node: ", data.Id())
	return
}
