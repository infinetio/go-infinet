// Copyright 2018 The go-infinet Authors
// This file is part of the go-infinet library.
//
// The go-infinet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-infinet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-infinet library. If not, see <http://www.gnu.org/licenses/>.

package protocol

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
	"github.com/juchain/go-juchain/common"
	"github.com/juchain/go-juchain/core"
	"github.com/juchain/go-juchain/core/state"
	"github.com/juchain/go-juchain/core/types"
	"github.com/juchain/go-juchain/common/crypto"
	"github.com/juchain/go-juchain/p2p/protocol/downloader"
	"github.com/juchain/go-juchain/core/store"
	"github.com/juchain/go-juchain/p2p"
	"github.com/juchain/go-juchain/config"
	"github.com/juchain/go-juchain/p2p/discover"
	"os"
	"github.com/juchain/go-juchain/common/log"
	"time"
	"reflect"
	"strings"
	"github.com/juchain/go-juchain/vm/solc/abi"
	"github.com/juchain/go-juchain/consensus/dpos"
	"github.com/juchain/go-juchain/consensus"
	"github.com/juchain/go-juchain/vm/solc"
	"crypto/ecdsa"
	"github.com/juchain/go-juchain/common/event"
)

// Tests that protocol versions and modes of operations are matched up properly.
func TestProtocolCompatibility(t *testing.T) {
	// Define the compatibility chart
	tests := []struct {
		version    uint
		mode       downloader.SyncMode
		compatible bool
	}{
		{00, downloader.FullSync, true}, {OBOD01, downloader.FullSync, true},
	}
	// Make sure anything we screw up is restored
	backup := ProtocolVersions
	defer func() { ProtocolVersions = backup }()

	// Try all available compatibility configs and check for errors
	for i, tt := range tests {
		ProtocolVersions = []uint{tt.version}

		pm, _, err := newTestProtocolManager(tt.mode, 0, nil, nil, true)
		if pm != nil {
			defer pm.Stop()
		}
		if (err == nil && !tt.compatible) || (err != nil && tt.compatible) {
			t.Errorf("test %d: compatibility mismatch: have error %v, want compatibility %v", i, err, tt.compatible)
		}
	}
}

// Tests that block headers can be retrieved from a remote chain based on user queries.
func TestGetBlockHeaders(t *testing.T) { testGetBlockHeaders(t, OBOD01) }

