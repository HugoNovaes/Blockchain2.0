package blockchain

import (
	"golang.org/x/crypto/sha3"
)

type MerkleNode struct {
	Hash   HashBlock
	Parent *MerkleNode
	Left   *MerkleNode
	Right  *MerkleNode
}

type MerkleTree struct {
	Root *MerkleNode `json:"merkle"`
}

func (m *MerkleNode) buildHash() (result HashBlock, err error) {
	hash := sha3.New256()
	hash.Write(m.Left.Hash[:])
	hash.Write(m.Right.Hash[:])
	copy(result[:], hash.Sum(nil))
	return result, err
}

func (m *MerkleTree) BuildMarkleTree(transactions []Transaction) (result *MerkleNode, err error) {

	if len(transactions) == 0 {
		return nil, ErrNoTransactions
	}

	leaves := make([]Transaction, 0)
	leaves = append(leaves, transactions...)

	if len(leaves)%2 == 1 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}

	joinLeaves := make([]*MerkleNode, 0)

	for i := 0; i < len(leaves)/2; i++ {
		item := &MerkleNode{}
		item.Left = &MerkleNode{Hash: leaves[i*2].Hash}
		item.Right = &MerkleNode{Hash: leaves[i*2+1].Hash}
		item.Hash, err = item.buildHash()
		joinLeaves = append(joinLeaves, item)
	}

	result = buildNodes(joinLeaves)

	return result, err
}

func buildNodes(nodes []*MerkleNode) (result *MerkleNode) {
	lenNoded := len(nodes)
	if lenNoded == 0 {
		return nil
	}

	if lenNoded == 1 {
		return nodes[0]
	}

	if lenNoded%2 == 1 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	items := make([]*MerkleNode, 0)

	for i := 0; i < len(nodes)/2; i++ {
		item := &MerkleNode{}
		item.Left = nodes[i*2]
		item.Right = nodes[i*2+1]
		item.Hash, _ = item.buildHash()
		item.Left.Parent = item
		item.Right.Parent = item

		items = append(items, item)
	}

	result = buildNodes(items)

	return result
}
