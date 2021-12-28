package blockchain

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-auth/selector"
	"github.com/ipfs/go-ipfs-auth/standard/model"
	"github.com/ipfs/go-ipfs-backup/allocate"
	"github.com/ipfs/go-ipfs-backup/backup"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/ipfs/go-ipfs/util"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prometheus/common/log"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	cmds "github.com/ipfs/go-ipfs-cmds"
	files "github.com/ipfs/go-ipfs-files"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	mh "github.com/multiformats/go-multihash"
)

// ErrDepthLimitExceeded indicates that the max depth has been exceeded.
var ErrDepthLimitExceeded = fmt.Errorf("depth limit exceeded")

type AddEvent struct {
	Name   string
	Hash   string `json:",omitempty"`
	Bytes  int64  `json:",omitempty"`
	Size   string `json:",omitempty"`
	Time   int64
	Backup string
}

const (
	quietOptionName       = "quiet"
	quieterOptionName     = "quieter"
	silentOptionName      = "silent"
	progressOptionName    = "progress"
	trickleOptionName     = "trickle"
	wrapOptionName        = "wrap-with-directory"
	onlyHashOptionName    = "only-hash"
	chunkerOptionName     = "chunker"
	pinOptionName         = "pin"
	rawLeavesOptionName   = "raw-leaves"
	noCopyOptionName      = "nocopy"
	fstoreCacheOptionName = "fscache"
	cidVersionOptionName  = "cid-version"
	hashOptionName        = "hash"
	inlineOptionName      = "inline"
	inlineLimitOptionName = "inline-limit"
	privateOptionName     = "private"
	recursive             = "recursive"
	fileStoreDays         = "fileStoreDays"
	isFile                = "isFile"
	ownerAddress          = "owner"
)

const adderOutChanSize = 8

var BlockchainCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "区块链相关的命令",
		ShortDescription: `区块链相关的命令`,
		LongDescription:  `区块链相关的命令`,
	},
	Subcommands: map[string]*cmds.Command{
		"file": FileCmd,
		"peer": PeerCmd,
	},
}

var FileCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "文件相关的命令",
		ShortDescription: `文件相关的命令`,
		LongDescription:  `文件相关的命令`,
	},
	Subcommands: map[string]*cmds.Command{
		"add":      AddCmd,
		"delete":   DeleteCmd,
		"backup":   BackupInfoCmd,
		"recharge": RechargeCmd,
	},
}

var AddCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Add a file or directory to IPFS.",
		ShortDescription: `
Adds the content of <path> to IPFS. Use -r to add directories (recursively).
`,
		LongDescription: `
Adds the content of <path> to IPFS. Use -r to add directories.
Note that directories are added recursively, to form the IPFS
MerkleDAG.

If the daemon is not running, it will just add locally.
If the daemon is started later, it will be advertised after a few
seconds when the reprovider runs.

The wrap option, '-w', wraps the file (or files, if using the
recursive option) in a directory. This directory contains only
the files which have been added, and means that the file retains
its filename. For example:

  > ipfs add example.jpg
  added QmbFMke1KXqnYyBBWxB74N4c5SBnJMVAiMNRcGu6x1AwQH example.jpg
  > ipfs add example.jpg -w
  added QmbFMke1KXqnYyBBWxB74N4c5SBnJMVAiMNRcGu6x1AwQH example.jpg
  added QmaG4FuMqEBnQNn3C8XJ5bpW8kLs7zq2ZXgHptJHbKDDVx

You can now refer to the added file in a gateway, like so:

  /ipfs/QmaG4FuMqEBnQNn3C8XJ5bpW8kLs7zq2ZXgHptJHbKDDVx/example.jpg

The chunker option, '-s', specifies the chunking strategy that dictates
how to break files into blocks. Blocks with same content can
be deduplicated. Different chunking strategies will produce different
hashes for the same file. The default is a fixed block size of
256 * 1024 bytes, 'size-262144'. Alternatively, you can use the
Buzhash or Rabin fingerprint chunker for content defined chunking by
specifying buzhash or rabin-[min]-[avg]-[max] (where min/avg/max refer
to the desired chunk sizes in bytes), e.g. 'rabin-262144-524288-1048576'.

The following examples use very small byte sizes to demonstrate the
properties of the different chunkers on a small file. You'll likely
want to use a 1024 times larger chunk sizes for most files.

  > ipfs add --chunker=size-2048 ipfs-logo.svg
  added QmafrLBfzRLV4XSH1XcaMMeaXEUhDJjmtDfsYU95TrWG87 ipfs-logo.svg
  > ipfs add --chunker=rabin-512-1024-2048 ipfs-logo.svg
  added Qmf1hDN65tR55Ubh2RN1FPxr69xq3giVBz1KApsresY8Gn ipfs-logo.svg

You can now check what blocks have been created by:

  > ipfs object links QmafrLBfzRLV4XSH1XcaMMeaXEUhDJjmtDfsYU95TrWG87
  QmY6yj1GsermExDXoosVE3aSPxdMNYr6aKuw3nA8LoWPRS 2059
  Qmf7ZQeSxq2fJVJbCmgTrLLVN9tDR9Wy5k75DxQKuz5Gyt 1195
  > ipfs object links Qmf1hDN65tR55Ubh2RN1FPxr69xq3giVBz1KApsresY8Gn
  QmY6yj1GsermExDXoosVE3aSPxdMNYr6aKuw3nA8LoWPRS 2059
  QmerURi9k4XzKCaaPbsK6BL5pMEjF7PGphjDvkkjDtsVf3 868
  QmQB28iwSriSUSMqG2nXDTLtdPHgWb4rebBrU7Q1j4vxPv 338

Finally, a note on hash determinism. While not guaranteed, adding the same
file/directory with the same flags will almost always result in the same output
hash. However, almost all of the flags provided by this command (other than pin,
only-hash, and progress/status related flags) will change the final hash.
`,
	},

	Arguments: []cmds.Argument{
		cmds.FileArg("path", true, true, "The path to a file to be added to IPFS.").EnableRecursive().EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.OptionRecursivePath, // a builtin option that allows recursive paths (-r, --recursive)
		cmds.OptionDerefArgs,     // a builtin option that resolves passed in filesystem links (--dereference-args)
		cmds.OptionStdinName,     // a builtin option that optionally allows wrapping stdin into a named file
		cmds.OptionHidden,
		cmds.OptionIgnore,
		cmds.OptionIgnoreRules,
		cmds.BoolOption(quietOptionName, "q", "Write minimal output."),
		cmds.BoolOption(privateOptionName, "pri", "是否为私密文件，私密文件只在内网传播（后续可实现权限控制），从公网无法获取该文件").WithDefault(false),
		cmds.BoolOption(quieterOptionName, "Q", "Write only final hash."),
		cmds.BoolOption(silentOptionName, "Write no output."),
		cmds.BoolOption(progressOptionName, "p", "Stream progress data."),
		cmds.BoolOption(trickleOptionName, "t", "Use trickle-dag format for dag generation."),
		cmds.BoolOption(onlyHashOptionName, "n", "Only chunk and hash - do not write to disk."),
		cmds.BoolOption(wrapOptionName, "w", "Wrap files with a directory object."),
		cmds.StringOption(chunkerOptionName, "s", "Chunking algorithm, size-[bytes], rabin-[min]-[avg]-[max] or buzhash").WithDefault("size-262144"),
		cmds.BoolOption(pinOptionName, "Pin this object when adding.").WithDefault(true),
		cmds.BoolOption(rawLeavesOptionName, "Use raw blocks for leaf nodes. (experimental)"),
		cmds.BoolOption(noCopyOptionName, "Add the file using filestore. Implies raw-leaves. (experimental)"),
		cmds.BoolOption(fstoreCacheOptionName, "Check the filestore for pre-existing blocks. (experimental)"),
		cmds.IntOption(cidVersionOptionName, "CID version. Defaults to 0 unless an option that depends on CIDv1 is passed. (experimental)"),
		cmds.StringOption(hashOptionName, "Hash function to use. Implies CIDv1 if not sha2-256. (experimental)").WithDefault("sha2-256"),
		cmds.BoolOption(inlineOptionName, "Inline small blocks into CIDs. (experimental)"),
		cmds.IntOption(inlineLimitOptionName, "Maximum block size to inline. (experimental)").WithDefault(32),
		cmds.IntOption(fileStoreDays, "how many days you want to store in blockchain").WithDefault(30),
		cmds.StringOption(ownerAddress, ""),
	},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		quiet, _ := req.Options[quietOptionName].(bool)
		quieter, _ := req.Options[quieterOptionName].(bool)
		quiet = quiet || quieter

		silent, _ := req.Options[silentOptionName].(bool)

		if quiet || silent {
			return nil
		}

		// ipfs cli progress bar defaults to true unless quiet or silent is used
		_, found := req.Options[progressOptionName].(bool)
		if !found {
			req.Options[progressOptionName] = true
		}

		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		startTime := time.Now().UnixNano()
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}

		// 提前检查网络状态
		setting := allocate.Setting{
			Strategy:  0,
			TargetNum: 1,
		}
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		cfg, err := node.Repo.Config()
		if err != nil {
			return err
		}
		if cfg.BackupNum > 0 {
			setting.TargetNum = cfg.BackupNum
		}

		peerList, err := getReliablePeer(req.Context, node, api, 10)
		if err != nil {
			return err
		}
		if len(peerList) < setting.TargetNum {
			return fmt.Errorf("在线节点数不满足备份条件")
		}

		owner := req.Options[ownerAddress].(string)
		if owner != "" && owner != "self" {
			// todo 检查信任状态

		}

		progress, _ := req.Options[progressOptionName].(bool)
		trickle, _ := req.Options[trickleOptionName].(bool)
		wrap, _ := req.Options[wrapOptionName].(bool)
		hash, _ := req.Options[onlyHashOptionName].(bool)
		silent, _ := req.Options[silentOptionName].(bool)
		chunker, _ := req.Options[chunkerOptionName].(string)
		dopin, _ := req.Options[pinOptionName].(bool)
		rawblks, rbset := req.Options[rawLeavesOptionName].(bool)
		nocopy, _ := req.Options[noCopyOptionName].(bool)
		fscache, _ := req.Options[fstoreCacheOptionName].(bool)
		cidVer, cidVerSet := req.Options[cidVersionOptionName].(int)
		hashFunStr, _ := req.Options[hashOptionName].(string)
		inline, _ := req.Options[inlineOptionName].(bool)
		inlineLimit, _ := req.Options[inlineLimitOptionName].(int)
		b := req.Options[privateOptionName].(bool)
		days := req.Options[fileStoreDays].(int)
		fmt.Println(days)

		if b {
			cidVerSet = true
			cidVer = 2
		}

		hashFunCode, ok := mh.Names[strings.ToLower(hashFunStr)]
		if !ok {
			return fmt.Errorf("unrecognized hash function: %s", strings.ToLower(hashFunStr))
		}

		enc, err := cmdenv.GetCidEncoder(req)
		if err != nil {
			return err
		}

		toadd := req.Files
		if wrap {
			toadd = files.NewSliceDirectory([]files.DirEntry{
				files.FileEntry("", req.Files),
			})
		}

		opts := []options.UnixfsAddOption{
			options.Unixfs.Hash(hashFunCode),

			options.Unixfs.Inline(inline),
			options.Unixfs.InlineLimit(inlineLimit),

			options.Unixfs.Chunker(chunker),

			options.Unixfs.Pin(dopin),
			options.Unixfs.HashOnly(hash),
			options.Unixfs.FsCache(fscache),
			options.Unixfs.Nocopy(nocopy),

			options.Unixfs.Progress(progress),
			options.Unixfs.Silent(silent),
		}

		if cidVerSet {
			opts = append(opts, options.Unixfs.CidVersion(cidVer))
		}

		if rbset {
			opts = append(opts, options.Unixfs.RawLeaves(rawblks))
		}

		if trickle {
			opts = append(opts, options.Unixfs.Layout(options.TrickleLayout))
		}

		opts = append(opts, nil) // events option placeholder

		var added int
		addit := toadd.Entries()
		for addit.Next() {
			_, dir := addit.Node().(files.Directory)
			errCh := make(chan error, 1)
			events := make(chan interface{}, adderOutChanSize)
			opts[len(opts)-1] = options.Unixfs.Events(events)

			go func() {
				var err error
				defer close(events)
				_, err = api.Unixfs().Add(req.Context, addit.Node(), opts...)
				errCh <- err
			}()

			for event := range events {
				output, ok := event.(*coreiface.AddEvent)
				if !ok {
					return errors.New("unknown event type")
				}

				h := ""
				if output.Path != nil {
					h = enc.Encode(output.Path.Cid())
				}
				if !dir && addit.Name() != "" {
					output.Name = addit.Name()
				} else {
					output.Name = path.Join(addit.Name(), output.Name)
				}

				// 分发给随机peer
				uid, err := util.GetUUIDString()
				if err != nil {
					return err
				}
				// 计算文件大小(kb)
				s1, err := strconv.Atoi(output.Size)
				s := s1 / 1024
				if s1%1024 != 0 {
					s += 1
				}
				// 链上文件信息记录  暂时不需要
				err = selector.AddFile(model.IpfsFileInfo{
					Cid:       h,
					Uid:       uid,
					State:     0,
					Size:      int64(s),
					StoreDays: int64(days),
					Owner:     owner,
				})
				if err != nil {
					return err
				}

				backupInfo := "备份运行中"
				_, err = backup.GetFileBackupInfo(node.Repo.Datastore(), h)
				if err == datastore.ErrNotFound {
					// TODO 是否要等待分发任务结果再返回
					var allocateFunc func() error = func() error {
						// 检查节点是否可以连接
						c, err := cid.Decode(h)
						if err != nil {
							return err
						}
						ctx := context.TODO()
						// 查找所有块
						blockList, err := BlockGetRecursive(ctx, api, c)
						if err != nil {
							return err
						}

						peerList, err = getReliablePeer(req.Context, node, api, 10)
						if err != nil {
							return err
						}
						if len(peerList) < setting.TargetNum {
							return fmt.Errorf("在线节点数不满足备份条件")
						}

						return Allocate(node, blockList, peerList, setting, uid, uint64(s))
					}
					errChan := make(chan error)
					go func() {
						errChan <- allocateFunc()
					}()
					select {
					case err = <-errChan:
					case <-req.Context.Done():
						return nil
					}

					if err != nil {
						backupInfo = err.Error()
					}
				} else {
					backupInfo = "文件已有备份"
				}

				if err := res.Emit(&AddEvent{
					Name:   output.Name,
					Hash:   h,
					Bytes:  output.Bytes,
					Size:   output.Size,
					Time:   time.Now().UnixNano() - startTime,
					Backup: backupInfo,
				}); err != nil {
					return err
				}
			}

			if err := <-errCh; err != nil {
				return err
			}
			added++
		}

		if addit.Err() != nil {
			return addit.Err()
		}

		if added == 0 {
			return fmt.Errorf("expected a file argument")
		}

		return nil
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			sizeChan := make(chan int64, 1)
			outChan := make(chan interface{})
			req := res.Request()

			// Could be slow.
			go func() {
				size, err := req.Files.Size()
				if err != nil {
					log.Warnf("error getting files size: %s", err)
					// see comment above
					return
				}

				sizeChan <- size
			}()

			progressBar := func(wait chan struct{}) {
				defer close(wait)

				quiet, _ := req.Options[quietOptionName].(bool)
				quieter, _ := req.Options[quieterOptionName].(bool)
				quiet = quiet || quieter

				progress, _ := req.Options[progressOptionName].(bool)

				var bar *pb.ProgressBar
				if progress {
					bar = pb.New64(0).SetUnits(pb.U_BYTES)
					bar.ManualUpdate = true
					bar.ShowTimeLeft = false
					bar.ShowPercent = false
					bar.Output = os.Stderr
					bar.Start()
				}

				lastFile := ""
				lastHash := ""
				var totalProgress, prevFiles, lastBytes int64

			LOOP:
				for {
					select {
					case out, ok := <-outChan:
						if !ok {
							if quieter {
								fmt.Fprintln(os.Stdout, lastHash)
							}

							break LOOP
						}
						output := out.(*AddEvent)
						if len(output.Hash) > 0 {
							lastHash = output.Hash
							if quieter {
								continue
							}

							if progress {
								// clear progress bar line before we print "added x" output
								fmt.Fprintf(os.Stderr, "\033[2K\r")
							}
							if quiet {
								fmt.Fprintf(os.Stdout, "%s\n", output.Hash)
							} else {
								fmt.Fprintf(os.Stdout, "added %s %s\n", output.Hash, cmdenv.EscNonPrint(output.Name))
							}

						} else {
							if !progress {
								continue
							}

							if len(lastFile) == 0 {
								lastFile = output.Name
							}
							if output.Name != lastFile || output.Bytes < lastBytes {
								prevFiles += lastBytes
								lastFile = output.Name
							}
							lastBytes = output.Bytes
							delta := prevFiles + lastBytes - totalProgress
							totalProgress = bar.Add64(delta)
						}

						if progress {
							bar.Update()
						}
					case size := <-sizeChan:
						if progress {
							bar.Total = size
							bar.ShowPercent = true
							bar.ShowBar = true
							bar.ShowTimeLeft = true
						}
					case <-req.Context.Done():
						// don't set or print error here, that happens in the goroutine below
						return
					}
				}

				if progress && bar.Total == 0 && bar.Get() != 0 {
					bar.Total = bar.Get()
					bar.ShowPercent = true
					bar.ShowBar = true
					bar.ShowTimeLeft = true
					bar.Update()
				}
			}

			if e := res.Error(); e != nil {
				close(outChan)
				return e
			}

			wait := make(chan struct{})
			go progressBar(wait)

			defer func() { <-wait }()
			defer close(outChan)

			for {
				v, err := res.Next()
				if err != nil {
					if err == io.EOF {
						return nil
					}

					return err
				}

				select {
				case outChan <- v:
				case <-req.Context.Done():
					return req.Context.Err()
				}
			}
		},
	},
	Type: AddEvent{},
}

