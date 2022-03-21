package mining

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/bdengine/go-ipfs-blockchain-standard/algorithm"
	"github.com/bdengine/go-ipfs-blockchain-standard/dto"
	"github.com/google/uuid"
	block "github.com/ipfs/go-block-format"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mr-tron/base58"
	mh "github.com/multiformats/go-multihash"
	"sort"
	"testing"
)

func getRandomStr() string {
	binary1, _ := uuid.New().MarshalBinary()
	binary2, _ := uuid.New().MarshalBinary()
	binary1 = append(binary1, binary2...)
	return base64.StdEncoding.EncodeToString(binary1)
}

func TestBuildMerkleTreeStore(t *testing.T) {
	challenge := "KXrAMPQCTaOWijhqTLmPdwLHjNMJDktYpFMMKtUCBJk="
	challengeByte, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		t.Fatal(err)
	}
	hashSha256 := sha256.New()

	n := 100
	pls := make([]*algorithm.ProofLeaf, n)

	pid, err := peer.Decode("12D3KooWLUTBUkLnfdcJbyV1C7ZRsdFcDoN89mmYknFH5ef9pTyM")
	if err != nil {
		t.Fatal(err)
	}
	decode, _ := base58.Decode("12D3KooWLUTBUkLnfdcJbyV1C7ZRsdFcDoN89mmYknFH5ef9pTyM")
	fmt.Println(decode)
	d, _ := mh.Decode(mh.Multihash(pid))
	pidByte := d.Digest

	for i := 0; i < n; i++ {
		rowData := []byte(fmt.Sprintf("test%v", i))
		c := block.NewBlock(rowData).Cid()
		pls[i] = &algorithm.ProofLeaf{C: c, ChallengeHash: algorithm.HashMerkleBranchesForFile(challengeByte, rowData, hashSha256)}
	}
	// 此时是无序的
	sort.Slice(pls, func(i, j int) bool {
		return less(getHash(pls[i].C), getHash(pls[j].C))
	})

	merkleArr := algorithm.BuildMerkleTreeStore(pls, pidByte)
	root := merkleArr[len(merkleArr)-1]
	ok := checkIncreasing(merkleArr[:n])
	if !ok {
		t.Fatal(ok)
	}

	for i := 0; i < n; i++ {
		best := merkleArr[i]
		//content := []byte(fmt.Sprintf("test%v", i))
		merkleBlock := algorithm.GetMerkleBlock(merkleArr, i)
		proofDTO := dto.SVProofDTO{
			Cid:            best.C.String(),
			Pid:            pid.String(),
			SvProof:        merkleBlock,
			StoreChallenge: challenge,
			ProofRoot:      root.Hash.String(),
			ChallengeHash:  best.ChallengeHash[:],
			ProofLeaf:      best.Hash,
		}
		pb := GetPidByte(pid)
		ok := algorithm.VerifyProofLeaf2(pb, proofDTO.ChallengeHash, proofDTO.ProofLeaf, hashSha256)
		if !ok {
			t.Fatal("verify ProofLeaf2 fail")
		}
		r, _ := algorithm.GetFromString(proofDTO.ProofRoot)
		ok = algorithm.VerifyMerkleBlock(proofDTO.SvProof, proofDTO.ProofLeaf, r)
		if !ok {
			t.Fatal("verify Merkle Block fail")
		}
	}
}

func checkIncreasing(pls []*algorithm.ProofLeaf) bool {
	for i := 1; i < len(pls); i++ {
		if !less(getHash(pls[i-1].C), getHash(pls[i].C)) {
			return false
		}
	}
	return true
}
