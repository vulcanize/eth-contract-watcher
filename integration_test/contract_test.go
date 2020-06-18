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

package integration

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-header-sync/test_config"

	"github.com/vulcanize/eth-contract-watcher/pkg/fetcher"
	"github.com/vulcanize/eth-contract-watcher/pkg/testing"
)

var _ = Describe("Reading contracts", func() {
	Describe("Fetching Contract data", func() {
		It("returns the correct attribute for a real contract", func() {
			rawRPCClient, err := rpc.Dial(test_config.TestClient.RPCPath)
			Expect(err).NotTo(HaveOccurred())
			ethClient := ethclient.NewClient(rawRPCClient)
			f := fetcher.NewFetcher(ethClient)

			contract := testing.SampleContract()
			var balance = new(big.Int)

			args := make([]interface{}, 1)
			args[0] = common.HexToHash("0xd26114cd6ee289accf82350c8d8487fedb8a0c07")

			err = f.FetchContractData(contract.Abi, "0xd26114cd6ee289accf82350c8d8487fedb8a0c07", "balanceOf", args, &balance, 5167471)
			Expect(err).NotTo(HaveOccurred())
			expected := new(big.Int)
			expected.SetString("10897295492887612977137", 10)
			Expect(balance).To(Equal(expected))
		})
	})
})
