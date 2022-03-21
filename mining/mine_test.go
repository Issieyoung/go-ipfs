package mining

/*func TestBucket(t *testing.T) {
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
}*/
