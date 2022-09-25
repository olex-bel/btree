package main

import (
	"fmt"

	"github.com/olex-bel/btree/btree"
)

func main() {
	fileMgr := btree.BTreeFileMgr{}
	tree := btree.CreateTree(&fileMgr, "test.db", 2)

	for i := 1; i < 100; i++ {
		tree.Insert(int32(i), int32(i))
	}
	tree.Close()

	tree = btree.OpenTree(&fileMgr, "test.db")

	node, i := tree.Find(53)
	fmt.Printf("%v\n", node)
	fmt.Printf("Position: %d", i)

}
