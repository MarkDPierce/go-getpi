package main

import (
	"log"
	"os"
	"strings"
	"time"

	"go-getpi/config"
	"go-getpi/helpers"
)

const version = "v0.0.1"

func main() {
	log.Printf("running app version: %v", version)
	helpers.InitLogger()

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s --config=config.json", os.Args[0])
	}

	configArg := os.Args[1]
	if !strings.HasPrefix(configArg, "--config=") {
		log.Fatalf("Usage: %s --config=config.json", os.Args[0])
	}

	cfgFile := strings.TrimPrefix(configArg, "--config=")

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if cfg.RunOnce {
		err := helpers.SyncPihole(cfg, cfg.UpdateGravity, cfg.IntervalMinutes, cfg.SecondaryHostsAsStringSlice(), cfg.PrimaryHost.SslSecure)
		if err != nil {
			log.Fatalf("Error during sync: %v", err)
		}
	} else {
		for {
			err := helpers.SyncPihole(cfg, cfg.UpdateGravity, cfg.IntervalMinutes, cfg.SecondaryHostsAsStringSlice(), cfg.PrimaryHost.SslSecure)
			if err != nil {
				log.Printf("Error during sync: %v", err)
			}
			time.Sleep(time.Duration(cfg.IntervalMinutes) * time.Minute)
		}
	}
}
