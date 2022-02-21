package mine

import (
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"math/bits"
	"sort"
)

type SortList []cid.Cid

func (s SortList) Len() int { return len(s) }
func (s SortList) Less(i, j int) bool {
	return less(getHash(s[i]), getHash(s[j]))
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

func lessEqual(a, b []byte) int {
	for i, i1 := range a {
		if i1 != b[i] {
			if i1 > b[i] {
				return 1
			} else {
				return 2
			}
		}
	}
	return 0
}

func (s SortList) Search(b []byte) cid.Cid {
	l, r := 0, len(s)-1
	for l < r {
		n := int(uint(l+r) >> 1)
		if less(getHash(s[n]), b) {
			l = n + 1
		} else {
			r = n
		}
	}
	return s[l]
}

// CommonPrefixLen 计算两个32位byte数组的共同前导
func CommonPrefixLen(a, b []byte) int {
	if len(a) != 32 || len(b) != 32 {
		log.Error("前导零计算，位数错误")
		return 0
	}
	for i := 0; i < 32; i++ {
		c := a[i] ^ b[i]
		if c != 0 {
			return i*8 + bits.LeadingZeros8(uint8(c))
		}
	}
	return 32 * 8
}

func (s SortList) Add(c cid.Cid) Certifier {
	s = append(s, c)
	return s
}

func (s SortList) Size() int {
	return len(s)
}

func (s SortList) Construct(cl ...cid.Cid) Certifier {
	return append(s, cl...)
}
