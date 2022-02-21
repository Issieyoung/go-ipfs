package mine

import (
	"fmt"
	"github.com/ipfs/go-cid"
)

const (
	Red   = true
	Black = false
)

type node struct {
	key   cid.Cid
	value int
	color bool
	left  *node
	right *node
}

type redBlackTree struct {
	size int
	root *node
}

func newNode(key cid.Cid, val int) *node {
	// 默认添加红节点
	return &node{key, val, Red, nil, nil}
}

func NewRedBlackTree() *redBlackTree {
	return new(redBlackTree)
}

func (nd *node) isRed() bool {
	if nd == nil {
		return Black
	}
	return nd.color
}

func (tree *redBlackTree) GetSize() int {
	return tree.size
}

// 向红黑树中添加元素
func (tree *redBlackTree) Add(key cid.Cid, val int) {
	isAdd, nd := tree.root.add(key, val)
	tree.size += isAdd
	tree.root = nd
	tree.root.color = Black //根节点为黑色节点
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
func (nd *node) add(key cid.Cid, val int) (int, *node) {
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
func (tree *redBlackTree) PrintPreOrder() {
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

var max *node
var maxLen int = -1

func (nd *node) Search(c cid.Cid) (*node, int) {
	if nd == nil {
		return nil, -1
	}
	lroot := CommonPrefixLen(getHash(nd.key), getHash(c))
	if lroot > maxLen {
		maxLen = lroot
		max = nd
	}

	l, ll := nd.left.Search(c)
	if ll > lroot {
		maxLen = ll
		max = l
	}
	r, lr := nd.right.Search(c)
	if lr > lroot {
		maxLen = lr
		max = r
	}
	return max, maxLen
}
