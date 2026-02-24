package merkle

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func clone2DBytes(in [][]byte) [][]byte {
	out := make([][]byte, len(in))
	for i := range in {
		out[i] = append([]byte(nil), in[i]...)
	}
	return out
}

func makeHashedLeaves(vals ...string) [][]byte {
	leaves := make([][]byte, len(vals))
	for i := range vals {
		leaves[i] = crypto.Keccak256([]byte(vals[i]))
	}
	return leaves
}

func verifyProof(leaf []byte, proof [][]byte, root []byte) bool {
	computed := append([]byte(nil), leaf...)
	for _, sibling := range proof {
		computed = hashPair(computed, sibling)
	}
	return bytes.Equal(computed, root)
}

func expectedRootFromSortedLeaves(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return make([]byte, 32)
	}

	current := clone2DBytes(leaves)
	sort.Slice(current, func(i, j int) bool {
		return bytes.Compare(current[i], current[j]) < 0
	})

	for len(current) > 1 {
		next := make([][]byte, 0, (len(current)+1)/2)
		for i := 0; i < len(current); i += 2 {
			if i+1 == len(current) {
				next = append(next, current[i])
				continue
			}
			next = append(next, hashPair(current[i], current[i+1]))
		}
		current = next
	}

	return current[0]
}

func TestNewMerkleTree_DoesNotMutateInputLeaves(t *testing.T) {
	// Keccak("alice") > Keccak("bob"), so internal sorting will reorder these two.
	leaves := [][]byte{
		crypto.Keccak256([]byte("alice")),
		crypto.Keccak256([]byte("bob")),
	}
	original := clone2DBytes(leaves)

	_ = NewMerkleTree(leaves)

	require.Equal(t, original, leaves,
		"NewMerkleTree should not mutate caller-owned leaves in place")
}

func TestHashLeaf_UsesFullUint64IndexWithoutOverflow(t *testing.T) {
	index := uint64(math.MaxUint64)
	account := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	amount := big.NewInt(12345)

	got := HashLeaf(index, account, amount)

	want := crypto.Keccak256(
		common.LeftPadBytes(new(big.Int).SetUint64(index).Bytes(), 32),
		account.Bytes(),
		common.LeftPadBytes(amount.Bytes(), 32),
	)

	require.Equal(t, fmt.Sprintf("%x", want), fmt.Sprintf("%x", got),
		"HashLeaf should hash full uint64 range without int64 overflow")
}

func TestNewMerkleTree_RootCases(t *testing.T) {
	tests := []struct {
		name   string
		leaves [][]byte
	}{
		{
			name:   "empty leaves",
			leaves: nil,
		},
		{
			name:   "single leaf",
			leaves: makeHashedLeaves("alice"),
		},
		{
			name:   "even leaves",
			leaves: makeHashedLeaves("alice", "bob", "carol", "dave"),
		},
		{
			name:   "odd leaves",
			leaves: makeHashedLeaves("alice", "bob", "carol", "dave", "eve"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			working := clone2DBytes(tt.leaves)
			tree := NewMerkleTree(working)

			require.NotNil(t, tree)
			require.NotNil(t, tree.Root)
			require.NotEmpty(t, tree.Layers)
			require.Equal(t, fmt.Sprintf("%x", expectedRootFromSortedLeaves(tt.leaves)), fmt.Sprintf("%x", tree.Root))
		})
	}
}

func TestGetProof_LeafNotFound(t *testing.T) {
	tree := NewMerkleTree(clone2DBytes(makeHashedLeaves("alice", "bob", "carol")))

	proof, ok := tree.GetProof(crypto.Keccak256([]byte("missing")))
	require.False(t, ok)
	require.Nil(t, proof)
}

func TestGetProof_SingleLeafReturnsEmptyProof(t *testing.T) {
	leaf := crypto.Keccak256([]byte("alice"))
	tree := NewMerkleTree([][]byte{append([]byte(nil), leaf...)})

	proof, ok := tree.GetProof(leaf)
	require.True(t, ok)
	require.Empty(t, proof)
	require.True(t, verifyProof(leaf, proof, tree.Root))
}

func TestGetProof_ValidForAllLeavesAcrossSizes(t *testing.T) {
	tests := []struct {
		name   string
		leaves [][]byte
	}{
		{
			name:   "2 leaves",
			leaves: makeHashedLeaves("alice", "bob"),
		},
		{
			name:   "3 leaves",
			leaves: makeHashedLeaves("alice", "bob", "carol"),
		},
		{
			name:   "4 leaves",
			leaves: makeHashedLeaves("alice", "bob", "carol", "dave"),
		},
		{
			name:   "5 leaves",
			leaves: makeHashedLeaves("alice", "bob", "carol", "dave", "eve"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewMerkleTree(clone2DBytes(tt.leaves))

			for i := range tree.Leafs {
				leaf := tree.Leafs[i]
				proof, ok := tree.GetProof(leaf)
				require.True(t, ok)
				require.True(t, verifyProof(leaf, proof, tree.Root), "leaf at sorted index %d should verify", i)
			}
		})
	}
}

func TestNewMerkleTree_DeterministicRootAcrossInputOrder(t *testing.T) {
	base := makeHashedLeaves("alice", "bob", "carol", "dave", "eve")
	reversed := clone2DBytes(base)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	treeA := NewMerkleTree(clone2DBytes(base))
	treeB := NewMerkleTree(clone2DBytes(reversed))

	require.Equal(t, fmt.Sprintf("%x", treeA.Root), fmt.Sprintf("%x", treeB.Root))
}

func TestGetProof_DuplicateLeaves(t *testing.T) {
	dup := crypto.Keccak256([]byte("same-leaf"))
	other := crypto.Keccak256([]byte("other-leaf"))
	tree := NewMerkleTree([][]byte{
		append([]byte(nil), dup...),
		append([]byte(nil), other...),
		append([]byte(nil), dup...),
	})

	proof, ok := tree.GetProof(dup)
	require.True(t, ok)
	require.True(t, verifyProof(dup, proof, tree.Root))
}

func TestGetProof_DoesNotContainLeafItself_SimpleCase(t *testing.T) {
	leaf1 := crypto.Keccak256([]byte("leaf1"))
	leaf2 := crypto.Keccak256([]byte("leaf2"))
	leaf3 := crypto.Keccak256([]byte("leaf3"))

	tree := NewMerkleTree([][]byte{
		append([]byte(nil), leaf1...),
		append([]byte(nil), leaf2...),
		append([]byte(nil), leaf3...),
	})

	proof, ok := tree.GetProof(leaf1)
	require.True(t, ok)
	require.NotEmpty(t, proof)

	for i := range proof {
		require.NotEqual(t, fmt.Sprintf("%x", leaf1), fmt.Sprintf("%x", proof[i]), "proof element %d should not be the leaf itself", i)
	}
	require.True(t, verifyProof(leaf1, proof, tree.Root))
}

func TestHashLeaf_TableCases(t *testing.T) {
	account := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	amount := big.NewInt(12345)

	tests := []struct {
		name  string
		index uint64
	}{
		{name: "zero", index: 0},
		{name: "small", index: 7},
		{name: "boundary max int64", index: math.MaxInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashLeaf(tt.index, account, amount)
			want := crypto.Keccak256(
				common.LeftPadBytes(new(big.Int).SetUint64(tt.index).Bytes(), 32),
				account.Bytes(),
				common.LeftPadBytes(amount.Bytes(), 32),
			)
			require.Equal(t, fmt.Sprintf("%x", want), fmt.Sprintf("%x", got))
		})
	}
}
