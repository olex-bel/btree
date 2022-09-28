package test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/olex-bel/btree/btree"
)

type TestIO struct {
}

type TestFile struct {
	data       []byte
	position   int64
	Descriptor *btree.BTreeFileDescriptor
}

func (f *TestFile) UpdateDescriptor() error {
	return nil
}

func (f *TestFile) WriteNode(node *btree.BTreeNode) error {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, &node.Leaf)
	binary.Write(&buffer, binary.LittleEndian, &node.Size)
	binary.Write(&buffer, binary.LittleEndian, node.Keys)
	binary.Write(&buffer, binary.LittleEndian, node.Values)
	binary.Write(&buffer, binary.LittleEndian, node.Childs)

	position := int(node.Position)

	for i, value := range buffer.Bytes() {
		f.data[position+i] = value
	}

	return nil
}

func (f *TestFile) ReadNode(node *btree.BTreeNode) error {
	buffer := bytes.NewBuffer(f.data[int(node.Position):])
	binary.Read(buffer, binary.LittleEndian, &node.Leaf)
	binary.Read(buffer, binary.LittleEndian, &node.Size)
	binary.Read(buffer, binary.LittleEndian, node.Keys)
	binary.Read(buffer, binary.LittleEndian, node.Values)
	binary.Read(buffer, binary.LittleEndian, node.Childs)
	return nil
}

func (f *TestFile) PrintNode(node *btree.BTreeNode) {

}

func (f *TestFile) AllocateBlock() int64 {
	length := btree.GetNodeSize(uint8(f.Descriptor.TreeDegree))
	data := make([]byte, length)
	position := f.Descriptor.FirstFreeBlock
	f.Descriptor.FirstFreeBlock += int64(length)
	f.data = append(f.data, data...)
	return position
}

func (f *TestFile) UpdateRootPosition(position int64) {
	f.Descriptor.RootNodePosition = position
}

func (f *TestFile) GetTreeDegree() int16 {
	return f.Descriptor.TreeDegree
}

func (f *TestFile) GetRootBlock() int64 {
	return f.Descriptor.RootNodePosition
}

func (f *TestFile) Close() {
}

func (i *TestIO) CreateFile(fullFileName string, tree *btree.BTree, pageSize int) (btree.IOBTreeFile, error) {
	header := btree.BTreeFileDescriptor{
		TreeDegree:     int16(tree.Degree),
		PageSize:       int16(pageSize),
		FirstFreeBlock: btree.HEADER_SIZE,
	}
	btreeFile := TestFile{
		Descriptor: &header,
		data:       make([]byte, 0),
	}

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.LittleEndian, &header)
	btreeFile.data = append(btreeFile.data, []byte{64, 62}...)
	btreeFile.data = append(btreeFile.data, buffer.Bytes()...)

	return &btreeFile, nil
}

func (i *TestIO) OpenFile(fullFileName string) (btree.IOBTreeFile, error) {
	file := TestFile{
		position: 0,
		data:     make([]byte, 0),
	}
	return &file, nil
}

func (i *TestIO) WriteNode(node *btree.BTreeNode) error {
	return nil
}

func TestCreateTree(t *testing.T) {
	var degree uint8 = 2
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", degree)

	if tree.Degree != degree {
		t.Fatalf("Expected degree %d, got %d\n", degree, tree.Degree)
	}

	if tree.Root == nil {
		t.Fatalf("Root must not be nul")
	}

	if tree.Root.Size != 0 {
		t.Fatalf("Root keys list must be empty")
	}
}

func TestInsertOneKey(t *testing.T) {
	var key int32 = 4
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)
	tree.Insert(key, 0)

	if tree.Root.Size != 1 {
		t.Fatalf("Root keys list must have one item")
	}

	if tree.Root.Keys[0] != key {
		t.Fatalf("Expected key %d, got %d\n", key, tree.Root.Keys[0])
	}
}

