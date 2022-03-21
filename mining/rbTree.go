package mining

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/bdengine/go-ipfs-blockchain-standard/algorithm"
	"github.com/ipfs/go-cid"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"sync"
)

const (
	Red   = true
	Black = false
)

type node struct {
	key   cid.Cid
	value *algorithm.Hash
	color bool
	left  *node
	right *node
}

type redBlackTree struct {
	size           int
	root           *node
	StoreChallenge []byte
	lock           sync.RWMutex
	api            coreiface.CoreAPI
}

func newNode(key cid.Cid, val *algorithm.Hash) *node {
	// 默认添加红节点
	// val
	return &node{key, val, Red, nil, nil}
}

func (nd *node) isRed() bool {
	if nd == nil {
		return Black
	}
	return nd.color
}

func (tree *redBlackTree) getSize() int {
	tree.lock.RLock()
	defer tree.lock.RUnlock()
	return tree.size
}

// Add 向红黑树中添加元素
func (tree *redBlackTree) add(key cid.Cid) bool {
	tree.lock.Lock()
	defer tree.lock.Unlock()

	get, _ := tree.api.Dag().Get(context.Background(), key)
	s := sha256.Sum256(append(tree.StoreChallenge, get.RawData()...))
	val := algorithm.GetHashPointer(s[:])

	isAdd, nd := tree.root.add(key, val)
	tree.size += isAdd
	tree.root = nd
	tree.root.color = Black //根节点为黑色节点
	return isAdd == 1
}

// Add 向红黑树中添加元素
func (tree *redBlackTree) addMany(ctx context.Context, kl []cid.Cid) {
	tree.lock.Lock()
	defer tree.lock.Unlock()

	for _, key := range kl {
		get, _ := tree.api.Dag().Get(ctx, key)
		s := sha256.Sum256(append(tree.StoreChallenge, get.RawData()...))
		val := algorithm.GetHashPointer(s[:])

		isAdd, nd := tree.root.add(key, val)
		tree.size += isAdd
		tree.root = nd
		tree.root.color = Black //根节点为黑色节点
	}
}

const (
	bigger  = 1
	smaller = -1
	isEqual = 0
)

func equal(ca, cb cid.Cid) int {
	a := getHash(ca)
	b := getHash(cb)

	for i, i1 := range a {
		if i1 > b[i] {
			return bigger
		}
		if i1 < b[i] {
			return smaller
		}
	}
	return isEqual
}

// 递归写法:向树的root节点中插入key,val
// 返回1,代表加了节点
// 返回0,代表没有添加新节点,只更新key对应的value值
func (nd *node) add(key cid.Cid, val *algorithm.Hash) (int, *node) {
	if nd == nil { // 默认插入红色节点
		return 1, newNode(key, val)
	}
	flag := equal(key, nd.key)
	isAdd := 0

	if flag == smaller {
		isAdd, nd.left = nd.left.add(key, val)
	} else if flag == bigger {
		isAdd, nd.right = nd.right.add(key, val)
	} else { // nd.key == key
		// 对value值更新,节点数量不增加,isAdd = 0
		nd.value = val
	}

	// 维护红黑树
	nd = nd.updateRedBlackTree(isAdd)
	return isAdd, nd
}

// 红黑树维护
func (nd *node) updateRedBlackTree(isChange int) *node {
	// 0说明无新节点,不必维护
	if isChange == 0 {
		return nd
	}

	// 维护
	// 判断是否为情形2，是需要左旋转
	if nd.right.isRed() == Red && nd.left.isRed() != Red {
		nd = nd.leftRotate()
	}

	// 判断是否为情形3，是需要右旋转
	if nd.left.isRed() == Red && nd.left.left.isRed() == Red {
		nd = nd.rightRotate()
	}

	// 判断是否为情形4，是需要颜色翻转
	if nd.left.isRed() == Red && nd.right.isRed() == Red {
		nd.flipColors()
	}
	return nd
}

