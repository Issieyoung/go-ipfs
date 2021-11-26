package blockchain

import (
	"context"
	"fmt"
	"github.com/ipfs/go-bitswap"
	bsmsg "github.com/ipfs/go-bitswap/message"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-auth/standard/model"
	"github.com/ipfs/go-ipfs-backup/allocate"
	"github.com/ipfs/go-ipfs-backup/backup"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

func BlockGetRecursive(ctx context.Context, api coreiface.CoreAPI, cid cid.Cid) ([]blocks.Block, error) {
	var blockList []blocks.Block
	obj, err := api.Dag().Get(ctx, cid)
	if err != nil {
		return blockList, err
	}
	blockList = append(blockList, obj)
	links := obj.Links()
	if len(links) != 0 {
		for _, link := range links {
			b, err := BlockGetRecursive(ctx, api, link.Cid)
			if err != nil {
				return blockList, err
			}
			blockList = append(blockList, b...)
		}
	}
	return blockList, nil
}

func Allocate(node *core.IpfsNode, blockList []blocks.Block, serverList []model.CorePeer, setting allocate.Setting, uid string) error {
	ds := node.Repo.Datastore()
	bs, oneLineFlag := node.Exchange.(*bitswap.Bitswap)

	// 查询文件在本节点（或者全网络，暂未实现）已有的分布情况
	loadList, filePeerMap, err := findAllocateConditionLocal(ds, blockList, serverList, uid, setting.TargetNum)
	if err != nil {
		return err
	}

	switch setting.Strategy {
	// 策略一：文件头只在本组织存储，尽量使每个组织都拿不到完整文件
	case 0:
		// 如果是线上模式
		if oneLineFlag {
			// 分片落点算法
			err = allocate.AllocateBlocks_LOOP(loadList, serverList, setting.TargetNum, filePeerMap)
			if err != nil {
				if err == allocate.ErrBackupNotEnough {
					// 再试一次
					loadList, filePeerMap, err = findAllocateConditionLocal(ds, blockList, serverList, uid, setting.TargetNum)
					if err != nil {
						return err
					}
					err = allocate.AllocateBlocks_LOOP(loadList, serverList, setting.TargetNum, filePeerMap)
					if err != nil {
						return err
					}
				}
			}
			// 分片分发
			bs.PushTasks(loadList)
			// 先分发文件片，再记录分片备份信息
			// 记录的逻辑在func (e *Engine) MessageSent(p peer.ID, m bsmsg.BitSwapMessage)中
			// todo 或者考虑在这里先记录分发信息，在messageSent中响应分发是否成功
			return nil
		} else {
			// todo 线下模式，记录未分发，提示用户
			return fmt.Errorf("线下模式无法分发为文件")
		}

	// 策略二： 随机分配
	case 1:
		return fmt.Errorf("未知的备份策略")
	}
	return nil

}

// 查询文件在本节点记录的分布情况
func findAllocateConditionLocal(ds repo.Datastore, blockList []blocks.Block, peerList []model.CorePeer, uid string, n int) ([]bsmsg.Load, map[string]backup.StringSet, error) {
	load := make([]bsmsg.Load, len(blockList))
	filePeerMap := map[string]backup.StringSet{}
	for i, block := range blockList {
		idHash, err := backup.GetIdHash(block.Cid().String(), uid)
		l := bsmsg.Load{
			TargetPeerList: []string{},
			IdHash:         idHash,
			Block:          block,
		}

		info, err := backup.Get(ds, block.Cid().String())
		switch err {
		case datastore.ErrNotFound:
		case nil:
			for pName, _ := range info.TargetPeerList {
				// todo 去除不在peerList的节点
				filePeerMap[block.Cid().String()] = backup.StringSet{pName: {}}
				l.TargetPeerList = append(l.TargetPeerList, pName)
				if len(l.TargetPeerList) >= n {
					break
				}
			}
		default:
			return nil, nil, err
		}
		load[i] = l
	}
	return load, filePeerMap, nil
}
