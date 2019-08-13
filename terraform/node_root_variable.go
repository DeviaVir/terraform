package terraform

import (
	"github.com/DeviaVir/terraform/addrs"
	"github.com/DeviaVir/terraform/configs"
	"github.com/DeviaVir/terraform/dag"
)

// NodeRootVariable represents a root variable input.
type NodeRootVariable struct {
	Addr   addrs.InputVariable
	Config *configs.Variable
}

var (
	_ GraphNodeSubPath       = (*NodeRootVariable)(nil)
	_ GraphNodeReferenceable = (*NodeRootVariable)(nil)
	_ dag.GraphNodeDotter    = (*NodeApplyableModuleVariable)(nil)
)

func (n *NodeRootVariable) Name() string {
	return n.Addr.String()
}

// GraphNodeSubPath
func (n *NodeRootVariable) Path() addrs.ModuleInstance {
	return addrs.RootModuleInstance
}

// GraphNodeReferenceable
func (n *NodeRootVariable) ReferenceableAddrs() []addrs.Referenceable {
	return []addrs.Referenceable{n.Addr}
}

// dag.GraphNodeDotter impl.
func (n *NodeRootVariable) DotNode(name string, opts *dag.DotOpts) *dag.DotNode {
	return &dag.DotNode{
		Name: name,
		Attrs: map[string]string{
			"label": n.Name(),
			"shape": "note",
		},
	}
}
