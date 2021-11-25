package util

import (
	"fmt"
	"github.com/Hyperledger-TWGC/tjfoc-gm/sm4"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
)

func EncryptBlock(block blocks.Block, secret []byte) (blocks.Block, error) {
	tar, bType, crypt, auth, err := cid.ParseBlocInfo(block.Cid().Prefix().BlockInfo)
	if err != nil {
		return nil, err
	}
	if crypt == cid.Crypt_Y {
		return nil, fmt.Errorf("块已加密")
	}
	/*if auth == cid.Auth_N {
		return nil, fmt.Errorf("不需鉴权块不应该加密")
	}*/
	/*if btype != cid.BlockType_root {
		return nil, fmt.Errorf("当前非文件头无需加密")
	}*/
	// 当前加密文件块默认需要鉴权 todo 可能加密文件块不需要鉴权
	info := cid.GetBlockInfo(tar, bType, cid.Crypt_Y, auth)

	data := block.RawData()

	iv := make([]byte, sm4.BlockSize)
	encryptData, err := sm4Encrypt(secret, iv, data)
	if err != nil {
		return nil, err
	}

	return merkledag.NewRawNodeWithBlockInfo(encryptData, info).Block, nil
}

func DecryptBlock(nd ipld.Node, secret []byte) (ipld.Node, error) {
	_, _, crypt, _, err := cid.ParseBlocInfo(nd.Cid().Prefix().BlockInfo)
	if err != nil {
		return nil, err
	}
	if crypt == cid.Crypt_N || secret == nil {
		return nd, nil
	}

	data := nd.RawData()

	iv := make([]byte, sm4.BlockSize)
	decryptData, err := sm4Decrypt(secret, iv, data)
	if err != nil {
		return nil, err
	}
	pnd, err := merkledag.DecodeProtobuf(decryptData)
	if err != nil {
		return nil, err
	}
	return pnd, nil
}
