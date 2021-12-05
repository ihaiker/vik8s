package main

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/ihaiker/vik8s/terraform"
	"log"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:       hclog.Trace,
		JSONFormat:  false,
		DisableTime: true,
	})
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: false}))

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc:        terraform.Provider,
		Logger:              logger,
		NoLogOutputOverride: true,
	})
}
