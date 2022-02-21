package mine

import (
	"fmt"
	block "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"sort"
	"testing"
	"time"
)

func TestCidSortList(t *testing.T) {
	times := 50000
	ch := SortList{}
	start0 := time.Now()
	for i := 0; i < times; i++ {
		c := block.NewBlock([]byte(fmt.Sprintf("test%v", i))).Cid()
		ch = append(ch, c)
	}
	start1 := time.Now()
	sort.Sort(ch)
	start2 := time.Now()

	sortTime := time.Now()
	for _, c := range ch {
		i1 := ch.Search(getHash(c))
		if i1.String() != c.String() {
			t.Fatal("没找到对应cid")
		}
	}
	start3 := time.Now()
	for i := times; i < times+10; i++ {
		c := block.NewBlock([]byte(fmt.Sprintf("test%v", i))).Cid()
		hash := getHash(c)

		index := ch.Search(hash)
		r := getHash(index)
		rLen := CommonPrefixLen(hash, r)
		lLen := -1
		/*if index > 0 {
			l := getHash(ch[index-1])
			lLen = CommonPrefixLen(hash, l)
		}*/
		fmt.Println(rLen, lLen)
		if rLen == lLen {
			fmt.Println("左右前导零一致")
		}
	}
	start4 := time.Now()
	fmt.Printf("生成列表时间：%v  \n 排序时间：%v  \n 检索时间：%v  \n ", start1.Sub(start0), start2.Sub(start1), start4.Sub(start3))
	fmt.Printf("第二次排序时间：%v  \n", sortTime.Sub(start2))
}

func TestRBTree(t *testing.T) {
	times := 50000
	rbt := NewRedBlackTree()
	start0 := time.Now()
	cList := make([]cid.Cid, times)
	for i := 0; i < times; i++ {
		c := block.NewBlock([]byte(fmt.Sprintf("test%v", i))).Cid()
		cList[i] = c
	}
	start1 := time.Now()

	for i, c := range cList {
		rbt.Add(c, i)
	}

	start2 := time.Now()
	for _, c := range cList {
		search := Search(rbt.root, c)
		if equal(search.key, c) != isEqual {
			t.Fatal("查找到错误的节点")
		}

	}

	start3 := time.Now()
	addnum := 100
	addList := make([]cid.Cid, addnum)
	for i := times; i < times+addnum; i++ {
		c := block.NewBlock([]byte(fmt.Sprintf("test%v", i))).Cid()
		addList[i-times] = c
	}
	start4 := time.Now()
	/*for i, c := range addList {
		rbt.Add(c,i)
	}*/
	start5 := time.Now()
	for _, c := range addList {
		max = nil
		maxLen = -1
		_, l := rbt.root.Search(c)
		fmt.Println(l)
	}
	start6 := time.Now()
	fmt.Printf("生成时间：%v  \n构建时间：%v  \n  检索时间：%v  \n ", start1.Sub(start0), start2.Sub(start1), start3.Sub(start2))
	fmt.Printf("生成时间：%v  \n构建时间：%v  \n  检索时间：%v  \n ", start4.Sub(start3), start5.Sub(start4), start6.Sub(start5))

}

func TestBucket(t *testing.T) {
	times := 500000
	bucket := NewKBucket()
	start0 := time.Now()
	cl := make([]cid.Cid, times)
	for i := 0; i < times; i++ {
		c := block.NewBlock([]byte(fmt.Sprintf("test%v", i))).Cid()
		cl[i] = c
	}

	start1 := time.Now()
	bucket.Construct(cl...)
	start2 := time.Now()

	addNum := 100
	addList := make([]cid.Cid, addNum)
	for i := times; i < times+addNum; i++ {
		c := block.NewBlock([]byte(fmt.Sprintf("test%v", i))).Cid()
		addList[i-times] = c
	}

	start3 := time.Now()
	for _, c := range addList {
		bucket.Add(c)
	}
	start4 := time.Now()

	fmt.Printf("生成列表时间： %v \n 生成结构时间：%v  \n", start1.Sub(start0), start2.Sub(start1))
	fmt.Printf("生成列表时间： %v \n 生成结构时间：%v  \n", start3.Sub(start2), start4.Sub(start3))
}
