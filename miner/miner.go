// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package miner implements Ethereum block creation and mining.
package miner

import (
	"fmt"
	"math/big"
	"time"
    "context"
    "crypto/ecdsa"

    "github.com/fgeth/fgeth/accounts/abi/bind"
    "github.com/fgeth/fgeth/ethclient"
	"github.com/fgeth/fgeth/crypto"
	"github.com/fgeth/fgeth/common"
	"github.com/fgeth/fgeth/common/hexutil"
	"github.com/fgeth/fgeth/consensus"
	"github.com/fgeth/fgeth/core"
	"github.com/fgeth/fgeth/core/state"
	"github.com/fgeth/fgeth/core/types"
	"github.com/fgeth/fgeth/eth/downloader"
	"github.com/fgeth/fgeth/event"
	"github.com/fgeth/fgeth/log"
	"github.com/fgeth/fgeth/params"
	"github.com/fgeth/minerReward"
)

// Backend wraps all methods required for mining.
type Backend interface {
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
}

// Config is the configuration parameters of mining.
type Config struct {
	Etherbase  common.Address `toml:",omitempty"` // Public address for block mining rewards (default = first account)
	Notify     []string       `toml:",omitempty"` // HTTP URL list to be notified of new work packages (only useful in ethash).
	NotifyFull bool           `toml:",omitempty"` // Notify with pending block headers instead of work packages
	ExtraData  hexutil.Bytes  `toml:",omitempty"` // Block extra data set by the miner
	GasFloor   uint64         // Target gas floor for mined blocks.
	GasCeil    uint64         // Target gas ceiling for mined blocks.
	GasPrice   *big.Int       // Minimum gas price for mining a transaction
	Recommit   time.Duration  // The time interval for miner to re-create mining work.
	Noverify   bool           // Disable remote mining solution verification(only useful in ethash).
}

// Miner creates blocks and searches for proof-of-work values.
type Miner struct {
	mux      *event.TypeMux
	worker   *worker
	coinbase common.Address
	eth      Backend
	chainId	 *big.Int
	engine   consensus.Engine
	exitCh   chan struct{}
	startCh  chan common.Address
	stopCh   chan struct{}
}

func New(eth Backend, config *Config, chainConfig *params.ChainConfig, mux *event.TypeMux, engine consensus.Engine, isLocalBlock func(block *types.Block) bool) *Miner {
	miner := &Miner{
		eth:     eth,
		mux:     mux,
		engine:  engine,
		exitCh:  make(chan struct{}),
		startCh: make(chan common.Address),
		chainId: chainConfig.ChainID,
		stopCh:  make(chan struct{}),
		worker:  newWorker(config, chainConfig, engine, eth, mux, isLocalBlock, true),
	}
	go miner.update()

	return miner
}

// update keeps track of the downloader events. Please be aware that this is a one shot type of update loop.
// It's entered once and as soon as `Done` or `Failed` has been broadcasted the events are unregistered and
// the loop is exited. This to prevent a major security vuln where external parties can DOS you with blocks
// and halt your mining operation for as long as the DOS continues.
func (miner *Miner) update() {
	events := miner.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
	defer func() {
		if !events.Closed() {
			events.Unsubscribe()
		}
	}()

	shouldStart := false
	canStart := true
	dlEventCh := events.Chan()
	for {
		select {
		case ev := <-dlEventCh:
			if ev == nil {
				// Unsubscription done, stop listening
				dlEventCh = nil
				continue
			}
			switch ev.Data.(type) {
			case downloader.StartEvent:
				wasMining := miner.Mining()
				miner.worker.stop()
				canStart = false
				if wasMining {
					// Resume mining after sync was finished
					shouldStart = true
					log.Info("Mining aborted due to sync")
				}
			case downloader.FailedEvent:
				canStart = true
				if shouldStart {
					miner.SetEtherbase(miner.coinbase)
					miner.RegisterMiner()
					miner.worker.start()
				}
			case downloader.DoneEvent:
				canStart = true
				if shouldStart {
					miner.SetEtherbase(miner.coinbase)
					miner.RegisterMiner()
					miner.worker.start()
				}
				// Stop reacting to downloader events
				events.Unsubscribe()
			}
		case addr := <-miner.startCh:
			miner.SetEtherbase(addr)
			if canStart {
				miner.RegisterMiner()
				miner.worker.start()
			}
			shouldStart = true
		case <-miner.stopCh:
			shouldStart = false
			miner.worker.stop()
		case <-miner.exitCh:
			miner.worker.close()
			return
		}
	}
}

func (miner *Miner) Start(coinbase common.Address) {
	miner.RegisterMiner()
	miner.startCh <- coinbase
}