func TestInsertSortedKeys(t *testing.T) {
	keys := []int32{4, 1, 2}
	expectedKeys := []int32{1, 2, 4}
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)

	for _, key := range keys {
		tree.Insert(key, 0)
	}

	if len(tree.Root.Keys) != len(keys) {
		t.Fatalf("Root keys list must have %d items", len(keys))
	}

	for i, key := range expectedKeys {
		if tree.Root.Keys[i] != key {
			t.Fatalf("Expected key %d, got %d\n", key, tree.Root.Keys[i])
		}
	}
}

func TestInsertToFullRoot(t *testing.T) {
	keys := []int32{4, 1, 2}
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)

	for _, key := range keys {
		tree.Insert(key, 0)
	}

	tree.Insert(3, 0)

	if tree.Root.Size != 1 {
		t.Fatalf("Root keys list must have 1 item")
	}

	if tree.Root.Size+1 != 2 {
		t.Fatalf("Root must have 2 child got %d", len(tree.Root.Childs))
	}

	if tree.Root.Keys[0] != 2 {
		t.Fatalf("Root expected to contain 1 got %d", tree.Root.Keys[0])
	}

	leftNode := tree.CreateNode()
	leftNode.Position = tree.Root.Childs[0]

	rightNode := tree.CreateNode()
	rightNode.Position = tree.Root.Childs[1]

	tree.File.ReadNode(leftNode)
	tree.File.ReadNode(rightNode)

	if leftNode.Size != 1 {
		t.Fatalf("Left node keys list must have 1 item got %d", len(leftNode.Keys))
	}

	if rightNode.Size != 2 {
		t.Fatalf("Left node keys list must have 2 items")
	}

	if rightNode.Keys[0] != 3 && rightNode.Keys[1] != 4 {
		t.Fatalf("Incorrect keys in right node")
	}
}

func TestInsertFullLeaf(t *testing.T) {
	keys := []int32{4, 1, 2, 3, 6, 5}
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)

	for _, key := range keys {
		tree.Insert(key, 0)
	}

	if tree.Root.Size != 2 {
		t.Fatalf("Root keys list must have 1 item")
	}

	middleNode := tree.CreateNode()
	middleNode.Position = tree.Root.Childs[1]

	rightNode := tree.CreateNode()
	rightNode.Position = tree.Root.Childs[2]

	tree.File.ReadNode(middleNode)
	tree.File.ReadNode(rightNode)

	if middleNode.Size != 1 {
		t.Fatalf("Middle node keys list must have 1 item")
	}

	if rightNode.Size != 2 {
		t.Fatalf("Right node keys list must have 1 item")
	}

	if tree.Root.Keys[0] != 2 && tree.Root.Keys[1] != 4 {
		t.Fatalf("Root node contains invalid keys %v", tree.Root.Keys)
	}

	if rightNode.Keys[0] != 5 && rightNode.Keys[1] != 6 {
		t.Fatalf("Right node contains invalid keys %v", rightNode.Keys)
	}

	if middleNode.Keys[0] != 3 {
		t.Fatalf("Middle node contains invalid keys %v", middleNode.Keys)
	}
}

func TestFindRoot(t *testing.T) {
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)

	tree.Insert(5, 0)
	tree.Insert(2, 0)

	node, i := tree.Find(5)

	if node == nil || i != 1 {
		t.Fatalf("Key value 5 must exist.")
	}

	node, i = tree.Find(10)

	if node != nil && i != -1 {
		t.Fatalf("Key value 10 must not exist.")
	}
}

func TestFindInLeaf(t *testing.T) {
	keys := []int32{4, 1, 2, 3}
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)

	for _, key := range keys {
		tree.Insert(key, 0)
	}

	node, i := tree.Find(3)

	if node == nil || i != 0 {
		t.Fatalf("Key value 3 must exist.")
	}
}

func TestFind(t *testing.T) {
	keys := []int32{4, 1, 2, 3, 6, 5}
	io := TestIO{}
	tree := btree.CreateTree(&io, "test.db", 2)

	for _, key := range keys {
		tree.Insert(key, 0)
	}

	node, i := tree.Find(6)

	if node == nil || i != 1 {
		t.Fatalf("Key value 3 must exist.")
	}

	node, i = tree.Find(10)

	if node != nil && i != -1 {
		t.Fatalf("Key value 10 must not exist.")
	}
}
