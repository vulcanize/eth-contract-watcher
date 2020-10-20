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

package transformer

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/vulcanize/eth-header-sync/pkg/core"
	"github.com/vulcanize/eth-header-sync/pkg/postgres"

	"github.com/vulcanize/eth-contract-watcher/pkg/config"
	"github.com/vulcanize/eth-contract-watcher/pkg/contract"
	"github.com/vulcanize/eth-contract-watcher/pkg/converter"
	"github.com/vulcanize/eth-contract-watcher/pkg/fetcher"
	"github.com/vulcanize/eth-contract-watcher/pkg/parser"
	"github.com/vulcanize/eth-contract-watcher/pkg/poller"
	"github.com/vulcanize/eth-contract-watcher/pkg/repository"
	"github.com/vulcanize/eth-contract-watcher/pkg/retriever"
	"github.com/vulcanize/eth-contract-watcher/pkg/types"
)

// Transformer is the top level struct for transforming watched contract data
// Requires a header synced vDB (headers) and a running eth node (or infura)
type Transformer struct {
	// Database interfaces
	EventRepository  repository.EventRepository  // Holds transformed watched event log data
	HeaderRepository repository.HeaderRepository // Interface for interaction with header repositories

	// Pre-processing interfaces
	Parser    parser.Parser            // Parses events and methods out of contract abi fetched using contract address
	Retriever retriever.BlockRetriever // Retrieves first block for contract

	// Processing interfaces
	Fetcher   fetcher.LogFetcher     // Fetches event logs, using header hashes
	Converter converter.LogConverter // Converts watched event logs into custom log
	Poller    poller.Poller          // Polls methods using arguments collected from events and persists them using a method datastore

	// Store contract configuration information
	Config config.ContractConfig

	// Store contract info as mapping to contract address
	Contracts map[string]*contract.Contract

	// Internally configured transformer variables
	contractAddresses []string            // Holds all contract addresses, for batch fetching of logs
	sortedEventIds    map[string][]string // Map to sort event column ids by contract, for post fetch processing and persisting of logs
	sortedMethodIds   map[string][]string // Map to sort method column ids by contract, for post fetch method polling
	eventIds          []string            // Holds event column ids across all contract, for batch fetching of headers
	eventFilters      []common.Hash       // Holds topic0 hashes across all contracts, for batch fetching of logs
	Start             int64               // Hold the lowest starting block and the highest ending block
}

// Order-of-operations:
// 1. Create new transformer
// 2. Load contract addresses and their parameters
// 3. Init
// 4. Execute

// NewTransformer takes in a contract config, fetcher, and database, and returns a new Transformer
func NewTransformer(con config.ContractConfig, client core.EthClient, db *postgres.DB, timeout time.Duration) *Transformer {
	return &Transformer{
		Poller:           poller.NewPoller(client, db, types.HeaderSync, timeout),
		Fetcher:          fetcher.NewFetcher(client, timeout),
		Parser:           parser.NewParser(con.Network),
		HeaderRepository: repository.NewHeaderRepository(db),
		Retriever:        retriever.NewBlockRetriever(db),
		Converter:        &converter.Converter{},
		Contracts:        map[string]*contract.Contract{},
		EventRepository:  repository.NewEventRepository(db, types.HeaderSync),
		Config:           con,
	}
}

