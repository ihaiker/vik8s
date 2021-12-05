package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/ihaiker/vik8s/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: terraform.Provider,
	})
}
