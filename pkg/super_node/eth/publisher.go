// VulcanizeDB
// Copyright © 2019 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/vulcanize/eth-block-extractor/pkg/ipfs"
	"github.com/vulcanize/eth-block-extractor/pkg/ipfs/eth_block_header"
	"github.com/vulcanize/eth-block-extractor/pkg/ipfs/eth_block_receipts"
	"github.com/vulcanize/eth-block-extractor/pkg/ipfs/eth_block_transactions"
	"github.com/vulcanize/eth-block-extractor/pkg/ipfs/eth_state_trie"
	"github.com/vulcanize/eth-block-extractor/pkg/ipfs/eth_storage_trie"
	rlp2 "github.com/vulcanize/eth-block-extractor/pkg/wrappers/rlp"
)

// IPLDPublisher satisfies the IPLDPublisher for ethereum
type IPLDPublisher struct {
	HeaderPutter      ipfs.DagPutter
	TransactionPutter ipfs.DagPutter
	ReceiptPutter     ipfs.DagPutter
	StatePutter       ipfs.DagPutter
	StoragePutter     ipfs.DagPutter
}

// NewIPLDPublisher creates a pointer to a new Publisher which satisfies the IPLDPublisher interface
func NewIPLDPublisher(ipfsPath string) (*IPLDPublisher, error) {
	node, err := ipfs.InitIPFSNode(ipfsPath)
	if err != nil {
		return nil, err
	}
	return &IPLDPublisher{
		HeaderPutter:      eth_block_header.NewBlockHeaderDagPutter(node, rlp2.RlpDecoder{}),
		TransactionPutter: eth_block_transactions.NewBlockTransactionsDagPutter(node),
		ReceiptPutter:     eth_block_receipts.NewEthBlockReceiptDagPutter(node),
		StatePutter:       eth_state_trie.NewStateTrieDagPutter(node),
		StoragePutter:     eth_storage_trie.NewStorageTrieDagPutter(node),
	}, nil
}

// Publish publishes an IPLDPayload to IPFS and returns the corresponding CIDPayload
func (pub *IPLDPublisher) Publish(payload interface{}) (interface{}, error) {
	ipldPayload, ok := payload.(*IPLDPayload)
	if !ok {
		return nil, fmt.Errorf("eth publisher expected payload type %T got %T", &IPLDPayload{}, payload)
	}
	// Process and publish headers
	headerCid, err := pub.publishHeader(ipldPayload.HeaderRLP)
	if err != nil {
		return nil, err
	}
	header := HeaderModel{
		CID:             headerCid,
		ParentHash:      ipldPayload.Block.ParentHash().String(),
		BlockNumber:     ipldPayload.Block.Number().String(),
		BlockHash:       ipldPayload.Block.Hash().String(),
		TotalDifficulty: ipldPayload.TotalDifficulty.String(),
	}

	// Process and publish uncles
	uncleCids := make([]UncleModel, 0, len(ipldPayload.Block.Uncles()))
	for _, uncle := range ipldPayload.Block.Uncles() {
		uncleRlp, err := rlp.EncodeToBytes(uncle)
		if err != nil {
			return nil, err
		}
		uncleCid, err := pub.publishHeader(uncleRlp)
		if err != nil {
			return nil, err
		}
		uncleCids = append(uncleCids, UncleModel{
			CID:        uncleCid,
			ParentHash: uncle.ParentHash.String(),
			BlockHash:  uncle.Hash().String(),
		})
	}

	// Process and publish transactions
	transactionCids, err := pub.publishTransactions(ipldPayload.Block.Body(), ipldPayload.TxMetaData)
	if err != nil {
		return nil, err
	}

	// Process and publish receipts
	receiptsCids, err := pub.publishReceipts(ipldPayload.Receipts, ipldPayload.ReceiptMetaData)
	if err != nil {
		return nil, err
	}

	// Process and publish state leafs
	stateNodeCids, err := pub.publishStateNodes(ipldPayload.StateNodes)
	if err != nil {
		return nil, err
	}

	// Process and publish storage leafs
	storageNodeCids, err := pub.publishStorageNodes(ipldPayload.StorageNodes)
	if err != nil {
		return nil, err
	}

	// Package CIDs and their metadata into a single struct
	return &CIDPayload{
		HeaderCID:       header,
		UncleCIDs:       uncleCids,
		TransactionCIDs: transactionCids,
		ReceiptCIDs:     receiptsCids,
		StateNodeCIDs:   stateNodeCids,
		StorageNodeCIDs: storageNodeCids,
	}, nil
}