func testGetBlockHeaders(t *testing.T, protocol uint) {
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxHashFetch+15, nil, nil, true)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()

	// Create a "random" unknown hash for testing
	var unknown common.Hash
	for i := range unknown {
		unknown[i] = byte(i)
	}
	// Create a batch of tests for various scenarios
	limit := uint64(downloader.MaxHeaderFetch)
	tests := []struct {
		query  *getBlockHeadersData // The query to execute for header retrieval
		expect []common.Hash        // The hashes of the block whose headers are expected
	}{
		// A single random block should be retrievable by hash and number too
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: pm.blockchain.GetBlockByNumber(limit / 2).Hash()}, Amount: 1},
			[]common.Hash{pm.blockchain.GetBlockByNumber(limit / 2).Hash()},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: limit / 2}, Amount: 1},
			[]common.Hash{pm.blockchain.GetBlockByNumber(limit / 2).Hash()},
		},
		// Multiple headers should be retrievable in both directions
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: limit / 2}, Amount: 3},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(limit / 2).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 + 1).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 + 2).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: limit / 2}, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(limit / 2).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 - 1).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 - 2).Hash(),
			},
		},
		// Multiple headers with skip lists should be retrievable
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: limit / 2}, Skip: 3, Amount: 3},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(limit / 2).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 + 4).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 + 8).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: limit / 2}, Skip: 3, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(limit / 2).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 - 4).Hash(),
				pm.blockchain.GetBlockByNumber(limit/2 - 8).Hash(),
			},
		},
		// The chain endpoints should be retrievable
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: 0}, Amount: 1},
			[]common.Hash{pm.blockchain.GetBlockByNumber(0).Hash()},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: pm.blockchain.CurrentBlock().NumberU64()}, Amount: 1},
			[]common.Hash{pm.blockchain.CurrentBlock().Hash()},
		},
		// Ensure protocol limits are honored
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: pm.blockchain.CurrentBlock().NumberU64() - 1}, Amount: limit + 10, Reverse: true},
			pm.blockchain.GetBlockHashesFromHash(pm.blockchain.CurrentBlock().Hash(), limit),
		},
		// Check that requesting more than available is handled gracefully
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: pm.blockchain.CurrentBlock().NumberU64() - 4}, Skip: 3, Amount: 3},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(pm.blockchain.CurrentBlock().NumberU64() - 4).Hash(),
				pm.blockchain.GetBlockByNumber(pm.blockchain.CurrentBlock().NumberU64()).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: 4}, Skip: 3, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(4).Hash(),
				pm.blockchain.GetBlockByNumber(0).Hash(),
			},
		},
		// Check that requesting more than available is handled gracefully, even if mid skip
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: pm.blockchain.CurrentBlock().NumberU64() - 4}, Skip: 2, Amount: 3},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(pm.blockchain.CurrentBlock().NumberU64() - 4).Hash(),
				pm.blockchain.GetBlockByNumber(pm.blockchain.CurrentBlock().NumberU64() - 1).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: 4}, Skip: 2, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(4).Hash(),
				pm.blockchain.GetBlockByNumber(1).Hash(),
			},
		},
		// Check a corner case where requesting more can iterate past the endpoints
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: 2}, Amount: 5, Reverse: true},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(2).Hash(),
				pm.blockchain.GetBlockByNumber(1).Hash(),
				pm.blockchain.GetBlockByNumber(0).Hash(),
			},
		},
		// Check a corner case where skipping overflow loops back into the chain start
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: pm.blockchain.GetBlockByNumber(3).Hash()}, Amount: 2, Reverse: false, Skip: math.MaxUint64 - 1},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(3).Hash(),
			},
		},
		// Check a corner case where skipping overflow loops back to the same header
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: pm.blockchain.GetBlockByNumber(1).Hash()}, Amount: 2, Reverse: false, Skip: math.MaxUint64},
			[]common.Hash{
				pm.blockchain.GetBlockByNumber(1).Hash(),
			},
		},
		// Check that non existing headers aren't returned
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: unknown}, Amount: 1},
			[]common.Hash{},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: pm.blockchain.CurrentBlock().NumberU64() + 1}, Amount: 1},
			[]common.Hash{},
		},
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the headers to expect in the response
		headers := []*types.Header{}
		for _, hash := range tt.expect {
			headers = append(headers, pm.blockchain.GetBlockByHash(hash).Header())
		}
		// Send the hash request and verify the response
		p2p.Send(peer.app, 0x03, tt.query)
		if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
			t.Errorf("test %d: headers mismatch: %v", i, err)
		}
		// If the test used number origins, repeat with hashes as the too
		if tt.query.Origin.Hash == (common.Hash{}) {
			if origin := pm.blockchain.GetBlockByNumber(tt.query.Origin.Number); origin != nil {
				tt.query.Origin.Hash, tt.query.Origin.Number = origin.Hash(), 0

				p2p.Send(peer.app, 0x03, tt.query)
				if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
					t.Errorf("test %d: headers mismatch: %v", i, err)
				}
			}
		}
	}
}

// Tests that block contents can be retrieved from a remote chain based on their hashes.
func TestGetBlockBodies(t *testing.T) { testGetBlockBodies(t, OBOD01) }

