package btree

type BTreeNode struct {
	Position int64
	Leaf     bool
	Size     uint8
	Keys     []int32
	Values   []int32
	Childs   []int64
}

func GetNodeSize(degree uint8) int {
	maxNumberOfKeys := int(degree*2 - 1)
	maxNumberOfChilds := int(degree * 2)

	return 2 + maxNumberOfKeys*8 + maxNumberOfChilds*8
}

func (n *BTreeNode) Insert(index int, key int32, value int32) {

	for i := int(n.Size); i > index; i-- {
		n.Keys[i] = n.Keys[i-1]
		n.Values[i] = n.Values[i-1]
	}

	n.Keys[index] = key
	n.Values[index] = value
	n.Size++
}

func (n *BTreeNode) RemoveKey(index int) {
	n.Keys = append(n.Keys[:index], n.Keys[index+1:]...)
}
