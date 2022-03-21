package mining

import (
	"context"
	"encoding/base64"
	"encoding/json"
	selector "github.com/bdengine/go-ipfs-blockchain-selector"
	"github.com/bdengine/go-ipfs-blockchain-standard/algorithm"
	"github.com/bdengine/go-ipfs-blockchain-standard/standardConst"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p-core/peer"
	mh "github.com/multiformats/go-multihash"
	"sync"
	"time"
)

type PreSend struct {
	FileNum        int
	StoreChallenge []byte
	MerkleRoot     string
}

type Server struct {
	Tree           *redBlackTree
	Pre            *PreSend
	Pid            peer.ID
	api            coreiface.CoreAPI
	ChallengeStage int
	S              string
	FileSet        *sync.Map
}

var MServer *Server

const preSendPrefix = "/preSend"

func NewMineServer(sc string, pid peer.ID, api coreiface.CoreAPI, ds datastore.Datastore) {
	b, _ := base64.StdEncoding.DecodeString(sc)
	MServer = &Server{
		Tree: &redBlackTree{
			lock:           sync.RWMutex{},
			StoreChallenge: b,
			api:            api,
		},
		Pid:            pid,
		api:            api,
		ChallengeStage: standardConst.CycleStageWait,
		FileSet:        &sync.Map{},
	}
	err := LoadFileSetLocal(ds)
	// 初始化Pre
	pre := &PreSend{}
	// 从ds中加载持久化的pre
	preKey := datastore.NewKey(preSendPrefix)
	get, err := ds.Get(preKey)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	// 检查pre的storeChallenge是否是现阶段的
	err = json.Unmarshal(get, pre)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if ByteEqual(pre.StoreChallenge, b) {
		MServer.Pre = pre
	}
}

func IsSend() bool {
	MServer.Tree.lock.RLock()
	defer MServer.Tree.lock.RUnlock()
	if MServer.Pre == nil {
		return false
	}
	return MServer.Tree.getSize() == MServer.Pre.FileNum && ByteEqual(MServer.Tree.StoreChallenge, MServer.Pre.StoreChallenge)
}

func AddFile(ctx context.Context, c cid.Cid) error {
	ctx1, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cList, err := dagGetRecursive(ctx1, MServer.api, c)
	if err != nil {
		return err
	}
	MServer.Tree.addMany(ctx, cList)
	return nil
}

func AddFileSlice(c cid.Cid) bool {
	return MServer.Tree.add(c)
}

func UpdateCTreeChallenge(c string) {
	MServer.Tree.updateStoreChallenge(c)
}

func GetMerkleTree() (SortList, int, bool) {
	tree, i := MServer.Tree.getMerkleTree(GetPidByte(MServer.Pid))
	if MServer.Pre == nil {
		return tree, i, false
	}
	rootStr, _ := algorithm.GetFromString(MServer.Pre.MerkleRoot)
	return tree, i, tree[len(tree)-1].Hash.Equal(rootStr[:])
}

func GetPidByte(pid peer.ID) []byte {
	d, _ := mh.Decode(mh.Multihash(pid))
	pidByte := d.Digest
	return pidByte
}

func UpdatePre(root string, ds datastore.Datastore) {
	p := &PreSend{
		FileNum:        MServer.Tree.getSize(),
		StoreChallenge: MServer.Tree.StoreChallenge,
		MerkleRoot:     root,
	}
	MServer.Pre = p
	marshal, _ := json.Marshal(p)
	preKey := datastore.NewKey(preSendPrefix)
	err := ds.Put(preKey, marshal)
	if err != nil {
		log.Errorf(err.Error())
	}
}

func GetCTreeSize() int {
	return MServer.Tree.size
}

const FileListKeyPre = "/chainFile"

// LoadFileSetLocal 数据库存储有效的文件列表
func LoadFileSetLocal(ds datastore.Datastore) error {
	q := query.Query{
		Prefix:   FileListKeyPre,
		KeysOnly: true,
	}
	results, err := ds.Query(q)
	if err != nil {
		return err
	}
	for result := range results.Next() {
		c, err := cid.Parse(result.Key)
		if err != nil {
			continue
		}
		MServer.FileSet.Store(c, nil)
	}
	return nil
}

func StoreToFileSet(c string, ds datastore.Datastore) error {
	k := datastore.NewKey(FileListKeyPre + "/" + c)
	return ds.Put(k, nil)
}

const MerkleTreeKeyPreFix = "/merkleTree"

type MerkleTreeStore struct {
	SortList
	Size           int
	StoreChallenge string
	RootStr        string
}

func StoreMerkleTree(ds datastore.Datastore, sList SortList, storeChallenge string, size int) error {
	rootStr := sList[len(sList)-1].Hash.String()
	store := MerkleTreeStore{
		SortList:       sList,
		Size:           size,
		StoreChallenge: storeChallenge,
		RootStr:        rootStr,
	}
	marshal, err := json.Marshal(store)
	if err != nil {
		return err
	}
	key := datastore.NewKey(MerkleTreeKeyPreFix + "/" + storeChallenge + "/" + rootStr)
	return ds.Put(key, marshal)
}

func FindMerkleTree(ds datastore.Datastore, storeChallenge string, rootStr string) (*MerkleTreeStore, error) {
	key := datastore.NewKey(MerkleTreeKeyPreFix + "/" + storeChallenge + "/" + rootStr)
	bytes, err := ds.Get(key)
	if err != nil {
		return nil, err
	}
	res := MerkleTreeStore{}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func PullNewFile(ctx context.Context, ds datastore.Datastore) {
	fileList, _ := selector.GetFileList(0)
	// 更新列表
	for _, s := range fileList {
		c, err := cid.Parse(s)
		if err != nil {
			continue
		}

		_, ok := MServer.FileSet.Load(c)
		if !ok {
			// 拉取文件
			err = AddFile(ctx, c)
			if err != nil {
				continue
			}
			// 添加进set
			MServer.FileSet.Store(c, nil)
			// 存储到datastore
			StoreToFileSet(c.String(), ds)
		}
	}
}
