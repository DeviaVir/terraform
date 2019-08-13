package main

import (
	"github.com/DeviaVir/terraform/builtin/provisioners/remote-exec"
	"github.com/DeviaVir/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: remoteexec.Provisioner,
	})
}
