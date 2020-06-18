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

package fakes

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-header-sync/pkg/core"
)

type MockFetcher struct {
	fetchContractDataErr               error
	fetchContractDataPassedAbi         string
	fetchContractDataPassedAddress     string
	fetchContractDataPassedMethod      string
	fetchContractDataPassedMethodArgs  []interface{}
	fetchContractDataPassedResult      interface{}
	fetchContractDataPassedBlockNumber int64
	logQuery                           ethereum.FilterQuery
	logQueryErr                        error
	logQueryReturnLogs                 []types.Log
	node                               core.Node
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{}
}

func (fethcer *MockFetcher) SetFetchContractDataErr(err error) {
	fethcer.fetchContractDataErr = err
}

func (fethcer *MockFetcher) SetGetEthLogsWithCustomQueryErr(err error) {
	fethcer.logQueryErr = err
}

func (fethcer *MockFetcher) SetGetEthLogsWithCustomQueryReturnLogs(logs []types.Log) {
	fethcer.logQueryReturnLogs = logs
}

func (fethcer *MockFetcher) FetchContractData(abiJSON string, address string, method string, methodArgs []interface{}, result interface{}, blockNumber int64) error {
	fethcer.fetchContractDataPassedAbi = abiJSON
	fethcer.fetchContractDataPassedAddress = address
	fethcer.fetchContractDataPassedMethod = method
	fethcer.fetchContractDataPassedMethodArgs = methodArgs
	fethcer.fetchContractDataPassedResult = result
	fethcer.fetchContractDataPassedBlockNumber = blockNumber
	return fethcer.fetchContractDataErr
}

func (fethcer *MockFetcher) GetEthLogsWithCustomQuery(query ethereum.FilterQuery) ([]types.Log, error) {
	fethcer.logQuery = query
	return fethcer.logQueryReturnLogs, fethcer.logQueryErr
}

func (fethcer *MockFetcher) CallContract(contractHash string, input []byte, blockNumber *big.Int) ([]byte, error) {
	return []byte{}, nil
}

func (fethcer *MockFetcher) AssertFetchContractDataCalledWith(abiJSON string, address string, method string, methodArgs []interface{}, result interface{}, blockNumber int64) {
	Expect(fethcer.fetchContractDataPassedAbi).To(Equal(abiJSON))
	Expect(fethcer.fetchContractDataPassedAddress).To(Equal(address))
	Expect(fethcer.fetchContractDataPassedMethod).To(Equal(method))
	if methodArgs != nil {
		Expect(fethcer.fetchContractDataPassedMethodArgs).To(Equal(methodArgs))
	}
	Expect(fethcer.fetchContractDataPassedResult).To(BeAssignableToTypeOf(result))
	Expect(fethcer.fetchContractDataPassedBlockNumber).To(Equal(blockNumber))
}

func (fethcer *MockFetcher) AssertGetEthLogsWithCustomQueryCalledWith(query ethereum.FilterQuery) {
	Expect(fethcer.logQuery).To(Equal(query))
}

func (fetcher *MockFetcher) Node() core.Node {
	return fetcher.node
}
