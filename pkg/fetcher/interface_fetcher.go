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
	"fmt"

	"github.com/vulcanize/eth-contract-watcher/pkg/constants"
)

// InterfaceFetcher is used to derive the interface of a contract
type InterfaceFetcher interface {
	FetchABI(resolverAddr string, blockNumber int64) (string, error)
}

// FetchABI is used to construct a custom ABI based on the results from calling supportsInterface
func (f *Fetcher) FetchABI(resolverAddr string, blockNumber int64) (string, error) {
	a := constants.SupportsInterfaceABI
	args := make([]interface{}, 1)
	args[0] = constants.MetaSig.Bytes()
	supports, err := f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err != nil {
		return "", fmt.Errorf("call to getSupportsInterface failed: %v", err)
	}
	if !supports {
		return "", fmt.Errorf("contract does not support interface")
	}

	abiStr := `[`
	args[0] = constants.AddrChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.AddrChangeInterface + ","
	}
	args[0] = constants.NameChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.NameChangeInterface + ","
	}
	args[0] = constants.ContentChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.ContentChangeInterface + ","
	}
	args[0] = constants.AbiChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.AbiChangeInterface + ","
	}
	args[0] = constants.PubkeyChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.PubkeyChangeInterface + ","
	}
	args[0] = constants.ContentHashChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.ContenthashChangeInterface + ","
	}
	args[0] = constants.MultihashChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.MultihashChangeInterface + ","
	}
	args[0] = constants.TextChangeSig.Bytes()
	supports, err = f.getSupportsInterface(a, resolverAddr, blockNumber, args)
	if err == nil && supports {
		abiStr += constants.TextChangeInterface + ","
	}
	abiStr = abiStr[:len(abiStr)-1] + `]`

	return abiStr, nil
}

// Use this method to check whether or not a contract supports a given method/event interface
func (f *Fetcher) getSupportsInterface(contractAbi, contractAddress string, blockNumber int64, methodArgs []interface{}) (bool, error) {
	return f.FetchBool("supportsInterface", contractAbi, contractAddress, blockNumber, methodArgs)
}
