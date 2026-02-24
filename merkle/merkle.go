package merkle

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// MerkleTree implements a simple Merkle Tree compatible with OpenZeppelin's MerkleProof.
type MerkleTree struct {
	Leafs  [][]byte
	Layers [][][]byte
	Root   []byte
}

// NewMerkleTree creates a new Merkle Tree from a list of data blobs.
// The data should already be hashed (leaves).
func NewMerkleTree(leaves [][]byte) *MerkleTree {
	// Sort leaves to ensure deterministic tree (if needed) - OZ doesn't strictly require sorted leaves,
	// but often it's good practice. However, OZ MerkleProof verification *does* require sorted pairs (lesser element first).
	// Here we just take the leaves as is, usually the caller should sort them if they want a specific order (e.g. by address).
	// For standard airdrops (Uniswap style), leaves are usually sorted.

	// Copy caller-owned leaves so tree construction is deterministic
	// without mutating external state.
	sortedLeaves := make([][]byte, len(leaves))
	for i := range leaves {
		sortedLeaves[i] = append([]byte(nil), leaves[i]...)
	}

	// Let's sort the leaves to be safe and deterministic.
	sort.Slice(sortedLeaves, func(i, j int) bool {
		return bytes.Compare(sortedLeaves[i], sortedLeaves[j]) < 0
	})

	tree := &MerkleTree{
		Leafs: sortedLeaves,
	}
	tree.build()
	return tree
}

func (t *MerkleTree) build() {
	layers := [][][]byte{t.Leafs}

	// Recursive layer generation
	currentLayer := t.Leafs
	for len(currentLayer) > 1 {
		currentLayer = t.getNextLayer(currentLayer)
		layers = append(layers, currentLayer)
	}

	if len(currentLayer) > 0 {
		t.Root = currentLayer[0]
	} else {
		t.Root = make([]byte, 32)
	}
	t.Layers = layers
}

func (t *MerkleTree) getNextLayer(layer [][]byte) [][]byte {
	var nextLayer [][]byte
	for i := 0; i < len(layer); i += 2 {
		if i+1 == len(layer) {
			nextLayer = append(nextLayer, layer[i])
		} else {
			nextLayer = append(nextLayer, hashPair(layer[i], layer[i+1]))
		}
	}
	return nextLayer
}

// GetProof generates the Merkle Proof for a leaf at a given index.
func (t *MerkleTree) GetProof(leaf []byte) ([][]byte, bool) {
	// Find index
	var index int = -1
	for i, l := range t.Leafs {
		if bytes.Equal(l, leaf) {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, false
	}

	proof := [][]byte{}
	for _, layer := range t.Layers {
		if len(layer) == 1 {
			break
		}

		pairIndex := index ^ 1 // 0 -> 1, 1 -> 0, etc. (Flip last bit)
		if pairIndex < len(layer) {
			proof = append(proof, layer[pairIndex])
		}
		// Move to next layer index
		index /= 2
	}
	return proof, true
}

func hashPair(a, b []byte) []byte {
	// OpenZeppelin style: sort pairs before hashing
	if bytes.Compare(a, b) > 0 {
		a, b = b, a
	}
	combined := make([]byte, len(a)+len(b))
	copy(combined, a)
	copy(combined[len(a):], b)
	return crypto.Keccak256(combined)
}

// Helper to Hash (Index, Account, Amount) to match Solidity:
// keccak256(abi.encodePacked(index, account, amount));
func HashLeaf(index uint64, account common.Address, amount *big.Int) []byte {
	// Packed encoding:
	// Index (uint256/uint64) - solidity uses uint256 usually, need to pad?
	// abi.encodePacked simply concatenates bytes.
	// Uint256 = 32 bytes. Address = 20 bytes.

	// NOTE: In solidity "abi.encodePacked(uint256(index), account, amount)"
	// Index: 32 bytes (Big Endian)
	// Account: 20 bytes
	// Amount: 32 bytes (Big Endian)

	// Go's binary.BigEndian can be used, or math/big with specific byte length.

	return crypto.Keccak256(
		common.LeftPadBytes(new(big.Int).SetUint64(index).Bytes(), 32),
		account.Bytes(),
		common.LeftPadBytes(amount.Bytes(), 32),
	)
}