// Init initialized the Transformer
// Use after creating and setting transformer
// Loops over all of the addr => filter sets
// Uses parser to pull event info from abi
// Use this info to generate event filters
func (tr *Transformer) Init() error {
	// Initialize internally configured transformer settings
	tr.contractAddresses = make([]string, 0)       // Holds all contract addresses, for batch fetching of logs
	tr.sortedEventIds = make(map[string][]string)  // Map to sort event column ids by contract, for post fetch processing and persisting of logs
	tr.sortedMethodIds = make(map[string][]string) // Map to sort method column ids by contract, for post fetch method polling
	tr.eventIds = make([]string, 0)                // Holds event column ids across all contract, for batch fetching of headers
	tr.eventFilters = make([]common.Hash, 0)       // Holds topic0 hashes across all contracts, for batch fetching of logs
	tr.Start = math.MaxInt64

	// Iterate through all internal contract addresses
	for contractAddr := range tr.Config.Addresses {
		// Configure Abi
		if tr.Config.Abis[contractAddr] == "" {
			// If no abi is given in the config, this method will try fetching from internal look-up table and etherscan
			parseErr := tr.Parser.Parse(contractAddr)
			if parseErr != nil {
				return fmt.Errorf("error parsing contract by address: %s", parseErr.Error())
			}
		} else {
			// If we have an abi from the config, load that into the parser
			parseErr := tr.Parser.ParseAbiStr(tr.Config.Abis[contractAddr])
			if parseErr != nil {
				return fmt.Errorf("error parsing contract abi: %s", parseErr.Error())
			}
		}

		// Get first block and most recent block number in the header repo
		firstBlock, retrieveErr := tr.Retriever.RetrieveFirstBlock()
		if retrieveErr != nil {
			if retrieveErr == sql.ErrNoRows {
				logrus.Error(fmt.Errorf("error retrieving first block: %s", retrieveErr.Error()))
				firstBlock = 0
			} else {
				return fmt.Errorf("error retrieving first block: %s", retrieveErr.Error())
			}
		}

		// Set to specified range if it falls within the bounds
		if firstBlock < tr.Config.StartingBlocks[contractAddr] {
			firstBlock = tr.Config.StartingBlocks[contractAddr]
		}

		// Get contract name if it has one
		var name = new(string)
		pollingErr := tr.Poller.FetchContractData(tr.Parser.Abi(), contractAddr, "name", nil, name, -1)
		if pollingErr != nil {
			// can't return this error because "name" might not exist on the contract
			logrus.Warnf("error fetching contract data: %s", pollingErr.Error())
		}

		// Remove any potential accidental duplicate inputs
		eventArgs := map[string]bool{}
		for _, arg := range tr.Config.EventArgs[contractAddr] {
			eventArgs[arg] = true
		}
		methodArgs := map[string]bool{}
		for _, arg := range tr.Config.MethodArgs[contractAddr] {
			methodArgs[arg] = true
		}

		// Aggregate info into contract object and store for execution
		con := contract.Contract{
			Name:          *name,
			Network:       tr.Config.Network,
			Address:       contractAddr,
			Abi:           tr.Parser.Abi(),
			ParsedAbi:     tr.Parser.ParsedAbi(),
			StartingBlock: firstBlock,
			Events:        tr.Parser.GetEvents(tr.Config.Events[contractAddr]),
			Methods:       tr.Parser.GetSelectMethods(tr.Config.Methods[contractAddr]),
			FilterArgs:    eventArgs,
			MethodArgs:    methodArgs,
			Piping:        tr.Config.Piping[contractAddr],
		}.Init()
		tr.Contracts[contractAddr] = con
		tr.contractAddresses = append(tr.contractAddresses, con.Address)

		// Create checked_headers columns for each event id and append to list of all event ids
		tr.sortedEventIds[con.Address] = make([]string, 0, len(con.Events))
		for _, event := range con.Events {
			eventID := strings.ToLower(event.Name + "_" + con.Address)
			addColumnErr := tr.HeaderRepository.AddCheckColumn(eventID)
			if addColumnErr != nil {
				return fmt.Errorf("error adding check column: %s", addColumnErr.Error())
			}
			// Keep track of this event id; sorted and unsorted
			tr.sortedEventIds[con.Address] = append(tr.sortedEventIds[con.Address], eventID)
			tr.eventIds = append(tr.eventIds, eventID)
			// Append this event sig to the filters
			tr.eventFilters = append(tr.eventFilters, event.Sig())
		}

		// Create checked_headers columns for each method id and append list of all method ids
		tr.sortedMethodIds[con.Address] = make([]string, 0, len(con.Methods))
		for _, m := range con.Methods {
			methodID := strings.ToLower(m.Name + "_" + con.Address)
			addColumnErr := tr.HeaderRepository.AddCheckColumn(methodID)
			if addColumnErr != nil {
				return fmt.Errorf("error adding check column: %s", addColumnErr.Error())
			}
			tr.sortedMethodIds[con.Address] = append(tr.sortedMethodIds[con.Address], methodID)
		}

		// Update start to the lowest block
		if con.StartingBlock < tr.Start {
			tr.Start = con.StartingBlock
		}
	}

	return nil
}