func (pub *IPLDPublisher) publishHeader(headerRLP []byte) (string, error) {
	headerCids, err := pub.HeaderPutter.DagPut(headerRLP)
	if err != nil {
		return "", err
	}
	if len(headerCids) != 1 {
		return "", errors.New("single CID expected to be returned for header")
	}
	return headerCids[0], nil
}

func (pub *IPLDPublisher) publishTransactions(blockBody *types.Body, trxMeta []TxModel) ([]TxModel, error) {
	transactionCids, err := pub.TransactionPutter.DagPut(blockBody)
	if err != nil {
		return nil, err
	}
	if len(transactionCids) != len(trxMeta) {
		return nil, errors.New("expected one CID for each transaction")
	}
	mappedTrxCids := make([]TxModel, len(transactionCids))
	for i, cid := range transactionCids {
		mappedTrxCids[i] = TxModel{
			CID:    cid,
			Index:  trxMeta[i].Index,
			TxHash: trxMeta[i].TxHash,
			Src:    trxMeta[i].Src,
			Dst:    trxMeta[i].Dst,
		}
	}
	return mappedTrxCids, nil
}

func (pub *IPLDPublisher) publishReceipts(receipts types.Receipts, receiptMeta []ReceiptModel) (map[common.Hash]ReceiptModel, error) {
	receiptsCids, err := pub.ReceiptPutter.DagPut(receipts)
	if err != nil {
		return nil, err
	}
	if len(receiptsCids) != len(receipts) {
		return nil, errors.New("expected one CID for each receipt")
	}
	// Map receipt cids to their transaction hashes
	mappedRctCids := make(map[common.Hash]ReceiptModel, len(receiptsCids))
	for i, rct := range receipts {
		mappedRctCids[rct.TxHash] = ReceiptModel{
			CID:      receiptsCids[i],
			Contract: receiptMeta[i].Contract,
			Topic0s:  receiptMeta[i].Topic0s,
			Topic1s:  receiptMeta[i].Topic1s,
			Topic2s:  receiptMeta[i].Topic2s,
			Topic3s:  receiptMeta[i].Topic3s,
		}
	}
	return mappedRctCids, nil
}

func (pub *IPLDPublisher) publishStateNodes(stateNodes []TrieNode) ([]StateNodeModel, error) {
	stateNodeCids := make([]StateNodeModel, 0, len(stateNodes))
	for _, node := range stateNodes {
		stateNodeCid, err := pub.StatePutter.DagPut(node.Value)
		if err != nil {
			return nil, err
		}
		if len(stateNodeCid) != 1 {
			return nil, errors.New("single CID expected to be returned for state leaf")
		}
		stateNodeCids = append(stateNodeCids, StateNodeModel{
			StateKey: node.Key.String(),
			CID:      stateNodeCid[0],
			Leaf:     node.Leaf,
		})
	}
	return stateNodeCids, nil
}

func (pub *IPLDPublisher) publishStorageNodes(storageNodes map[common.Hash][]TrieNode) (map[common.Hash][]StorageNodeModel, error) {
	storageLeafCids := make(map[common.Hash][]StorageNodeModel)
	for addrKey, storageTrie := range storageNodes {
		storageLeafCids[addrKey] = make([]StorageNodeModel, 0, len(storageTrie))
		for _, node := range storageTrie {
			storageNodeCid, err := pub.StoragePutter.DagPut(node.Value)
			if err != nil {
				return nil, err
			}
			if len(storageNodeCid) != 1 {
				return nil, errors.New("single CID expected to be returned for storage leaf")
			}
			// Map storage node cids to their state key hashes
			storageLeafCids[addrKey] = append(storageLeafCids[addrKey], StorageNodeModel{
				StorageKey: node.Key.Hex(),
				CID:        storageNodeCid[0],
				Leaf:       node.Leaf,
			})
		}
	}
	return storageLeafCids, nil
}
