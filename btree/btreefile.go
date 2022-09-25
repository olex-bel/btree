package btree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

const HEADER_SIZE = 22
const DESCRIPTOR_SIZE = 20

type BTreeFileDescriptor struct {
	TreeDegree       int16
	PageSize         int16
	RootNodePosition int64
	FirstFreeBlock   int64
}

type IOBTree interface {
	CreateFile(fullFileName string, tree *BTree, pageSize int) (*BTreeFile, error)
	OpenFile(fullFileName string) (*BTreeFile, error)
}

type BTreeFile struct {
	file       *os.File
	Descriptor *BTreeFileDescriptor
}

type BTreeFileMgr struct {
}

func getHeaderFormatId() []byte {
	return []byte{64, 62}
}

func (f *BTreeFile) readNextBytes(number int) ([]byte, error) {
	bytes := make([]byte, number)

	_, err := f.file.Read(bytes)

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (f *BTreeFile) UpdateDescriptor() error {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, f.Descriptor)
	f.file.Seek(int64(len(getHeaderFormatId())), 0)
	f.file.Write(buffer.Bytes())

	return nil
}

func (f *BTreeFile) WriteNode(node *BTreeNode) error {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, &node.Leaf)
	binary.Write(&buffer, binary.LittleEndian, &node.Size)
	binary.Write(&buffer, binary.LittleEndian, node.Keys)
	binary.Write(&buffer, binary.LittleEndian, node.Values)
	binary.Write(&buffer, binary.LittleEndian, node.Childs)

	f.file.Seek(node.Position, 0)
	f.file.Write(buffer.Bytes())
	// f.PrintNode(node)
	return nil
}

func (f *BTreeFile) ReadNode(node *BTreeNode) error {
	f.file.Seek(node.Position, 0)
	binary.Read(f.file, binary.LittleEndian, &node.Leaf)
	binary.Read(f.file, binary.LittleEndian, &node.Size)
	binary.Read(f.file, binary.LittleEndian, node.Keys)
	binary.Read(f.file, binary.LittleEndian, node.Values)
	binary.Read(f.file, binary.LittleEndian, node.Childs)

	return nil
}

func (f *BTreeFile) PrintNode(node *BTreeNode) {
	fmt.Printf("Positions: %d\n", node.Position)
	fmt.Printf("Keys: %v\n", node.Keys[0:node.Size])
	if !node.Leaf {
		fmt.Printf("Childs: %v\n", node.Childs[0:node.Size+1])
	}
	fmt.Println("-----------------------")
}

func (m *BTreeFileMgr) OpenFile(fullFileName string) (*BTreeFile, error) {
	var file, err = os.OpenFile(fullFileName, os.O_RDWR, 0644)

	if err != nil {
		return nil, err
	}

	btreeFile := BTreeFile{
		file: file,
	}
	header := BTreeFileDescriptor{}

	headerId, err := btreeFile.readNextBytes(2)

	if err != nil {
		return nil, err
	}

	if bytes.Compare(headerId, getHeaderFormatId()) != 0 {
		return nil, errors.New("Invalid file format.")
	}

	data, err := btreeFile.readNextBytes(DESCRIPTOR_SIZE)

	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.LittleEndian, &header)

	if err != nil {
		return nil, err
	}

	btreeFile.Descriptor = &header

	return &btreeFile, nil
}

func (m *BTreeFileMgr) CreateFile(fullFileName string, tree *BTree, pageSize int) (*BTreeFile, error) {
	file, err := os.Create(fullFileName)

	if err != nil {
		return nil, err
	}

	header := BTreeFileDescriptor{
		TreeDegree:     int16(tree.Degree),
		PageSize:       int16(pageSize),
		FirstFreeBlock: HEADER_SIZE,
	}

	btreeFile := BTreeFile{
		file:       file,
		Descriptor: &header,
	}

	var buffer bytes.Buffer
	err = binary.Write(&buffer, binary.LittleEndian, &header)

	if err != nil {
		return nil, err
	}

	fmt.Print(len(buffer.Bytes()))
	file.Write(getHeaderFormatId())
	file.Write(buffer.Bytes())

	return &btreeFile, nil
}
