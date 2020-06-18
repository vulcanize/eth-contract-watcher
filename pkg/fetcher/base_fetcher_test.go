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

package fetcher_test

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-header-sync/pkg/fakes"

	f "github.com/vulcanize/eth-contract-watcher/pkg/fetcher"
)

var _ = Describe("Geth fetcher", func() {
	var (
		mockClient *fakes.MockEthClient
		fetcher    *f.Fetcher
	)

	BeforeEach(func() {
		mockClient = fakes.NewMockEthClient()
		fetcher = f.NewFetcher(mockClient)
	})

	Describe("fetching logs with a custom FilterQuery", func() {
		It("fetches logs from ethClient", func() {
			mockClient.SetFilterLogsReturnLogs([]types.Log{{}})
			address := common.HexToAddress("0x")
			startingBlockNumber := big.NewInt(1)
			endingBlockNumber := big.NewInt(2)
			topic := common.HexToHash("0x")
			query := ethereum.FilterQuery{
				FromBlock: startingBlockNumber,
				ToBlock:   endingBlockNumber,
				Addresses: []common.Address{address},
				Topics:    [][]common.Hash{{topic}},
			}

			_, err := fetcher.FetchEthLogsWithCustomQuery(query)

			Expect(err).NotTo(HaveOccurred())
			mockClient.AssertFilterLogsCalledWith(context.Background(), query)
		})

		It("returns err if ethClient returns err", func() {
			mockClient.SetFilterLogsErr(fakes.FakeError)
			startingBlockNumber := big.NewInt(1)
			endingBlockNumber := big.NewInt(2)
			query := ethereum.FilterQuery{
				FromBlock: startingBlockNumber,
				ToBlock:   endingBlockNumber,
				Addresses: []common.Address{common.HexToAddress(common.BytesToHash([]byte{1, 2, 3, 4, 5}).Hex())},
				Topics:    nil,
			}

			_, err := fetcher.FetchEthLogsWithCustomQuery(query)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakes.FakeError))
		})
	})
})
