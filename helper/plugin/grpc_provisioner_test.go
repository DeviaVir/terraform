package plugin

import proto "github.com/DeviaVir/terraform/internal/tfplugin5"

var _ proto.ProvisionerServer = (*GRPCProvisionerServer)(nil)
