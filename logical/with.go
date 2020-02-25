package logical

import (
	"context"

	"github.com/pkg/errors"

	"github.com/cube2222/octosql"
	"github.com/cube2222/octosql/physical"
)

type With struct {
	cteNames []string
	cteNodes []Node
	source   Node
}

func NewWith(cteNames []string, cteNodes []Node, source Node) *With {
	return &With{
		cteNames: cteNames,
		cteNodes: cteNodes,
		source:   source,
	}
}

func (node *With) Physical(ctx context.Context, physicalCreator *PhysicalPlanCreator) (physical.Node, octosql.Variables, error) {
	variables := octosql.NoVariables()
	for i := range node.cteNodes {
		physicalNode, nodeVariables, err := node.cteNodes[i].Physical(ctx, physicalCreator)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err,
				"couldn't get physical plan for common table expression %s with index %d", node.cteNames[i], i,
			)
		}
		variables, err = variables.MergeWith(nodeVariables)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err,
				"couldn't merge variables with those of common table expression %s with index %d", node.cteNames[i], i,
			)
		}

		physicalCreator = physicalCreator.WithCommonTableExpression(node.cteNames[i], physicalNode)
	}

	source, childVariables, err := node.source.Physical(ctx, physicalCreator)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't get physical plan for with source node")
	}

	variables, err = childVariables.MergeWith(variables)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't merge variables for with source")
	}

	return source, variables, nil
}
