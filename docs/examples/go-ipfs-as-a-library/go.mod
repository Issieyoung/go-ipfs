module github.com/ipfs/go-ipfs/examples/go-ipfs-as-a-library

go 1.14

require (
	github.com/ipfs/go-ipfs v0.7.0
	github.com/ipfs/go-ipfs-config v0.9.0
	github.com/ipfs/go-ipfs-files v0.0.8
	github.com/ipfs/interface-go-ipfs-core v0.4.0
	github.com/libp2p/go-libp2p-core v0.6.0
	github.com/libp2p/go-libp2p-peerstore v0.2.6
	github.com/multiformats/go-multiaddr v0.2.2
)

replace (
	github.com/ipfs/go-bitswap => ../../../../go-bitswap
	github.com/ipfs/go-cid => ../../../../go-cid
	github.com/bdengine/go-ipfs-blockchain-eth => ../../../../go-ipfs-blockchain-eth
	github.com/bdengine/go-ipfs-blockchain-selector => ../../../../go-ipfs-blockchain-selector
	github.com/bdengine/go-ipfs-blockchain-standard => ../../../../go-ipfs-blockchain-standard
	github.com/ipfs/go-ipfs-chunker => ../../../../go-ipfs-chunker
	github.com/ipfs/go-merkledag => ../../../../go-merkledag
	github.com/ipfs/go-peertaskqueue => ../../../../go-peertaskqueue
	github.com/ipfs/go-unixfs => ../../../../go-unixfs
    github.com/ipfs/go-ipfs => ./../../..
)
