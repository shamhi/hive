package main

import (
	"context"
	"flag"
	"fmt"
	rd "hive/pkg/db/redis"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

const (
	AllBasesKey string = "bases:all"
	BaseDataKey string = "bases:data:"
	BaseGeoKey  string = "bases:geo"
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
	flag.IntVar(&cfg.DB, "rdb", cfg.DB, "Redis database number")

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

	rdb, err := rd.New(cfg)
	if err != nil {
		fatal("failed to connect to redis", err)
	}
	defer rdb.Close()

	ctx := context.Background()

	switch *cmd {
	case "up":
		for _, item := range s.Items {
			pipe := rdb.Client.TxPipeline()

			pipe.Del(ctx, BaseDataKey+item.ID)

			pipe.SAdd(ctx, AllBasesKey, item.ID)

			pipe.HSet(ctx, BaseDataKey+item.ID,
				"name", item.Name,
				"address", item.Address,
			)

			pipe.GeoAdd(ctx, BaseGeoKey,
				&redis.GeoLocation{
					Name:      item.ID,
					Longitude: item.Lon,
					Latitude:  item.Lat,
				},
			)

			if _, err := pipe.Exec(ctx); err != nil {
				fatal("failed to save base to redis", err)
			}
		}

		fmt.Printf("UP completed: %d items written\n", len(s.Items))

	case "down":
		for _, item := range s.Items {
			pipe := rdb.Client.TxPipeline()

			pipe.Del(ctx, BaseDataKey+item.ID)
			pipe.ZRem(ctx, BaseGeoKey, item.ID)
			pipe.SRem(ctx, AllBasesKey, item.ID)

			if _, err := pipe.Exec(ctx); err != nil {
				fatal(fmt.Sprintf("failed to remove item %s", item.ID), err)
			}
		}

		fmt.Printf("DOWN completed: %d items removed\n", len(s.Items))
	default:
		fatal("unknown command", *cmd)
	}
}

func fatal(msg string, val any) {
	panic(fmt.Sprintf("%s: %v\n", msg, val))
}