func testGetBlockBodies(t *testing.T, protocol uint) {
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxBlockFetch+15, nil, nil, true)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()
	defer pm.Stop()
	// Create a batch of tests for various scenarios
	limit := downloader.MaxBlockFetch
	tests := []struct {
		random    int           // Number of blocks to fetch randomly from the chain
		explicit  []common.Hash // Explicitly requested blocks
		available []bool        // Availability of explicitly requested blocks
		expected  int           // Total number of existing blocks to expect
	}{
		{1, nil, nil, 1},                                                         // A single random block should be retrievable
		{10, nil, nil, 10},                                                       // Multiple random blocks should be retrievable
		{limit, nil, nil, limit},                                                 // The maximum possible blocks should be retrievable
		{limit + 1, nil, nil, limit},                                             // No more than the possible block count should be returned
		{0, []common.Hash{pm.blockchain.Genesis().Hash()}, []bool{true}, 1},      // The genesis block should be retrievable
		{0, []common.Hash{pm.blockchain.CurrentBlock().Hash()}, []bool{true}, 1}, // The chains head block should be retrievable
		{0, []common.Hash{{}}, []bool{false}, 0},                                 // A non existent block should not be returned

		// Existing and non-existing blocks interleaved should not cause problems
		{0, []common.Hash{
			{},
			pm.blockchain.GetBlockByNumber(1).Hash(),
			{},
			pm.blockchain.GetBlockByNumber(10).Hash(),
			{},
			pm.blockchain.GetBlockByNumber(100).Hash(),
			{},
		}, []bool{false, true, false, true, false, true, false}, 3},
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the hashes to request, and the response to expect
		hashes, seen := []common.Hash{}, make(map[int64]bool)
		bodies := []*blockBody{}

		for j := 0; j < tt.random; j++ {
			for {
				num := rand.Int63n(int64(pm.blockchain.CurrentBlock().NumberU64()))
				if !seen[num] {
					seen[num] = true

					block := pm.blockchain.GetBlockByNumber(uint64(num))
					hashes = append(hashes, block.Hash())
					if len(bodies) < tt.expected {
						bodies = append(bodies, &blockBody{Transactions: block.Transactions(), Uncles: block.Uncles()})
					}
					break
				}
			}
		}
		for j, hash := range tt.explicit {
			hashes = append(hashes, hash)
			if tt.available[j] && len(bodies) < tt.expected {
				block := pm.blockchain.GetBlockByHash(hash)
				bodies = append(bodies, &blockBody{Transactions: block.Transactions(), Uncles: block.Uncles()})
			}
		}
		// Send the hash request and verify the response
		p2p.Send(peer.app, 0x05, hashes)
		if err := p2p.ExpectMsg(peer.app, 0x06, bodies); err != nil {
			t.Errorf("test %d: bodies mismatch: %v", i, err)
		}
	}
}

// Tests that the node state database can be retrieved based on hashes.
func TestGetNodeData(t *testing.T) { testGetNodeData(t, OBOD01) }

