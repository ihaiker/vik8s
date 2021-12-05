package cluster

import (
	"context"
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/terraform/tools"
	"log"
	"strings"
	"time"
)

type input = func(context.Context, *schema.ResourceData, *config.Configuration) diag.Diagnostics
type output = func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics

func safe(method string, fn input) output {
	return func(ctx context.Context, data *schema.ResourceData, i interface{}) (dd diag.Diagnostics) {
		defer utils.Catch(func(err error) {
			log.Println(utils.Stack())
			dd = diag.FromErr(err)
		})
		log.Println(strings.Repeat("<", 30), method, strings.Repeat("<", 30))
		defer func() { log.Println(strings.Repeat(">", 30), method, strings.Repeat("<", 30)) }()
		return fn(ctx, data, i.(*config.Configuration))
	}
}

func Vik8sCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: safe("create", createClusterNodeContext),
		ReadWithoutTimeout:   safe("read", readClusterNodeContext),
		UpdateWithoutTimeout: safe("update", updateClusterNodeContext),
		DeleteWithoutTimeout: safe("delete", deleteClusterNodeContext),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customizeDiff,
		Schema:        vik8sClusterSchema(),
	}
}

func customizeDiff(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
	if len(diff.UpdatedKeys()) == 0 {
		return nil
	}

	configure := i.(*config.Configuration)
	if oldConfig, newConfig := diff.GetChange("etcd"); tools.Length(oldConfig) == 0 && tools.Length(newConfig) == 1 {
		return utils.Error("Does not support modification etcd after initialization")
	}

	if configure.ETCD != nil && diff.HasChange("external_etcd") {
		return utils.Error("Does not support modification to use an external ETCD cluster after initialization")
	}

	if configure.ETCD != nil && diff.HasChange("nodes") {
		if nodes, err := expendNodes(diff.Get("nodes")); err != nil {
			return err
		} else if len(nodes.roleNode("etcd")) == 0 {
			return utils.Error("At least one `etcd` role node must exist")
		}
	}

	return nil
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

	if err = checkAllNodes(configure, nodes); err != nil {
		return diag.FromErr(err)
	}
	if err = crateEtcd(configure, nodes, data); err != nil {
		return diag.FromErr(err)
	}
	return flattenVIK8SCluster(configure, nodes, data)
}

func updateClusterNodeContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) diag.Diagnostics {
	nodes, err := expendVIK8SCluster(data, configure)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = checkAllNodes(configure, nodes); err != nil {
		return diag.FromErr(err)
	}

	if err = updateEtcd(configure, nodes, data); err != nil {
		return diag.FromErr(err)
	}

	return flattenVIK8SCluster(configure, nodes, data)
}

func deleteClusterNodeContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) (dd diag.Diagnostics) {
	log.Println("delete cluster node: ", data.Id())
	return
}
