package mining

import (
	"context"
	selector "github.com/bdengine/go-ipfs-blockchain-selector"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

const fileExistKey = "fileExist/"

// log is the command logger
var log = logging.Logger("cmd/ipfs")

// todo 考虑放入数据库而非内存？
var storeCidMSet = cid.Set{}

func GetValidFileList() ([]cid.Cid, error) {
	fileList, err := selector.GetFileList(0)
	if err != nil {
		return nil, err
	}
	cidList := make([]cid.Cid, len(fileList))
	for i, s := range fileList {
		parse, err := cid.Parse(s)
		if err != nil {
			return nil, err
		}
		cidList[i] = parse
	}
	return cidList, nil
}

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

// 递归的构建dag结构
func dagGetRecursive(ctx context.Context, api coreiface.CoreAPI, c cid.Cid) ([]cid.Cid, error) {
	res := []cid.Cid{c}

	var recursiveFunc func(cid.Cid) error

	recursiveFunc = func(ci cid.Cid) error {
		obj, err := api.Dag().Get(ctx, ci)
		if err != nil {
			return err
		}
		if len(obj.Links()) > 0 {
			for _, link := range obj.Links() {
				res = append(res, link.Cid)
				err = recursiveFunc(link.Cid)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	err := recursiveFunc(c)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//GetNewFile 维护链上有效文件所有文件片的有序列表
func GetNewFile(ctx context.Context, ds datastore.Datastore, api coreiface.CoreAPI) error {
	// 拉取链上文件列表
	fileCidList, err := GetValidFileList()
	if err != nil {
		return err
	}
	// 拉取dag树
	for _, c := range fileCidList {
		if ok := storeCidMSet.Has(c); !ok {
			dagList, err := dagGetRecursive(ctx, api, c)
			if err != nil {
				return err
			}
			// 将dag树 的cid列表去重的放入列表
			for _, c2 := range dagList {
				if ok = storeCidMSet.Visit(c2); ok {
					// todo 维护有序列表
				}
			}
		}
	}

	return nil
}

func GenerateStoreProf(s string) (string, error) {
	return "", nil
}

func GetSPV(s string) ([]string, error) {
	return nil, nil
}

func FindBestCid(s string) (cid.Cid, error) {
	return cid.Cid{}, nil
}
