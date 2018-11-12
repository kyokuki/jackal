package enums

type NodeType string

const (
	Collection = NodeType("collection")
	Leaf       = NodeType("leaf")
)

func (x NodeType) String() string {
	return string(x)
}

func NewNodeType(ntype string) NodeType {
	switch ntype {
	case "collection":
		return Collection
	case "leaf":
		return Leaf
	default:
		return NodeType("")
	}

}