var DeleteCmd = &cmds.Command{
	Arguments: []cmds.Argument{
		cmds.StringArg("cid", true, false, "需要删除的文件的cid"),
	},
	Options: []cmds.Option{cmds.BoolOption(recursive).WithDefault(false)},
	Run: func(req *cmds.Request, emit cmds.ResponseEmitter, env cmds.Environment) error {
		// 清除指定cid的备份信息
		reFlag := req.Options[recursive].(bool)
		cStr := req.Arguments[0]
		c, err := cid.Decode(cStr)
		if err != nil {
			return err
		}
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}
		cids, err := CidGet(req.Context, api, c, reFlag)
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		err = selector.DeleteFile(cStr)
		if err != nil {
			return err
		}
		// 现在清除备份信息 todo 向备份节点传播删除信息
		return backup.Remove(node.Repo.Datastore(), cids...)
	},
	Helptext: cmds.HelpText{
		Tagline:          "",
		ShortDescription: "",
		LongDescription:  "清除备份信息（测试用）",
	},
	Type: nil,
}

var RechargeCmd = &cmds.Command{
	Arguments: []cmds.Argument{
		cmds.StringArg("cid", true, false, "需要充值的文件的cid"),
	},
	Options: []cmds.Option{cmds.IntOption(fileStoreDays).WithDefault(30)},
	Run: func(req *cmds.Request, emit cmds.ResponseEmitter, env cmds.Environment) error {
		// 清除指定cid的备份信息
		cid := req.Arguments[0]
		days := req.Options[fileStoreDays].(int)
		return selector.RechargeFile(cid, int64(days))
	},
	Helptext: cmds.HelpText{
		Tagline:          "",
		ShortDescription: "",
		LongDescription:  "",
	},
	Type: nil,
}

