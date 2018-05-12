// Copyright 2016 The go-utchain Authors
// This file is part of the go-utchain library.
//
// The go-utchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-utchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-utchain library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light UTChain Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/utchain/go-utchain/accounts"
	"github.com/utchain/go-utchain/common"
	"github.com/utchain/go-utchain/common/hexutil"
	"github.com/utchain/go-utchain/consensus"
	"github.com/utchain/go-utchain/core"
	"github.com/utchain/go-utchain/core/bloombits"
	"github.com/utchain/go-utchain/core/types"
	"github.com/utchain/go-utchain/tst"
	"github.com/utchain/go-utchain/tst/downloader"
	"github.com/utchain/go-utchain/tst/filters"
	"github.com/utchain/go-utchain/tst/gasprice"
	"github.com/utchain/go-utchain/tstdb"
	"github.com/utchain/go-utchain/event"
	"github.com/utchain/go-utchain/internal/ethapi"
	"github.com/utchain/go-utchain/light"
	"github.com/utchain/go-utchain/log"
	"github.com/utchain/go-utchain/node"
	"github.com/utchain/go-utchain/p2p"
	"github.com/utchain/go-utchain/p2p/discv5"
	"github.com/utchain/go-utchain/params"
	rpc "github.com/utchain/go-utchain/rpc"
)

type LightUTChain struct {
	config *tst.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb tstdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *tst.Config) (*LightUTChain, error) {
	chainDb, err := tst.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	ltst := &LightUTChain{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           tst.CreateConsensusEngine(ctx, &config.Tstash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     tst.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	ltst.relay = NewLesTxRelay(peers, ltst.reqDist)
	ltst.serverPool = newServerPool(chainDb, quitSync, &ltst.wg)
	ltst.retriever = newRetrieveManager(peers, ltst.reqDist, ltst.serverPool)
	ltst.odr = NewLesOdr(chainDb, ltst.chtIndexer, ltst.bloomTrieIndexer, ltst.bloomIndexer, ltst.retriever)
	if ltst.blockchain, err = light.NewLightChain(ltst.odr, ltst.chainConfig, ltst.engine); err != nil {
		return nil, err
	}
	ltst.bloomIndexer.Start(ltst.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		ltst.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	ltst.txPool = light.NewTxPool(ltst.chainConfig, ltst.blockchain, ltst.relay)
	if ltst.protocolManager, err = NewProtocolManager(ltst.chainConfig, true, ClientProtocolVersions, config.NetworkId, ltst.eventMux, ltst.engine, ltst.peers, ltst.blockchain, nil, chainDb, ltst.odr, ltst.relay, quitSync, &ltst.wg); err != nil {
		return nil, err
	}
	ltst.ApiBackend = &LesApiBackend{ltst, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	ltst.ApiBackend.gpo = gasprice.NewOracle(ltst.ApiBackend, gpoParams)
	return ltst, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Tsterbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Tsterbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Tsterbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the utereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightUTChain) APIs() []rpc.API {
	return append(ethapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "tst",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "tst",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "tst",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightUTChain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightUTChain) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightUTChain) TxPool() *light.TxPool              { return s.txPool }
func (s *LightUTChain) Engine() consensus.Engine           { return s.engine }
func (s *LightUTChain) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightUTChain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightUTChain) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightUTChain) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// UTChain protocol implementation.
func (s *LightUTChain) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// UTChain protocol.
func (s *LightUTChain) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
