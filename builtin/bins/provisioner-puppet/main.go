package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/puppet"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: puppet.Provisioner,
	})
}
