package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/habitat"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: habitat.Provisioner,
	})
}
