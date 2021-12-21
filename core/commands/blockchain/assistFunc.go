package blockchain

import (
	"context"
	"fmt"
	"github.com/ipfs/go-bitswap"
	bsmsg "github.com/ipfs/go-bitswap/message"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-auth/selector"
	"github.com/ipfs/go-ipfs-auth/standard/model"
	"github.com/ipfs/go-ipfs-backup/allocate"
	"github.com/ipfs/go-ipfs-backup/backup"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	"sync"
	"time"
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

func CidGet(ctx context.Context, api coreiface.CoreAPI, cid cid.Cid, reFlag bool) ([]string, error) {
	var cidList []string
	cidList = append(cidList, cid.String())
	if !reFlag {
		return cidList, nil
	}
	obj, err := api.Dag().Get(ctx, cid)
	if err != nil {
		return cidList, err
	}
	links := obj.Links()
	if len(links) != 0 {
		for _, link := range links {
			b, err := CidGet(ctx, api, link.Cid, reFlag)
			if err != nil {
				return cidList, err
			}
			cidList = append(cidList, b...)
		}
	}
	return cidList, nil
}

func Allocate(node *core.IpfsNode, blockList []blocks.Block, serverList []model.CorePeer, setting allocate.Setting, uid string, size uint64) error {
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
				} else {
					return err
				}
			}
			// 记录备份信息
			backup.AddFileBackupInfo(ds, loadList, uid, size)
			// 分片分发
			bs.PushTasks(loadList)
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

func Connect(ctx context.Context, addrs []string, node *core.IpfsNode, api coreiface.CoreAPI) error {
	pis, err := parseAddresses(ctx, addrs, node.DNSResolver)
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()
	for _, pi := range pis {
		select {
		case <-ctx.Done():
			return fmt.Errorf("连接节点超时")
		default:
			err = api.Swarm().Connect(ctx, pi)
			if err == nil {
				return nil
			}
		}
	}
	return err
}

// parseAddresses is a function that takes in a slice of string peer addresses
// (multiaddr + peerid) and returns a slice of properly constructed peers
func parseAddresses(ctx context.Context, addrs []string, rslv *madns.Resolver) ([]peer.AddrInfo, error) {
	// resolve addresses
	maddrs, err := resolveAddresses(ctx, addrs, rslv)
	if err != nil {
		return nil, err
	}

	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

const (
	dnsResolveTimeout = 10 * time.Second
)

// resolveAddresses resolves addresses parallelly
func resolveAddresses(ctx context.Context, addrs []string, rslv *madns.Resolver) ([]ma.Multiaddr, error) {
	ctx, cancel := context.WithTimeout(ctx, dnsResolveTimeout)
	defer cancel()

	var maddrs []ma.Multiaddr
	var wg sync.WaitGroup
	resolveErrC := make(chan error, len(addrs))

	maddrC := make(chan ma.Multiaddr)

	for _, addr := range addrs {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}

		// check whether address ends in `ipfs/Qm...`
		if _, last := ma.SplitLast(maddr); last.Protocol().Code == ma.P_IPFS {
			maddrs = append(maddrs, maddr)
			continue
		}
		wg.Add(1)
		go func(maddr ma.Multiaddr) {
			defer wg.Done()
			raddrs, err := rslv.Resolve(ctx, maddr)
			if err != nil {
				resolveErrC <- err
				return
			}
			// filter out addresses that still doesn't end in `ipfs/Qm...`
			found := 0
			for _, raddr := range raddrs {
				if _, last := ma.SplitLast(raddr); last != nil && last.Protocol().Code == ma.P_IPFS {
					maddrC <- raddr
					found++
				}
			}
			if found == 0 {
				resolveErrC <- fmt.Errorf("found no ipfs peers at %s", maddr)
			}
		}(maddr)
	}
	go func() {
		wg.Wait()
		close(maddrC)
	}()

	for maddr := range maddrC {
		maddrs = append(maddrs, maddr)
	}

	select {
	case err := <-resolveErrC:
		return nil, err
	default:
	}

	return maddrs, nil
}

// getReliablePeer 获取可信节点列表
func getReliablePeer(ctx context.Context, node *core.IpfsNode, api coreiface.CoreAPI, num int) ([]model.CorePeer, error) {
	pl, err := selector.GetPeerList(0)
	if err != nil {
		return pl, err
	}
	var peerList []model.CorePeer
	for _, p := range pl {
		if p.PeerId == "" {
			continue
		}
		err := Connect(ctx, p.Addresses, node, api)
		if err == nil {
			peerList = append(peerList, p)
			if len(peerList) >= num {
				return peerList, nil
			}
		}
	}
	return peerList, nil
}