// Execute runs the transformation processes
func (tr *Transformer) Execute() error {
	if len(tr.Contracts) == 0 {
		return errors.New("error: transformer has no initialized contracts")
	}

	// Find unchecked headers for all events across all contracts; these are returned in asc order
	missingHeaders, missingHeadersErr := tr.HeaderRepository.MissingHeadersForAll(tr.Start, -1, tr.eventIds)
	if missingHeadersErr != nil {
		return fmt.Errorf("error getting missing headers: %s", missingHeadersErr.Error())
	}

	// Iterate over headers
	for _, header := range missingHeaders {
		// Set `start` to this header
		// This way if we throw an error but don't bring the execution cycle down (how it is currently handled)
		// we restart the cycle at this header
		tr.Start = header.BlockNumber
		// Map to sort batch fetched logs by which contract they belong to, for post fetch processing
		sortedLogs := make(map[string][]gethTypes.Log)
		// And fetch all event logs across contracts at this header
		allLogs, fetchErr := tr.Fetcher.FetchLogs(tr.contractAddresses, tr.eventFilters, header)
		if fetchErr != nil {
			return fmt.Errorf("error fetching logs: %s", fetchErr.Error())
		}

		// If no logs are found mark the header checked for all of these eventIDs
		// and continue to method polling and onto the next iteration
		if len(allLogs) < 1 {
			markCheckedErr := tr.HeaderRepository.MarkHeaderCheckedForAll(header.ID, tr.eventIds)
			if markCheckedErr != nil {
				return fmt.Errorf("error marking header checked: %s", markCheckedErr.Error())
			}
			pollingErr := tr.methodPolling(header, tr.sortedMethodIds)
			if pollingErr != nil {
				return fmt.Errorf("error polling methods: %s", pollingErr.Error())
			}
			tr.Start = header.BlockNumber + 1 // Empty header; setup to start at the next header
			logrus.Tracef("no logs found for block %d, continuing", header.BlockNumber)
			continue
		}

		for _, log := range allLogs {
			addr := strings.ToLower(log.Address.Hex())
			sortedLogs[addr] = append(sortedLogs[addr], log)
		}

		// Process logs for each contract
		for conAddr, logs := range sortedLogs {
			if logs == nil {
				logrus.Tracef("no logs found for contract %s at block %d, continuing", conAddr, header.BlockNumber)
				continue
			}
			// Configure converter with this contract
			con := tr.Contracts[conAddr]
			tr.Converter.Update(con)

			// Convert logs into batches of log mappings (eventName => []types.Logs
			convertedLogs, convertErr := tr.Converter.ConvertBatch(logs, con.Events, header.ID)
			if convertErr != nil {
				return fmt.Errorf("error converting logs: %s", convertErr.Error())
			}
			// Cycle through each type of event log and persist them
			for eventName, logs := range convertedLogs {
				// If logs for this event are empty, mark them checked at this header and continue
				if len(logs) < 1 {
					logrus.Tracef("no logs found for event %s on contract %s at block %d, continuing", eventName, conAddr, header.BlockNumber)
					continue
				}
				// If logs aren't empty, persist them
				persistErr := tr.EventRepository.PersistLogs(logs, con.Events[eventName], con.Address, con.Name)
				if persistErr != nil {
					return fmt.Errorf("error persisting logs: %s", persistErr.Error())
				}
			}
		}

		markCheckedErr := tr.HeaderRepository.MarkHeaderCheckedForAll(header.ID, tr.eventIds)
		if markCheckedErr != nil {
			return fmt.Errorf("error marking header checked: %s", markCheckedErr.Error())
		}

		// Poll contracts at this block height
		pollingErr := tr.methodPolling(header, tr.sortedMethodIds)
		if pollingErr != nil {
			return fmt.Errorf("error polling methods: %s", pollingErr.Error())
		}
		// Success; setup to start at the next header
		tr.Start = header.BlockNumber + 1
	}

	return nil
}

// Used to poll contract methods at a given header
func (tr *Transformer) methodPolling(header core.Header, sortedMethodIds map[string][]string) error {
	for _, con := range tr.Contracts {
		// Skip method polling processes if no methods are specified
		// Also don't try to poll methods below this contract's specified starting block
		if len(con.Methods) == 0 || header.BlockNumber < con.StartingBlock {
			logrus.Tracef("not polling contract: %s", con.Address)
			continue
		}

		// Poll all methods for this contract at this header
		pollingErr := tr.Poller.PollContractAt(*con, header.BlockNumber)
		if pollingErr != nil {
			return fmt.Errorf("error polling contract %s: %s", con.Address, pollingErr.Error())
		}

		// Mark this header checked for the methods
		markCheckedErr := tr.HeaderRepository.MarkHeaderCheckedForAll(header.ID, sortedMethodIds[con.Address])
		if markCheckedErr != nil {
			return fmt.Errorf("error marking header checked: %s", markCheckedErr.Error())
		}
	}

	return nil
}

// GetConfig returns the transformers config; satisfies the transformer interface
func (tr *Transformer) GetConfig() config.ContractConfig {
	return tr.Config
}
