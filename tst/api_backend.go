// Copyright 2015 The go-utchain Authors
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

package tst

import (
	"context"
	"math/big"

	"github.com/utchain/go-utchain/accounts"
	"github.com/utchain/go-utchain/common"
	"github.com/utchain/go-utchain/common/math"
	"github.com/utchain/go-utchain/core"
	"github.com/utchain/go-utchain/core/bloombits"
	"github.com/utchain/go-utchain/core/state"
	"github.com/utchain/go-utchain/core/types"
	"github.com/utchain/go-utchain/core/vm"
	"github.com/utchain/go-utchain/tst/downloader"
	"github.com/utchain/go-utchain/tst/gasprice"
	"github.com/utchain/go-utchain/tstdb"
	"github.com/utchain/go-utchain/event"
	"github.com/utchain/go-utchain/params"
	"github.com/utchain/go-utchain/rpc"
)

// TstApiBackend implements ethapi.Backend for full nodes
type TstApiBackend struct {
	tst *UTChain
	gpo *gasprice.Oracle
}

func (b *TstApiBackend) ChainConfig() *params.ChainConfig {
	return b.tst.chainConfig
}

func (b *TstApiBackend) CurrentBlock() *types.Block {
	return b.tst.blockchain.CurrentBlock()
}

func (b *TstApiBackend) SetHead(number uint64) {
	b.tst.protocolManager.downloader.Cancel()
	b.tst.blockchain.SetHead(number)
}

func (b *TstApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.tst.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.tst.blockchain.CurrentBlock().Header(), nil
	}
	return b.tst.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *TstApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.tst.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.tst.blockchain.CurrentBlock(), nil
	}
	return b.tst.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *TstApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.tst.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.tst.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *TstApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.tst.blockchain.GetBlockByHash(blockHash), nil
}

func (b *TstApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.tst.chainDb, blockHash, core.GetBlockNumber(b.tst.chainDb, blockHash)), nil
}

func (b *TstApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.tst.chainDb, blockHash, core.GetBlockNumber(b.tst.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *TstApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.tst.blockchain.GetTdByHash(blockHash)
}

func (b *TstApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.tst.BlockChain(), nil)
	return vm.NewEVM(context, state, b.tst.chainConfig, vmCfg), vmError, nil
}

func (b *TstApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.tst.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *TstApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.tst.BlockChain().SubscribeChainEvent(ch)
}

func (b *TstApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.tst.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *TstApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.tst.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *TstApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.tst.BlockChain().SubscribeLogsEvent(ch)
}

func (b *TstApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.tst.txPool.AddLocal(signedTx)
}

func (b *TstApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.tst.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *TstApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.tst.txPool.Get(hash)
}

func (b *TstApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.tst.txPool.State().GetNonce(addr), nil
}

func (b *TstApiBackend) Stats() (pending int, queued int) {
	return b.tst.txPool.Stats()
}

func (b *TstApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.tst.TxPool().Content()
}

func (b *TstApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.tst.TxPool().SubscribeTxPreEvent(ch)
}

func (b *TstApiBackend) Downloader() *downloader.Downloader {
	return b.tst.Downloader()
}

func (b *TstApiBackend) ProtocolVersion() int {
	return b.tst.TstVersion()
}

func (b *TstApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *TstApiBackend) ChainDb() tstdb.Database {
	return b.tst.ChainDb()
}

func (b *TstApiBackend) EventMux() *event.TypeMux {
	return b.tst.EventMux()
}

func (b *TstApiBackend) AccountManager() *accounts.Manager {
	return b.tst.AccountManager()
}

func (b *TstApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.tst.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *TstApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.tst.bloomRequests)
	}
}
