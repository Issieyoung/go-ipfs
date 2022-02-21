package mine

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

func (k Bucket) Add(c cid.Cid) Certifier {
	i := getHash(c)[0]
	if k[i] == nil {
		k[i] = SortList{}
	}
	k[i] = k[i].Add(c)
	k[i].Sort()
	return k
}

func (k Bucket) Search(b []byte) (j cid.Cid) {
	j = k[b[0]].Search(b)
	return
}

func (k Bucket) Construct(cl ...cid.Cid) Certifier {
	for _, c := range cl {
		one := getHash(c)[0]
		if k[one] == nil {
			k[one] = SortList{}
		}
		a := k[one].Construct(c)
		k[one] = a
	}
	k.Sort()
	return k
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