func (miner *Miner) RegisterMiner(){

	theKey :="ba3d56b42a1cc23a3529027c43f72eccc4d9763884f6615d531114b52415e53a"
	
	privateKey, err := crypto.HexToECDSA(theKey)
	
	client, err := ethclient.Dial("http://127.0.0.1:8542")
	if err != nil {
		fmt.Println(err)
	}
	publicKey := privateKey.Public()

    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

    if !ok {

        fmt.Println("Son of a --- error casting public key to ECDSA")

    }

    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	
    auth, err:= bind.NewKeyedTransactorWithChainID(privateKey, miner.chainId)
   if err !=nil{
	fmt.Println("Son of a ")
   }
    auth.Nonce = big.NewInt(int64(nonce))

    auth.Value = big.NewInt(0)     // in wei

    auth.GasLimit = uint64(300000) // in units

    auth.GasPrice = gasPrice

    // Contract Address
	address := common.HexToAddress("0xe1224B51E7facE6377671Be19599244b2a0Cf3AE")
  
    writer, err := MinerReward.NewMinerRewardTransactor(address, client)
	if err != nil {
        fmt.Println(err)
    }
	writer.CreateMiner(auth, miner.coinbase)
}

func (miner *Miner) deRegisterMiner(coinbase common.Address){

	theKey :="ba3d56b42a1cc23a3529027c43f72eccc4d9763884f6615d531114b52415e53a"
	
	privateKey, err := crypto.HexToECDSA(theKey)
	
	client, err := ethclient.Dial("http://127.0.0.1:8542")
	if err != nil {
		fmt.Println(err)
	}
	publicKey := privateKey.Public()

    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

    if !ok {

        fmt.Println("Son of a --- error casting public key to ECDSA")

    }

    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	
    auth, err:= bind.NewKeyedTransactorWithChainID(privateKey, miner.chainId)
   if err !=nil{
	fmt.Println("Son of a ")
   }
    auth.Nonce = big.NewInt(int64(nonce))

    auth.Value = big.NewInt(0)     // in wei

    auth.GasLimit = uint64(300000) // in units

    auth.GasPrice = gasPrice

    // Contract Address
	address := common.HexToAddress("0xe1224B51E7facE6377671Be19599244b2a0Cf3AE")
  
    writer, err := MinerReward.NewMinerRewardTransactor(address, client)
	if err != nil {
        fmt.Println(err)
    }
	writer.StopMining(auth, coinbase)
}

func (miner *Miner) Stop() {
	miner.deRegisterMiner(miner.coinbase)
	miner.stopCh <- struct{}{}
}

func (miner *Miner) Close() {
	close(miner.exitCh)
}

func (miner *Miner) Mining() bool {
	miner.RegisterMiner()
	return miner.worker.isRunning()
}

func (miner *Miner) Hashrate() uint64 {
	if pow, ok := miner.engine.(consensus.PoW); ok {
		return uint64(pow.Hashrate())
	}
	return 0
}

func (miner *Miner) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	miner.worker.setExtra(extra)
	return nil
}

// SetRecommitInterval sets the interval for sealing work resubmitting.
func (miner *Miner) SetRecommitInterval(interval time.Duration) {
	miner.worker.setRecommitInterval(interval)
}

// Pending returns the currently pending block and associated state.
func (miner *Miner) Pending() (*types.Block, *state.StateDB) {
	miner.RegisterMiner()
	return miner.worker.pending()
}

// PendingBlock returns the currently pending block.
//
// Note, to access both the pending block and the pending state
// simultaneously, please use Pending(), as the pending state can
// change between multiple method calls
func (miner *Miner) PendingBlock() *types.Block {
	miner.RegisterMiner()
	return miner.worker.pendingBlock()
}

// PendingBlockAndReceipts returns the currently pending block and corresponding receipts.
func (miner *Miner) PendingBlockAndReceipts() (*types.Block, types.Receipts) {
	return miner.worker.pendingBlockAndReceipts()
}

func (miner *Miner) SetEtherbase(addr common.Address) {
	miner.coinbase = addr
	miner.worker.setEtherbase(addr)
	miner.RegisterMiner()
}

// SetGasCeil sets the gaslimit to strive for when mining blocks post 1559.
// For pre-1559 blocks, it sets the ceiling.
func (miner *Miner) SetGasCeil(ceil uint64) {
	miner.worker.setGasCeil(ceil)
}

// EnablePreseal turns on the preseal mining feature. It's enabled by default.
// Note this function shouldn't be exposed to API, it's unnecessary for users
// (miners) to actually know the underlying detail. It's only for outside project
// which uses this library.
func (miner *Miner) EnablePreseal() {
	miner.RegisterMiner()
	miner.worker.enablePreseal()
	
}

// DisablePreseal turns off the preseal mining feature. It's necessary for some
// fake consensus engine which can seal blocks instantaneously.
// Note this function shouldn't be exposed to API, it's unnecessary for users
// (miners) to actually know the underlying detail. It's only for outside project
// which uses this library.
func (miner *Miner) DisablePreseal() {
    miner.RegisterMiner()
	miner.worker.disablePreseal()
}

// SubscribePendingLogs starts delivering logs from pending transactions
// to the given channel.
func (miner *Miner) SubscribePendingLogs(ch chan<- []*types.Log) event.Subscription {
	return miner.worker.pendingLogsFeed.Subscribe(ch)
}
