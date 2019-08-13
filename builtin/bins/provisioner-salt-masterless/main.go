package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/salt-masterless"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: saltmasterless.Provisioner,
	})
}
