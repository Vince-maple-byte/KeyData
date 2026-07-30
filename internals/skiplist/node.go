package skiplist

type Node struct {
	key    string
	value  []byte
	levels []*Node
}

func (n *Node) GetKey() string {
	return n.key
}

func (n *Node) GetValue() []byte {
	return n.value
}

func (n *Node) GetLevels() []*Node {
	return n.levels
}