func testGetNodeData(t *testing.T, protocol uint) {
	// Define three accounts to simulate transactions with
	acc1Key, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	acc2Key, _ := crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	acc1Addr := crypto.PubkeyToAddress(acc1Key.PublicKey)
	acc2Addr := crypto.PubkeyToAddress(acc2Key.PublicKey)

	signer := types.HomesteadSigner{}
	// Create a chain generator with some simple transactions (blatantly stolen from @fjl/chain_markets_test)
	generator := func(i int, block *core.BlockGen) {
		switch i {
		case 0:
			// In block 1, the test bank sends account #1 some ether.
			tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(10000), config.TxGas, nil, nil), signer, testBankKey)
			block.AddTx(tx)
		case 1:
			// In block 2, the test bank sends some more ether to account #1.
			// acc1Addr passes it on to account #2.
			tx1, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(1000), config.TxGas, nil, nil), signer, testBankKey)
			tx2, _ := types.SignTx(types.NewTransaction(block.TxNonce(acc1Addr), acc2Addr, big.NewInt(1000), config.TxGas, nil, nil), signer, acc1Key)
			block.AddTx(tx1)
			block.AddTx(tx2)
		case 2:
			// Block 3 is empty but was mined by account #2.
			block.SetCoinbase(acc2Addr)
			block.SetExtra([]byte("yeehaw"))
		case 3:
			// Block 4 includes blocks 2 and 3 as uncle headers (with modified extra data).
			b2 := block.PrevBlock(1).Header()
			b2.Extra = []byte("foo")
			block.AddUncle(b2)
			b3 := block.PrevBlock(2).Header()
			b3.Extra = []byte("foo")
			block.AddUncle(b3)
		}
	}
	// Assemble the test environment
	pm, db := newTestProtocolManagerMust(t, downloader.FullSync, 4, generator, nil, true)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()
	defer pm.Stop()

	// Fetch for now the entire chain store
	hashes := []common.Hash{}
	for _, key := range db.Keys() {
		if len(key) == len(common.Hash{}) {
			hashes = append(hashes, common.BytesToHash(key))
		}
	}
	p2p.Send(peer.app, 0x0d, hashes)
	msg, err := peer.app.ReadMsg()
	if err != nil {
		t.Fatalf("failed to read node data response: %v", err)
	}
	if msg.Code != 0x0e {
		t.Fatalf("response packet code mismatch: have %x, want %x", msg.Code, 0x0c)
	}
	var data [][]byte
	if err := msg.Decode(&data); err != nil {
		t.Fatalf("failed to decode response node data: %v", err)
	}
	// Verify that all hashes correspond to the requested data, and reconstruct a state tree
	for i, want := range hashes {
		if hash := crypto.Keccak256Hash(data[i]); hash != want {
			t.Errorf("data hash mismatch: have %x, want %x", hash, want)
		}
	}
	statedb, _ := store.NewMemDatabase()
	for i := 0; i < len(data); i++ {
		statedb.Put(hashes[i].Bytes(), data[i])
	}
	accounts := []common.Address{testBank, acc1Addr, acc2Addr}
	for i := uint64(0); i <= pm.blockchain.CurrentBlock().NumberU64(); i++ {
		trie, _ := state.New(pm.blockchain.GetBlockByNumber(i).Root(), state.NewDatabase(statedb))

		for j, acc := range accounts {
			state, _ := pm.blockchain.State()
			bw := state.GetBalance(acc)
			bh := trie.GetBalance(acc)

			if (bw != nil && bh == nil) || (bw == nil && bh != nil) {
				t.Errorf("test %d, account %d: balance mismatch: have %v, want %v", i, j, bh, bw)
			}
			if bw != nil && bh != nil && bw.Cmp(bw) != 0 {
				t.Errorf("test %d, account %d: balance mismatch: have %v, want %v", i, j, bh, bw)
			}
		}
	}
}

// Tests that the transaction receipts can be retrieved based on hashes.
func TestGetReceipt(t *testing.T) { testGetReceipt(t, OBOD01) }

func testGetReceipt(t *testing.T, protocol uint) {
	// Define three accounts to simulate transactions with
	acc1Key, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	acc2Key, _ := crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	acc1Addr := crypto.PubkeyToAddress(acc1Key.PublicKey)
	acc2Addr := crypto.PubkeyToAddress(acc2Key.PublicKey)

	signer := types.HomesteadSigner{}
	// Create a chain generator with some simple transactions (blatantly stolen from @fjl/chain_markets_test)
	generator := func(i int, block *core.BlockGen) {
		switch i {
		case 0:
			// In block 1, the test bank sends account #1 some ether.
			tx, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(10000), config.TxGas, nil, nil), signer, testBankKey)
			block.AddTx(tx)
		case 1:
			// In block 2, the test bank sends some more ether to account #1.
			// acc1Addr passes it on to account #2.
			tx1, _ := types.SignTx(types.NewTransaction(block.TxNonce(testBank), acc1Addr, big.NewInt(1000), config.TxGas, nil, nil), signer, testBankKey)
			tx2, _ := types.SignTx(types.NewTransaction(block.TxNonce(acc1Addr), acc2Addr, big.NewInt(1000), config.TxGas, nil, nil), signer, acc1Key)
			block.AddTx(tx1)
			block.AddTx(tx2)
		case 2:
			// Block 3 is empty but was mined by account #2.
			block.SetCoinbase(acc2Addr)
			block.SetExtra([]byte("yeehaw"))
		case 3:
			// Block 4 includes blocks 2 and 3 as uncle headers (with modified extra data).
			b2 := block.PrevBlock(1).Header()
			b2.Extra = []byte("foo")
			block.AddUncle(b2)
			b3 := block.PrevBlock(2).Header()
			b3.Extra = []byte("foo")
			block.AddUncle(b3)
		}
	}
	// Assemble the test environment
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, 4, generator, nil, true)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()
	defer pm.Stop()

	// Collect the hashes to request, and the response to expect
	hashes, receipts := []common.Hash{}, []types.Receipts{}
	for i := uint64(0); i <= pm.blockchain.CurrentBlock().NumberU64(); i++ {
		block := pm.blockchain.GetBlockByNumber(i)

		hashes = append(hashes, block.Hash())
		receipts = append(receipts, pm.blockchain.GetReceiptsByHash(block.Hash()))
	}
	// Send the hash request and verify the response
	p2p.Send(peer.app, 0x0f, hashes)
	if err := p2p.ExpectMsg(peer.app, 0x10, receipts); err != nil {
		t.Errorf("receipts mismatch: %v", err)
	}

}


