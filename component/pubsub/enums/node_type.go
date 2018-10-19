package enums

type NodeType string

const (
	Collection = NodeType("collection")
	Leaf       = NodeType("leaf")
)

func (x NodeType) String() string {
	return string(x)
}