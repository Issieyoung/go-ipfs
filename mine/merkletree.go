package mine

import (
	"crypto/md5"
)
import merkle "github.com/pefish/go-blkchain-merkle-tree"

type Data struct {
	s []byte
}

func (data *Data) CalculateHash() ([]byte, error) {
	h := md5.New()
	if _, err := h.Write(data.s); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (data *Data) Equals(other merkle.Content) (bool, error) {
	return byteEqual(data.s, other.(*Data).s), nil
}

func byteEqual(a, b []byte) bool {
	l := len(a)
	if len(b) != l {
		return false
	}
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func GetMerkleTree(c []merkle.Content)(*merkle.MerkleTree, error) {
	return merkle.NewMerkleTree(c)
}
