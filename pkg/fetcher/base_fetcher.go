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

package fetcher

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/vulcanize/eth-header-sync/pkg/core"

	"github.com/vulcanize/eth-contract-watcher/pkg/abi"
)

type Fetcher struct {
	ethClient core.EthClient
	timeout   time.Duration
}

func NewFetcher(ethClient core.EthClient, timeout time.Duration) *Fetcher {
	return &Fetcher{
		ethClient: ethClient,
		timeout:   timeout,
	}
}

func (f *Fetcher) FetchEthLogsWithCustomQuery(query ethereum.FilterQuery) ([]types.Log, error) {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	gethLogs, err := f.ethClient.FilterLogs(ctx, query)
	logrus.Debug("GetEthLogsWithCustomQuery called")
	if err != nil {
		return []types.Log{}, err
	}
	return gethLogs, nil
}

func (f *Fetcher) FetchContractData(abiJSON string, address string, method string, methodArgs []interface{}, result interface{}, blockNumber int64) error {
	parsed, err := abi.ParseAbi(abiJSON)
	if err != nil {
		return err
	}
	var input []byte
	if methodArgs != nil {
		input, err = parsed.Pack(method, methodArgs...)
	} else {
		input, err = parsed.Pack(method)
	}
	if err != nil {
		return err
	}
	var bn *big.Int
	if blockNumber > 0 {
		bn = big.NewInt(blockNumber)
	}
	output, err := f.callContract(address, input, bn)
	if err != nil {
		return err
	}
	return parsed.Unpack(result, method, output)
}

func (f *Fetcher) callContract(contractHash string, input []byte, blockNumber *big.Int) ([]byte, error) {
	to := common.HexToAddress(contractHash)
	msg := ethereum.CallMsg{To: &to, Data: input}
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	return f.ethClient.CallContract(ctx, msg, blockNumber)
}
