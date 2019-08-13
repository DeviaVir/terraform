package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/chef"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: chef.Provisioner,
	})
}
