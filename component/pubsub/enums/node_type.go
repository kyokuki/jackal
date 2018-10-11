package enums

type nodeType string

const (
	Collection = nodeType("collection")
	Leaf       = nodeType("leaf")
)

func (x nodeType) String() string {
	return string(x)
}