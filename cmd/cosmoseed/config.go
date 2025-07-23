package main

import (
	"flag"
	"os"
	"path"

	"github.com/NibiruChain/cosmoseed/internal/utils"
)

const (
	defaultConfigDir = ".cosmoseed"
	configFileName   = "config.yaml"
)

var (
	home, chainID, seeds, logLevel string
	showVersion, showNodeID        bool
)

func init() {
	userHome, _ := os.UserHomeDir()
	defaultHome := path.Join(userHome, defaultConfigDir)

	flag.StringVar(&home,
		"home",
		utils.GetString("HOME_DIR", defaultHome),
		"path to home",
	)
	flag.StringVar(&chainID,
		"chain-id",
		utils.GetString("CHAIN_ID", ""),
		"chain ID to use",
	)
	flag.StringVar(&seeds,
		"seeds",
		utils.GetString("SEEDS", ""),
		"seeds to use",
	)
	flag.StringVar(&logLevel,
		"log-level",
		utils.GetString("LOG_LEVEL", "info"),
		"logging level",
	)

	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&showNodeID, "show-node-id", false, "print node ID and exit")
}