//    nd                      x
//  /   \     左旋转         /  \
// T1   x   --------->   node   T3
//     / \              /   \
//    T2 T3            T1   T2
func (nd *node) leftRotate() *node {
	// 左旋转
	retNode := nd.right
	nd.right = retNode.left

	retNode.left = nd
	retNode.color = nd.color
	nd.color = Red

	return retNode
}

//      nd                    x
//    /   \     右旋转       /  \
//   x    T2   ------->   y   node
//  / \                       /  \
// y  T1                     T1  T2
func (nd *node) rightRotate() *node {
	//右旋转
	retNode := nd.left
	nd.left = retNode.right
	retNode.right = nd

	retNode.color = nd.color
	nd.color = Red

	return retNode
}

// 颜色翻转
func (nd *node) flipColors() {
	nd.color = Red
	nd.left.color = Black
	nd.right.color = Black
}

// 前序遍历打印出key,val,color
func (tree *redBlackTree) printPreOrder() {
	tree.lock.RLock()
	defer tree.lock.RUnlock()
	resp := [][]interface{}{}
	tree.root.printPreOrder(&resp)
	fmt.Println(resp)
}

func Search(root *node, c cid.Cid) *node {
	if root == nil {
		return nil
	}
	flag := equal(root.key, c)
	if flag == isEqual {
		return root
	} else if flag == bigger {
		return Search(root.left, c)
	} else {
		return Search(root.right, c)
	}
}

func (nd *node) printPreOrder(resp *[][]interface{}) {
	if nd == nil {
		return
	}
	*resp = append(*resp, []interface{}{nd.key, nd.value, nd.color})
	nd.left.printPreOrder(resp)
	nd.right.printPreOrder(resp)
}

func (nd *node) middleOrder(resp *[]cid.Cid) {
	if nd == nil {
		return
	}
	nd.left.middleOrder(resp)
	*resp = append(*resp, nd.key)
	nd.right.middleOrder(resp)
}

func (tree *redBlackTree) getSortList(pidByte []byte) SortList {
	tree.lock.RLock()
	defer tree.lock.RUnlock()
	var res []*algorithm.ProofLeaf
	tree.root.getSortList(&res, pidByte)
	return res
}

func (tree *redBlackTree) getMerkleTree(pidByte []byte) (SortList, int) {
	tree.lock.RLock()
	defer tree.lock.RUnlock()
	var res []*algorithm.ProofLeaf
	tree.root.getSortList(&res, pidByte)

	return algorithm.BuildMerkleTreeStore(res, pidByte), tree.size
}

func ByteEqual(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

func (nd *node) getSortList(resp *[]*algorithm.ProofLeaf, pidByte []byte) {
	if nd == nil {
		return
	}
	nd.left.getSortList(resp, pidByte)
	*resp = append(*resp, algorithm.NewProofLeaf(nd.key, nd.value, pidByte))
	nd.right.getSortList(resp, pidByte)
}

// TODO cid，challenge计算证明
func getChallengeHash(challenge []byte, c cid.Cid, api coreiface.CoreAPI) *algorithm.Hash {
	get, _ := api.Dag().Get(context.Background(), c)
	h := sha256.New()
	h.Write(challenge)
	h.Write(get.RawData())
	sum := h.Sum(nil)
	return algorithm.GetHashPointer(sum)
}

func (nd *node) updateStoreChallenge(challenge []byte, api coreiface.CoreAPI) {
	if nd == nil {
		return
	}
	nd.left.updateStoreChallenge(challenge, api)

	nd.value = getChallengeHash(challenge, nd.key, api)

	nd.right.updateStoreChallenge(challenge, api)
}

func (tree *redBlackTree) updateStoreChallenge(challenge string) {
	tree.lock.Lock()
	defer tree.lock.Unlock()
	b, _ := base64.StdEncoding.DecodeString(challenge)
	if ByteEqual(b, tree.StoreChallenge) {
		return
	}
	tree.StoreChallenge = b
	tree.root.updateStoreChallenge(b, tree.api)
}
