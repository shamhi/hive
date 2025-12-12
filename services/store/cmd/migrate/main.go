package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	rd "hive/pkg/db/redis"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

type SeedFile struct {
	Items []SeedItem `yaml:"items"`
}

type SeedItem struct {
	ID      string  `yaml:"id"`
	Name    string  `yaml:"name"`
	Address string  `yaml:"address"`
	Lat     float64 `yaml:"lat"`
	Lon     float64 `yaml:"lon"`
}

func main() {
	var cfg rd.Config

	cmd := flag.String("command", "up", "migration command: up | down")
	inPath := flag.String("in", "migrations/seed.yaml", "input file path")

	flag.StringVar(&cfg.Host, "host", cfg.Host, "Redis host")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Redis port")
	flag.StringVar(&cfg.Password, "password", cfg.Password, "Redis password")
	flag.IntVar(&cfg.DB, "db", cfg.DB, "Redis datastore number")

	flag.Parse()

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		fatal("failed to parse env config", err)
	}

	data, err := os.ReadFile(*inPath)
	if err != nil {
		fatal("failed to read seed file", err)
	}

	var s SeedFile
	if err := yaml.Unmarshal(data, &s); err != nil {
		fatal("failed to unmarshal seed file", err)
	}

	db, err := rd.New(cfg)
	if err != nil {
		fatal("failed to connect to redis", err)
	}
	defer db.Close()

	ctx := context.Background()

	switch *cmd {
	case "up":
		for _, item := range s.Items {
			if err := db.Client.GeoAdd(
				ctx,
				"stores:geo",
				&redis.GeoLocation{
					Name:      item.ID,
					Longitude: item.Lon,
					Latitude:  item.Lat,
				},
			).Err(); err != nil {
				fatal(fmt.Sprintf("failed to save item %s to redis geo", item.ID), err)
			}

			itemJSON, err := json.Marshal(item)
			if err != nil {
				fatal(fmt.Sprintf("failed to marshal item %s to json", item.ID), err)
			}

			if err := db.Client.Set(
				ctx,
				fmt.Sprintf("stores:data:%s", item.ID),
				itemJSON,
				0,
			).Err(); err != nil {
				fatal(fmt.Sprintf("failed to save item %s data to redis", item.ID), err)
			}

			if err := db.Client.SAdd(
				ctx,
				"stores:all",
				item.ID,
			).Err(); err != nil {
				fatal(fmt.Sprintf("failed to save item %s to redis set", item.ID), err)
			}
		}

		fmt.Printf("UP completed: %d items written\n", len(s.Items))

	case "down":
		for _, item := range s.Items {
			if err := db.Client.ZRem(
				ctx,
				"stores:geo",
				item.ID,
			).Err(); err != nil {
				fatal(fmt.Sprintf("failed to remove item %s from redis geo", item.ID), err)
			}

			if err := db.Client.Del(
				ctx,
				fmt.Sprintf("stores:data:%s", item.ID),
			).Err(); err != nil {
				fatal(fmt.Sprintf("failed to delete item %s data from redis", item.ID), err)
			}

			if err := db.Client.SRem(
				ctx,
				"stores:all",
				item.ID,
			).Err(); err != nil {
				fatal(fmt.Sprintf("failed to remove item %s from redis set", item.ID), err)
			}
		}

		fmt.Printf("DOWN completed: %d items removed\n", len(s.Items))
	default:
		fatal("unknown command", *cmd)
	}
}

func fatal(msg string, val any) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, val)
	os.Exit(1)
}
