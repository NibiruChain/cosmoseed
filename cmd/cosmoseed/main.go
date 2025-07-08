package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/NibiruChain/cosmoseed/internal/cosmoseed"
)

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s\nCommit hash: %s\n", cosmoseed.Version, cosmoseed.CommitHash)
		os.Exit(0)
	}

	cfgPath := path.Join(home, configFileName)

	cfg, err := cosmoseed.ReadConfigFromFile(cfgPath)
	if err != nil {
		panic(err)
	}

	if cfg == nil {
		cfg, err = cosmoseed.DefaultConfig()
		if err != nil {
			panic(err)
		}
	}

	if err = cfg.Save(cfgPath); err != nil {
		panic(err)
	}

	if chainID != "" {
		cfg.ChainID = chainID
	}

	if seeds != "" {
		cfg.Seeds = seeds
	}

	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	seeder, err := cosmoseed.NewSeeder(home, cfg)
	if err != nil {
		panic(err)
	}

	if showNodeID {
		fmt.Println(seeder.GetNodeID())
		os.Exit(0)
	}

	if err = seeder.Start(); err != nil {
		panic(err)
	}
}
