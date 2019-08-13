package main

import (
	"github.com/DeviaVir/terraform/builtin/providers/test"
	"github.com/DeviaVir/terraform/plugin"
	"github.com/DeviaVir/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return test.Provider()
		},
	})
}
