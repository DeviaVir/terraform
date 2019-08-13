package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/local-exec"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: localexec.Provisioner,
	})
}
