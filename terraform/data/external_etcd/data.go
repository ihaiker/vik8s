package external_etcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/terraform/tools"
)

func ExternalEtcd() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: tools.Logger("vik8s_external_etcd read ", externalEtcdReadContext),
		Schema:             externalEtcdSchema(),
	}
}

func externalEtcdReadContext(ctx context.Context, data *schema.ResourceData, configure *config.Configuration) (err error) {
	if configure.ExternalETCD, err = expendExternalEtcd(data); err != nil {
		return
	}
	data.SetId(tools.Id("external-etcd", configure.ExternalETCD))
	return
}