type DelegatorVotingManagerImpl struct{}
func (d *DelegatorVotingManagerImpl) Refresh() (delegatorsTable []string, delegatorNodes []*discover.Node) {
	return []string{}, []*discover.Node{}
}

// Tests that the node state database can be retrieved based on hashes.
func TestVoteElection(t *testing.T) { testVoteElection(t, OBOD01) }

func testVoteElection(t *testing.T, protocol uint) {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
	TestMode = true;
	generator := func(i int, block *core.BlockGen) {}
	// Assemble the testing environment
	pm, _   := newTestProtocolManagerMust(t, downloader.FullSync, 4, generator, nil, false)
	peer, _  := newTestPeer("peer", protocol, pm, true)
	peer1, _  := newTestPeer("peer1", protocol, pm, true)
	defer peer1.close()
	defer peer.close()
	defer pm.Stop();

	NodeAIdHash := common.Hex2Bytes("aaaaa111");
	NodeBIdHash := common.Hex2Bytes("bbbbb111");

	pm.dposManager.scheduleElecting()
	activeTime := NextElectionInfo.activeTime;
	//expects I win. simply skip this request
	p2p.Send(peer.app, VOTE_ElectionNode_Request, &VoteElectionRequest{1,
		100, activeTime, currNodeIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}
	if NextElectionInfo.activeTime != activeTime {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	//expects agreed the request node as the election node
	p2p.Send(peer.app, VOTE_ElectionNode_Request, &VoteElectionRequest{1,
		2, activeTime, currNodeIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.round != 1 {
		t.Errorf("returned %v want     %v", NextElectionInfo.round, 2)
	}
	t.Logf("electionTickets returned %v", NextElectionInfo.electionTickets)

	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}
	if NextElectionInfo.activeTime != activeTime {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	//I am in agreed state already.
	p2p.Send(peer1.app, VOTE_ElectionNode_Request, &VoteElectionRequest{1,
		2, activeTime, NodeAIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.round != 1 {
		t.Errorf("returned %v want     %v", NextElectionInfo.round, 2)
	}
	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}
	if NextElectionInfo.activeTime != activeTime {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	//Mismatched request.round with less value
	p2p.Send(peer1.app, VOTE_ElectionNode_Request, &VoteElectionRequest{0,
		2, activeTime, NodeAIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.round != 1 {
		t.Errorf("returned %v want     %v", NextElectionInfo.round, 2)
	}
	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}
	if NextElectionInfo.activeTime != activeTime {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	//Mismatched request.round with greater value
	p2p.Send(peer1.app, VOTE_ElectionNode_Request, &VoteElectionRequest{2,
		2, activeTime, currNodeIdHash})
	p2p.Send(peer1.app, VOTE_ElectionNode_Request, &VoteElectionRequest{NextElectionInfo.round - 10,
		2, activeTime, currNodeIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.round != 1 {
		t.Errorf("returned %v want     %v", NextElectionInfo.round, 2)
	}
	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}
	if NextElectionInfo.activeTime != activeTime {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	//Voted Election Response must not have VOTESTATE_SELECTED state. rejected!
	p2p.Send(peer.app, VOTE_ElectionNode_Response, &VoteElectionResponse{1,
		2, NextElectionInfo.activeTime,
		VOTESTATE_SELECTED,currNodeIdHash})
	p2p.Send(peer.app, VOTE_ElectionNode_Response, &VoteElectionResponse{1,
		2, NextElectionInfo.activeTime,
		VOTESTATE_MISMATCHED_ROUND,currNodeIdHash})

	//Confirmed the final election node:
	p2p.Send(peer.app, VOTE_ElectionNode_Broadcast, &BroadcastVotedElection{1,
		2, activeTime, VOTESTATE_MISMATCHED_ROUND,currNodeIdHash})
	p2p.Send(peer.app, VOTE_ElectionNode_Broadcast, &BroadcastVotedElection{1,
		2, activeTime, VOTESTATE_MISMATCHED_ROUND,NodeAIdHash})
	p2p.Send(peer.app, VOTE_ElectionNode_Broadcast, &BroadcastVotedElection{1,
		2, activeTime, VOTESTATE_MISMATCHED_ROUND,NodeAIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}
	if NextElectionInfo.activeTime != activeTime {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	// switch to next round.
	pm.dposManager.scheduleElecting()
	if NextElectionInfo.round != 2 {
		t.Errorf("returned %v want     %v", NextElectionInfo.round, 2)
	}
	if NextElectionInfo.enodestate != VOTESTATE_LOOKING {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}

	//Mismatched request.round
	p2p.Send(peer.app, VOTE_ElectionNode_Request, &VoteElectionRequest{1,
		NextElectionInfo.electionTickets, activeTime, currNodeIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.enodestate != VOTESTATE_LOOKING {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}
	p2p.Send(peer.app, VOTE_ElectionNode_Request, &VoteElectionRequest{3,
		NextElectionInfo.electionTickets, activeTime,currNodeIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.enodestate != VOTESTATE_LOOKING {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_LOOKING)
	}
	p2p.Send(peer.app, VOTE_ElectionNode_Broadcast, &BroadcastVotedElection{2,
		2, activeTime, VOTESTATE_MISMATCHED_ROUND,currNodeIdHash})
	p2p.Send(peer.app, VOTE_ElectionNode_Broadcast, &BroadcastVotedElection{2,
		3, activeTime, VOTESTATE_MISMATCHED_ROUND,NodeBIdHash})
	p2p.Send(peer.app, VOTE_ElectionNode_Broadcast, &BroadcastVotedElection{2,
		4, activeTime, VOTESTATE_MISMATCHED_ROUND,NodeBIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextElectionInfo.enodestate != VOTESTATE_SELECTED {
		t.Errorf("returned %v want     %v", NextElectionInfo.enodestate, VOTESTATE_SELECTED)
	}

	pm.dposManager.scheduleElecting()
	TestMode = false
}

func TestDPosDelegator(t *testing.T) {
	//log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	DelegatorsTable = []string{"abcd"}
	TestMode = true;
	generator := func(i int, block *core.BlockGen) {}
	// Assemble the testing environment
	pm, _   := newTestProtocolManagerMust(t, downloader.FullSync, 4, generator, nil, false)
	peer, _  := newTestPeer("peer", OBOD01, pm, true)
	peer1, _  := newTestPeer("peer1", OBOD01, pm, true)
	defer peer1.close()
	defer peer.close()
	defer pm.Stop();

	NodeAIdHash := common.Hex2Bytes("aaaaa111");
	pm.dposManager.dposManager.syncDelegatedNodeSafely()

	p2p.Send(peer.app, SYNC_BIGPERIOD_REQUEST, &SyncBigPeriodRequest{NextGigPeriodInstance.round,
		NextGigPeriodInstance.activeTime,
		NextGigPeriodInstance.delegatedNodes,
		NextGigPeriodInstance.delegatedNodesSign,
		currNodeIdHash})
	p2p.Send(peer1.app, SYNC_BIGPERIOD_REQUEST, &SyncBigPeriodRequest{NextGigPeriodInstance.round,
		NextGigPeriodInstance.activeTime,
		NextGigPeriodInstance.delegatedNodes,
		NextGigPeriodInstance.delegatedNodesSign,
		NodeAIdHash})
	// mismatch and overflow
	p2p.Send(peer1.app, SYNC_BIGPERIOD_REQUEST, &SyncBigPeriodRequest{NextGigPeriodInstance.round-10,
		NextGigPeriodInstance.activeTime,
		NextGigPeriodInstance.delegatedNodes,
		NextGigPeriodInstance.delegatedNodesSign,
		NodeAIdHash})
	p2p.Send(peer1.app, SYNC_BIGPERIOD_REQUEST, &SyncBigPeriodRequest{NextGigPeriodInstance.round+10,
		NextGigPeriodInstance.activeTime,
		NextGigPeriodInstance.delegatedNodes,
		NextGigPeriodInstance.delegatedNodesSign,
		NodeAIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextGigPeriodInstance.state != STATE_CONFIRMED {
		t.Errorf("returned %v want     %v", NextGigPeriodInstance.state, STATE_CONFIRMED)
	}
	if len(NextGigPeriodInstance.delegatedNodes) == 0 {
		t.Errorf("returned %v want     %v", NextGigPeriodInstance.delegatedNodes, 1)
	}

	delegatedNodesA := []string{"aaa", "bbbb", "cccc", "ddddd", "eeeeee"}
	p2p.Send(peer1.app, SYNC_BIGPERIOD_RESPONSE, &SyncBigPeriodResponse{NextGigPeriodInstance.round,
		NextGigPeriodInstance.activeTime,
		delegatedNodesA,
		SignCandidates(delegatedNodesA),
		STATE_CONFIRMED,
		NodeAIdHash})
	p2p.Send(peer1.app, SYNC_BIGPERIOD_RESPONSE, &SyncBigPeriodResponse{NextGigPeriodInstance.round,
		NextGigPeriodInstance.activeTime,
		delegatedNodesA,
		SignCandidates(delegatedNodesA),
		STATE_CONFIRMED,
		NodeAIdHash})
	time.Sleep(time.Millisecond * time.Duration(500))
	if NextGigPeriodInstance.state != STATE_CONFIRMED {
		t.Errorf("returned %v want     %v", NextGigPeriodInstance.state, STATE_CONFIRMED)
	}
	if !reflect.DeepEqual(NextGigPeriodInstance.delegatedNodes, delegatedNodesA) {
		t.Errorf("returned %v want     %v", NextGigPeriodInstance.delegatedNodes, delegatedNodesA)
	}
	TestMode = false
}

func TestDPosDelegatorContract(t *testing.T) {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	TestMode = true
	generator := func(i int, block *core.BlockGen) {}
	// Assemble the testing environment
	pm, _   := newTestProtocolManagerMust(t, downloader.FullSync, 1, generator, nil, false)
	defer pm.Stop();

	dappabi, err := abi.JSON(strings.NewReader(core.DPOSBallotABI))
	if err != nil {
		log.Error("Unable to load DPoS Ballot ABI object!")
		return;
	}
	VotingAccessor = &DelegatorAccessorImpl{dappabi: dappabi, blockchain: pm.blockchain, b: pm.backend};
	DelegatorsTable, DelegatorNodeInfo, err = VotingAccessor.Refresh();
	if err != nil {
		t.Error(err.Error())
	}

	TestMode = false
}

func TestPackageBlock(t *testing.T) {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	generator := func(i int, block *core.BlockGen) {}
	// Assemble the testing environment
	pm, _   := newTestProtocolManagerMust(t, downloader.FullSync, 1, generator, nil, false)
	defer pm.Stop();

	ElectionInfo0 = &ElectionInfo{electionNodeId: currNodeId}
	ElectionInfo0.electionNodeId = currNodeId

	for i :=0; i < 100; i++ {
		pm.dposManager.schedulePackaging()
	}


}

func TestMultipleDAppChainsInsert(t *testing.T) {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	key, _  := crypto.GenerateKey()
	key2, _ := crypto.GenerateKey()
	dappIdA := crypto.PubkeyToAddress(key.PublicKey)
	dappIdB := crypto.PubkeyToAddress(key2.PublicKey)

	gspec   := core.Genesis{
		Config: config.TestChainConfig,
		Alloc: core.GenesisAlloc{
			dappIdA: {Balance: new(big.Int).SetUint64(2 * config.Ether)},
			dappIdB: {Balance: new(big.Int).SetUint64(2 * config.Ether)},
		},
		GasLimit: 100e6, // 100 M
	}
	engine := consensus.CreateFakeEngine()
	db, _ := store.NewMemDatabase()
	gspec.MustCommit(db)
	db1, _ := store.NewMemDatabase()
	gspec.MustCommit(db1)
	db2, _ := store.NewMemDatabase()
	gspec.MustCommit(db2)

	chain, err  := core.NewBlockChain(db, nil, config.TestChainConfig, engine, vm.Config{})
	if err != nil {
		t.Fatalf("failed to create main chain: %v", err)
	}
	chain1, err := core.NewBlockChain(db1, nil, config.TestChainConfig, engine, vm.Config{})
	if err != nil {
		t.Fatalf("failed to create dapp1 chain: %v", err)
	}
	chain2, err := core.NewBlockChain(db2, nil, config.TestChainConfig, engine, vm.Config{})
	if err != nil {
		t.Fatalf("failed to create dapp2 chain: %v", err)
	}
	dappChains := map[common.Address]*core.BlockChain{
		dappIdA: chain1,
		dappIdB: chain2,
	}
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	//statedb.SetBalance(dappIdA, new(big.Int).SetUint64(config.Ether))
	//statedb.SetBalance(dappIdB, new(big.Int).SetUint64(config.Ether))

	pool := core.NewTxPool(core.DefaultTxPoolConfig, config.TestChainConfig, chain, dappChains)
	defer pool.Stop()

	packager := dpos.NewPackager1(config.TestChainConfig, engine, dappIdA, chain, pool, &event.TypeMux{})
	defer packager.Stop()

	tx0 := dappTransaction(&dappIdA, 0, 100000, key)
	tx1 := dappTransaction(&dappIdA, 1, 100000, key)
	tx2 := dappTransaction(&dappIdA, 2, 100000, key)
	tx3 := dappTransaction(&dappIdA, 3, 100000, key)
	if (tx0 == nil) {
		t.Error("failed to create dapp tx.")
		return;
	}
	pool.AddLocals(types.Transactions{tx0, tx1, tx2, tx3})

	packager.GenerateNewBlock(1, "testnode1")

	t.Logf("chain.CurrentBlock().Number()= %v", chain.CurrentBlock().Number())
	t.Logf("chain1.CurrentBlock().Number()= %v", chain1.CurrentBlock().Number())
	t.Logf("chain2.CurrentBlock().Number()= %v", chain2.CurrentBlock().Number())
	t.Logf("db.Len() = %v", db.Len())
	t.Logf("db1.Len() = %v", db1.Len())
	t.Logf("db2.Len() = %v", db2.Len())

}
var dappTxData = []byte{1,2,3,4,5,6,7,8,10,4,5,6,7,8,10,1,2,3,4,5,6,7,8,10,4,5,6,7,8,10,1,2,3,4,5,6,7,8,10,4,5,6,7,8,10,1,2,3,4,5,6,7,8,10,4,5,6,7,8,101,2,3,4,5,6,7,8,10,4,5,6,7,8,10};
func dappTransaction(dapp *common.Address, nonce uint64, gaslimit uint64, key *ecdsa.PrivateKey) *types.Transaction {
	return pricedDappTransaction(dapp, nonce, gaslimit, big.NewInt(1), key)
}

func pricedDappTransaction(dapp *common.Address, nonce uint64, gaslimit uint64, gasprice *big.Int, key *ecdsa.PrivateKey) *types.Transaction {
	tx := types.NewDAppTransaction(dapp, nonce, gaslimit, gasprice, dappTxData)
	tx, _ = types.SignTx(tx, types.NewEIP155Signer(common.Big1), key)

	from, _ := types.Sender(types.NewEIP155Signer(common.Big1), tx)
	from1, _ := types.Sender(types.NewEIP155Signer(common.Big1), tx.DAppTx())
	from2, _ := types.Sender(types.NewEIP155Signer(common.Big2), tx.DAppTx())
	if from != from1 || from1 == from2 {
		return nil; //errors.New("signed error.")
	}
	return tx
}