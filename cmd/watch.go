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

package cmd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vulcanize/eth-header-sync/pkg/postgres"

	"github.com/vulcanize/eth-contract-watcher/pkg/config"
	st "github.com/vulcanize/eth-contract-watcher/pkg/transformer"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watches events at the provided contract address using fully synced vDB",
	Long: `Uses input contract address and event filters to watch events

Expects an ethereum node to be running
Expects an archival node synced into vulcanizeDB
Requires a .toml config file:

  [database]
    name     = "vulcanize_public"
    hostname = "localhost"
    port     = 5432

  [client]
    rpcPath  = "/Users/user/Library/Ethereum/geth.ipc"

  [contract]
    network  = ""
    addresses  = [
        "contractAddress1",
        "contractAddress2"
    ]
    [contract.contractAddress1]
        abi    = 'ABI for contract 1'
        startingBlock = 982463
    [contract.contractAddress2]
        abi    = 'ABI for contract 2'
        events = [
            "event1",
            "event2"
        ]
		eventArgs = [
			"arg1",
			"arg2"
		]
        methods = [
            "method1",
			"method2"
        ]
		methodArgs = [
			"arg1",
			"arg2"
		]
        startingBlock = 4448566
        piping = true
`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *log.WithField("SubCommand", subCommand)
		watch()
	},
}

func watch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	client, node := getClientAndNode()

	db, err := postgres.NewDB(databaseConfig, node)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	con := config.ContractConfig{}
	con.PrepConfig()
	transformer := st.NewTransformer(con, client, db, timeout)

	if err := transformer.Init(); err != nil {
		logWithCommand.Fatal(fmt.Sprintf("Failed to initialize transformer, err: %v ", err))
	}

	for range ticker.C {
		err = transformer.Execute()
		if err != nil {
			logWithCommand.Error("Execution error for transformer: ", transformer.GetConfig().Name, err)
		}
	}
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
