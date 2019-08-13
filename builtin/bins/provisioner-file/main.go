package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/file"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: file.Provisioner,
	})
}
