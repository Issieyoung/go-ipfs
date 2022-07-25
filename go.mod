module github.com/ipfs/go-ipfs

require (
	bazil.org/fuse v0.0.0-20200117225306-7b5117fecadc
	contrib.go.opencensus.io/exporter/prometheus v0.3.0
	github.com/Hyperledger-TWGC/tjfoc-gm v1.4.0
	github.com/blang/semver/v4 v4.0.0
	github.com/cheggaaa/pb v1.0.29
	github.com/coreos/go-systemd/v22 v22.3.1
	github.com/dustin/go-humanize v1.0.0
	github.com/elgris/jsondiff v0.0.0-20160530203242-765b5c24c302
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gabriel-vasile/mimetype v1.2.0
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/google/uuid v1.2.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/ipfs/go-bitswap v0.3.4
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-blockservice v0.1.7
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-cidutil v0.0.2
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-detect-race v0.0.1
	github.com/ipfs/go-ds-badger v0.2.6
	github.com/ipfs/go-ds-flatfs v0.4.5
	github.com/ipfs/go-ds-leveldb v0.4.2
	github.com/ipfs/go-ds-measure v0.1.0
	github.com/ipfs/go-filestore v0.0.3
	github.com/ipfs/go-fs-lock v0.0.6
	github.com/ipfs/go-graphsync v0.8.0
	github.com/ipfs/go-ipfs-auth/selector v0.0.0
	github.com/ipfs/go-ipfs-auth/standard v0.0.0
	github.com/ipfs/go-ipfs-backup v0.0.0
	github.com/ipfs/go-ipfs-blockstore v0.1.6
	github.com/ipfs/go-ipfs-chunker v0.0.5
	github.com/ipfs/go-ipfs-cmds v0.6.0
	github.com/ipfs/go-ipfs-config v0.14.0
	github.com/ipfs/go-ipfs-exchange-interface v0.0.1
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipfs-files v0.0.8
	github.com/ipfs/go-ipfs-keystore v0.0.2
	github.com/ipfs/go-ipfs-pinner v0.1.2
	github.com/ipfs/go-ipfs-posinfo v0.0.1
	github.com/ipfs/go-ipfs-provider v0.5.1
	github.com/ipfs/go-ipfs-routing v0.1.0
	github.com/ipfs/go-ipfs-util v0.0.2
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-ipld-git v0.0.4
	github.com/ipfs/go-ipns v0.1.0
	github.com/ipfs/go-log v1.0.5
	github.com/ipfs/go-merkledag v0.4.0
	github.com/ipfs/go-metrics-interface v0.0.1
	github.com/ipfs/go-metrics-prometheus v0.0.2
	github.com/ipfs/go-mfs v0.1.2
	github.com/ipfs/go-namesys v0.3.0
	github.com/ipfs/go-path v0.0.9
	github.com/ipfs/go-pinning-service-http-client v0.1.0
	github.com/ipfs/go-unixfs v0.2.5
	github.com/ipfs/go-verifcid v0.0.1
	github.com/ipfs/interface-go-ipfs-core v0.4.0
	github.com/ipfs/tar-utils v0.0.1
	github.com/ipld/go-car v0.3.1
	github.com/jbenet/go-random v0.0.0-20190219211222-123a90aedc0c
	github.com/jbenet/go-temp-err-catcher v0.1.0
	github.com/jbenet/goprocess v0.1.4
	github.com/libp2p/go-doh-resolver v0.3.1
	github.com/libp2p/go-libp2p v0.14.2
	github.com/libp2p/go-libp2p-circuit v0.4.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-discovery v0.5.1
	github.com/libp2p/go-libp2p-http v0.2.0
	github.com/libp2p/go-libp2p-kad-dht v0.12.2
	github.com/libp2p/go-libp2p-kbucket v0.4.7
	github.com/libp2p/go-libp2p-loggables v0.1.0
	github.com/libp2p/go-libp2p-mplex v0.4.1
	github.com/libp2p/go-libp2p-noise v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.2.7
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-pubsub-router v0.4.0
	github.com/libp2p/go-libp2p-quic-transport v0.11.1
	github.com/libp2p/go-libp2p-record v0.1.3
	github.com/libp2p/go-libp2p-routing-helpers v0.2.3
	github.com/libp2p/go-libp2p-swarm v0.5.0
	github.com/libp2p/go-libp2p-testing v0.4.0
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/libp2p/go-libp2p-yamux v0.5.4
	github.com/libp2p/go-socket-activation v0.0.2
	github.com/libp2p/go-tcp-transport v0.2.2
	github.com/libp2p/go-ws-transport v0.4.0
	github.com/lucas-clemente/quic-go v0.28.1
	github.com/miekg/dns v1.1.41
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.3.2
	github.com/multiformats/go-multiaddr-dns v0.3.1
	github.com/multiformats/go-multibase v0.0.3
	github.com/multiformats/go-multihash v0.0.15
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/common v0.18.0
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/whyrusleeping/go-sysinfo v0.0.0-20190219211824-4a357d4b90b1
	github.com/whyrusleeping/multiaddr-filter v0.0.0-20160516205228-e903e4adabd7
	go.opencensus.io v0.23.0
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a
)

go 1.15

replace (
	github.com/ipfs/go-bitswap => ../go-bitswap
	github.com/ipfs/go-cid => ../ipld/cid/go-cid
	github.com/ipfs/go-ipfs-auth/auth-source-eth => ../go-ipfs-auth/auth-source-eth
	github.com/ipfs/go-ipfs-auth/selector => ../go-ipfs-auth/selector
	github.com/ipfs/go-ipfs-auth/standard => ../go-ipfs-auth/standard
	github.com/ipfs/go-ipfs-backup => ../go-ipfs-backup
	github.com/ipfs/go-ipfs-chunker => ../go-ipfs-chunker
	github.com/ipfs/go-merkledag => ../ipld/go-ipld-merkledag
	github.com/ipfs/go-peertaskqueue => ../go-peertaskqueue
	github.com/ipfs/go-unixfs => ../go-unixfs
)
