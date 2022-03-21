package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	selector "github.com/bdengine/go-ipfs-blockchain-selector"
	"github.com/bdengine/go-ipfs-blockchain-standard/algorithm"
	"github.com/bdengine/go-ipfs-blockchain-standard/dto"
	"github.com/bdengine/go-ipfs-blockchain-standard/standardConst"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/mining"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"strings"
	"time"
)

func initCTree(ctx context.Context, node *core.IpfsNode) error {
	// 初始化红黑树
	s0 := time.Now()
	challenge, err := selector.GetStoreChallenge()
	if err != nil {
		return err
	}
	//challengeByte, _ := base64.StdEncoding.DecodeString(challenge)
	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		log.Errorf("failed to access CoreAPI: %v", err)
	}

	mining.NewMineServer(challenge, node.Identity, api, node.Repo.Datastore())
	//  红黑树构造
	mining.PullNewFile(ctx, node.Repo.Datastore())
	s1 := time.Now()
	go func() {
		// 定时查询fileList,拉取文件
		tick := time.Tick(1 * time.Hour)
		for {
			<-tick
			mining.PullNewFile(ctx, node.Repo.Datastore())
		}
	}()

	fmt.Printf("总文件片：%v，生成CTree树形结构时间：%v\n", mining.GetCTreeSize(), s1.Sub(s0))
	return nil
}

const (
	quick     = 30
	slow      = 90
	blockTime = 6
)

// miningStage 定时查询当前阶段
func miningStage(ctx context.Context, node *core.IpfsNode) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := checkStage(node.Repo.Datastore())
			//ticker.Reset(time.Duration(leftTime) * time.Second)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func checkStage(ds datastore.Datastore) (int64, error) {
	// 查询当前阶段
	stage, s, l, err := selector.GetChallengeStage()
	if err != nil {
		return quick, err
	}
	leftTime := l * blockTime

	// 同一个阶段不重复处理
	if mining.MServer.ChallengeStage == int(stage) && mining.MServer.S == s {
		return leftTime, nil
	}
	storeChallenge, err := selector.GetStoreChallenge()
	if err != nil {
		return quick, err
	}
	storeChallengeBytes, _ := base64.StdEncoding.DecodeString(storeChallenge)

	switch stage {
	case standardConst.CycleStageGen, standardConst.CycleStageUpdate:
		if mining.MServer.Tree.StoreChallenge != nil && string(storeChallengeBytes) == string(mining.MServer.Tree.StoreChallenge) {
			return leftTime, err
		}
		mining.MServer.Tree.StoreChallenge = storeChallengeBytes

		// 已发送
		if mining.IsSend() {
			return leftTime, nil
		}
		log.Debugf("stage:")
		// 更新storeChallenge
		mining.UpdateCTreeChallenge(s)
		// 生成store Challenge
		sList, size, ok := mining.GetMerkleTree()
		// 是前一个发送的
		if ok {
			return leftTime, nil
		}
		// 持久化存储
		err = mining.StoreMerkleTree(ds, sList, s, size)
		if err != nil {
			return quick, err
		}
		// 发送 存储证明存根
		storeProofDTO := dto.StoreProofDTO{
			ProofRoot:      sList[len(sList)-1].Hash.String(),
			StoreChallenge: s,
			PeerId:         mining.MServer.Pid.String(),
			PeerAddress:    "",
		}
		err = selector.UpdateOrGen(storeProofDTO)
		if err != nil {
			return quick, err
		}
		// 更新
		mining.UpdatePre(sList[len(sList)-1].Hash.String(), ds)
	case standardConst.CycleStageCollect:
		// 获取最近发送，存储的对应storeChallenge的存储证明
		pre := mining.MServer.Pre
		if pre == nil {
			return slow, fmt.Errorf("未发送存储证明")
		}
		// 检查challenge和storeChallenge是否能对应
		if !bytes.Equal(pre.StoreChallenge, storeChallengeBytes) {
			return slow, fmt.Errorf("存储证明未更新")
		}

		merkleTree, err := mining.FindMerkleTree(ds, base64.StdEncoding.EncodeToString(pre.StoreChallenge), pre.MerkleRoot)
		if err != nil {
			return slow, err
		}
		// 查找最佳cid
		sByte, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return quick, err
		}
		best, _, leadingZero := merkleTree.SortList[:pre.FileNum].Search(sByte)
		// 发送最佳cid
		miningDTO := dto.MiningDTO{
			Cid:            best.C.String(),
			Address:        "",
			Pid:            mining.MServer.Pid.String(),
			SPVProof:       nil,
			LeadingZero:    leadingZero,
			Challenge:      s,
			StoreChallenge: merkleTree.StoreChallenge,
			ProofRoot:      merkleTree.RootStr,
		}
		err = selector.Mining(miningDTO)
		if err != nil {
			return quick, err
		}
	case standardConst.CycleStageAnnounce:
		// 检查是本节点是否是获奖者
		if s != mining.MServer.Pid.String() {
			mining.MServer.ChallengeStage = int(stage)
			mining.MServer.S = s
			return slow, nil
		}
		// 获取获奖对应的cid，存储证明
		pre := mining.MServer.Pre
		if pre == nil {
			return slow, fmt.Errorf("未发送存储证明")
		}
		merkleTree, err := mining.FindMerkleTree(ds, base64.StdEncoding.EncodeToString(pre.StoreChallenge), pre.MerkleRoot)
		if err != nil {
			return slow, err
		}
		// 查找最佳cid
		c, err := selector.GetChallenge()
		if err != nil {
			return 0, err
		}
		sByte, err := base64.StdEncoding.DecodeString(c)
		if err != nil {
			return quick, err
		}
		best, i, _ := merkleTree.SortList[:pre.FileNum].Search(sByte)
		svProof := algorithm.GetMerkleBlock(merkleTree.SortList, i)
		// 生成简单证明并发送
		proofDTO := dto.SVProofDTO{
			Cid:            best.C.String(),
			Pid:            mining.MServer.Pid.String(),
			SvProof:        svProof,
			StoreChallenge: merkleTree.StoreChallenge,
			ProofRoot:      merkleTree.RootStr,
			ChallengeHash:  best.ChallengeHash[:],
			ProofLeaf:      best.Hash,
		}
		err = selector.Prove(proofDTO)
		if err != nil {
			return quick, err
		}
	case standardConst.CycleStageWait:
		// do nothing
		leftTime = slow
	}
	// 更新
	mining.MServer.ChallengeStage = int(stage)
	mining.MServer.S = s
	return leftTime, nil
}

// updateAddress 更新链上节点地址
func updateAddressAuto(node *core.IpfsNode) error {
	var addressList []string
	addrss, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(node.PeerHost))
	if err != nil {
		return err
	}
	var s string
	for _, addr := range addrss {
		s = addr.String()
		if !strings.Contains(s, "127.0.0.1") && !strings.Contains(s, "/::1/") {
			addressList = append(addressList, s)
		}
	}
	err = selector.UpdateAddress(addressList)
	if err != nil {
		log.Warn("更新节点地址失败", err)
		return err
	}
	// TODO 定时检查ip是否发生变化
	return nil
}

// updateAddress 更新链上节点地址
func updateAddress(addressList []string) error {
	err := selector.UpdateAddress(addressList)
	if err != nil {
		log.Warn("更新节点地主之失败", err)
		return err
	}
	// TODO 定时检查ip是否发生变化
	return nil
}
