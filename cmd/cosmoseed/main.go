package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	cosmoseed2 "github.com/NibiruChain/cosmoseed/pkg/cosmoseed"
)

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s\nCommit hash: %s\n", cosmoseed2.Version, cosmoseed2.CommitHash)
		os.Exit(0)
	}

	cfgPath := path.Join(home, configFileName)

	cfg, err := cosmoseed2.ReadConfigFromFile(cfgPath)
	if err != nil {
		panic(err)
	}

	if cfg == nil {
		cfg, err = cosmoseed2.DefaultConfig()
		if err != nil {
			panic(err)
		}
	}

	if !configReadOnly {
		if err = cfg.Save(cfgPath); err != nil {
			panic(err)
		}
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

	if nodeKeyFile != "" {
		cfg.NodeKeyFile = nodeKeyFile
	}

	if externalAddress != "" {
		cfg.ExternalAddress = externalAddress
	}

	seeder, err := cosmoseed2.NewSeeder(home, cfg)
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
