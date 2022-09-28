package btree

type BTree struct {
	Root   *BTreeNode
	Degree uint8
	File   IOBTreeFile
}

func (t *BTree) allocateNode() int64 {
	nodePosition := t.File.AllocateBlock()
	return nodePosition
}

func (t *BTree) Close() {
	t.File.UpdateDescriptor()
	t.File.Close()
}

func (t *BTree) CreateNode() *BTreeNode {
	node := BTreeNode{
		Position: 0,
		Leaf:     false,
		Size:     0,
		Keys:     make([]int32, t.Degree*2-1),
		Values:   make([]int32, t.Degree*2-1),
		Childs:   make([]int64, t.Degree*2),
	}

	return &node
}

func (t *BTree) CreateNodeAndAllocate() *BTreeNode {
	node := t.CreateNode()
	node.Position = t.allocateNode()
	return node
}

func CreateTree(io IOBTree, fullFileName string, degree uint8) *BTree {
	tree := BTree{
		Degree: degree,
	}

	file, _ := io.CreateFile(fullFileName, &tree, int(degree)*2*4)

	tree.File = file
	tree.Root = tree.CreateNodeAndAllocate()
	tree.Root.Leaf = true
	file.WriteNode(tree.Root)

	return &tree
}

func OpenTree(io IOBTree, fullFileName string) *BTree {
	file, _ := io.OpenFile(fullFileName)

	tree := BTree{
		Degree: uint8(file.GetTreeDegree()),
		File:   file,
	}

	root := tree.CreateNode()
	root.Position = file.GetRootBlock()
	file.ReadNode(root)
	tree.Root = root

	return &tree
}

func (t *BTree) Insert(key int32, value int32) {
	if t.Root.Size == 2*t.Degree-1 {
		root := t.CreateNodeAndAllocate()
		root.Childs[0] = t.Root.Position
		child := t.Root
		t.Root = root
		t.File.UpdateRootPosition(root.Position)
		t.splitChild(root, child, 0)
	}

	t.insertNonFull(t.Root, key, value)
}

func (t *BTree) Find(key int32) (*BTreeNode, int) {
	return t.findNode(t.Root, key)
}

func (t *BTree) findNode(node *BTreeNode, key int32) (*BTreeNode, int) {
	var index int = 0

	for index < int(node.Size) && key > node.Keys[index] {
		index += 1
	}

	if index < int(node.Size) && key == node.Keys[index] {
		return node, index
	}

	if node.Leaf {
		return nil, -1
	}

	child := t.CreateNode()
	child.Position = node.Childs[index]
	t.File.ReadNode(child)

	return t.findNode(child, key)
}

func (t *BTree) insertNonFull(node *BTreeNode, key int32, value int32) {
	var index int = 0

	if node.Leaf {
		for index < int(node.Size) && key > node.Keys[index] {
			index += 1
		}
		node.Insert(index, key, value)
		t.File.WriteNode(node)
	} else {
		for index < int(node.Size) && key > node.Keys[index] {
			index += 1
		}

		child := t.CreateNode()
		child.Position = node.Childs[index]
		t.File.ReadNode(child)

		if child.Size == 2*t.Degree-1 {
			t.splitChild(node, child, index)

			if key > node.Keys[index] {
				index += 1
				child = t.CreateNode()
				child.Position = node.Childs[index]
				t.File.ReadNode(child)
			}
		}

		t.insertNonFull(child, key, value)
	}

}

func (t *BTree) splitChild(parentNode, childNode *BTreeNode, index int) {
	node := t.CreateNodeAndAllocate()
	node.Size = t.Degree - 1
	node.Leaf = childNode.Leaf

	copy(node.Keys, childNode.Keys[t.Degree:2*t.Degree-1])
	copy(node.Values, childNode.Values[t.Degree:2*t.Degree-1])
	if !childNode.Leaf {
		copy(node.Childs, childNode.Childs[t.Degree:2*t.Degree])
	}

	parentNode.Insert(index, childNode.Keys[t.Degree-1], childNode.Values[t.Degree-1])
	parentNode.Childs[parentNode.Size] = node.Position
	copy(childNode.Keys, childNode.Keys[0:t.Degree-1])
	copy(childNode.Values, childNode.Values[0:t.Degree-1])
	childNode.Size = t.Degree - 1

	t.File.WriteNode(parentNode)
	t.File.WriteNode(node)
	t.File.WriteNode(childNode)
}
