package mining

import (
	"github.com/bdengine/go-ipfs-blockchain-standard/algorithm"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"sort"
)

type SortList []*algorithm.ProofLeaf

func (s SortList) Len() int { return len(s) }
func (s SortList) Less(i, j int) bool {
	return less(getHash(s[i].C), getHash(s[j].C))
}

func less(a, b []byte) bool {
	for i, i1 := range a {
		if i1 != b[i] {
			return i1 < b[i]
		}
	}
	return false
}

func (s SortList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortList) Sort() {
	sort.Sort(s)
}

func getHash(c cid.Cid) []byte {
	decode, _ := mh.Decode(c.Hash())
	return decode.Digest
}

func (s SortList) Search(b []byte) (*algorithm.ProofLeaf, int, int) {
	l, r := 0, len(s)-1
	for l < r {
		n := int(uint(l+r) >> 1)
		if less(getHash(s[n].C), b) {
			l = n + 1
		} else {
			r = n
		}
	}
	a, _ := algorithm.CommonPrefixLen(getHash(s[l].C), b)
	if l != 0 {
		// 对比l和l-1
		c, _ := algorithm.CommonPrefixLen(getHash(s[l-1].C), b)
		if c > a {
			return s[l-1], l - 1, c
		}
	}

	return s[l], l, a
}

func (s SortList) Add(c *algorithm.ProofLeaf) SortList {
	s = append(s, c)
	return s
}

func (s SortList) Size() int {
	return len(s)
}
