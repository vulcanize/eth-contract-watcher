package main

import (
	"github.com/vulcanize/eth-contract-watcher/cmd"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	cmd.Execute()
}
