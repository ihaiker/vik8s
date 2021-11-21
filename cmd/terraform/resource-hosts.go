package terraform

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func HoseResource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCoffeesRead,
		Schema:      map[string]*schema.Schema{},
	}
}
