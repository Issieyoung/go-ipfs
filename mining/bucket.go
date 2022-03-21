package mining

import (
	"github.com/ipfs/go-cid"
)

type Certifier interface {
	Add(c cid.Cid) Certifier
	Search(b []byte) cid.Cid
	Sort()
	Construct(cl ...cid.Cid) Certifier
	Size() int
}

type Bucket [256]Certifier

func NewKBucket() Bucket {
	var k [256]Certifier
	return k
}

func (k Bucket) Search(b []byte) (j cid.Cid) {
	j = k[b[0]].Search(b)
	return
}

func (k Bucket) Size() int {
	size := 0
	for _, s := range k {
		size += s.Size()
	}
	return size
}

func (k Bucket) Sort() {
	for _, searcher := range k {
		if searcher != nil && searcher.Size() > 1 {
			searcher.Sort()
		}
	}
	return
}
