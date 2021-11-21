package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/ihaiker/vik8s/cmd/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: terraform.Provider,
	})
}