var BackupInfoCmd = &cmds.Command{
	Arguments: []cmds.Argument{
		cmds.StringArg("cid", false, false, "需要查询的文件的cid"),
	},
	Options: []cmds.Option{
		cmds.BoolOption(recursive).WithDefault(false),
		cmds.BoolOption(isFile).WithDefault(true),
	},
	Run: func(req *cmds.Request, emit cmds.ResponseEmitter, env cmds.Environment) error {
		// 清除指定cid的备份信息
		reFlag := req.Options[recursive].(bool)
		fileFlag := req.Options[isFile].(bool)

		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if len(req.Arguments) == 0 {
			all, err := backup.GetAll(node.Repo.Datastore())
			if err != nil {
				return err
			}
			return emit.Emit(all)
		}

		cStr := req.Arguments[0]
		c, err := cid.Decode(cStr)
		if err != nil {
			return err
		}
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}

		if fileFlag {
			info, err := backup.GetFileBackupInfo(node.Repo.Datastore(), cStr)
			if err != nil {
				return err
			}
			return emit.Emit(*info)
		}

		cids, err := CidGet(req.Context, api, c, reFlag)

		res := map[string]interface{}{}
		for _, s := range cids {
			info, err := backup.Get(node.Repo.Datastore(), s)
			if err != nil {
				res[s] = err.Error()
			} else {
				res[s] = info
			}
		}
		return emit.Emit(res)
	},
	Helptext: cmds.HelpText{
		Tagline:          "",
		ShortDescription: "",
		LongDescription:  "查询备份信息",
	},
	Type: backup.FileInfo{},
}

var InitPeerCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "初始化区块链身份",
		ShortDescription: "绑定区块链身份和节点id",
		LongDescription:  "绑定区块链链上身份和ipfs节点id，暂时不提供解绑和改绑的可能，绑定前请认真核对",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		pid := n.Identity.String()

		var addressList []string
		addrss, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(n.PeerHost))
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

		if len(addressList) == 0 {
			return fmt.Errorf("无外网地址")
		}

		peer := model.CorePeer{
			PeerId:    pid,
			Addresses: addressList,
		}
		err = selector.InitPeer(peer)
		if err != nil {
			return err
		}

		return res.Emit(peer)
	},
	Type: model.CorePeer{},
}

const (
	peerId = "peerId"
	number = "num"
)

var PeerCmd = &cmds.Command{
	Helptext:  cmds.HelpText{},
	Arguments: nil,
	Subcommands: map[string]*cmds.Command{
		"init": InitPeerCmd,
	},
	Options: []cmds.Option{
		cmds.StringOption(peerId, "pid", "查询节点的id，不填代表查询本节点").WithDefault(""),
		cmds.IntOption(number, "numb", "how many peer you want to get from blockchain").WithDefault(1),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		num := req.Options[number].(int)
		if num != 1 {
			list, err := selector.GetPeerList(num)
			if err != nil {
				return err
			}
			return res.Emit(list)
		}

		pid := req.Options[peerId].(string)
		if pid == "" {
			n, err := cmdenv.GetNode(env)
			if err != nil {
				return err
			}
			pid = n.Identity.String()
		}

		_, err := peer.Decode(pid)
		if err != nil {
			return err
		}

		peer, err := selector.GetPeer(pid)
		if err != nil {
			return err
		}
		return res.Emit([]model.CorePeer{peer})
	},
	Type: []model.CorePeer{},
}
