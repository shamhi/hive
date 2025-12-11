package main

import (
	"flag"
	"fmt"
	"hive/pkg/geo"
	"os"

	"github.com/google/uuid"
	"github.com/jaswdr/faker/v2"
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
	n := flag.Int("n", 5, "count of items to generate")
	outPath := flag.String("out", "migrations/seed.yaml", "output file path")
	flag.Parse()

	f := faker.New()
	s := SeedFile{Items: make([]SeedItem, 0, *n)}

	for range *n {
		lat, lon := geo.RandMoscowPoint()
		s.Items = append(s.Items, SeedItem{
			ID:      uuid.NewString(),
			Name:    f.Company().Name(),
			Address: f.Address().Address(),
			Lat:     lat,
			Lon:     lon,
		})
	}

	data, err := yaml.Marshal(&s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal seed file: %v\n", err)
	}

	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write seed file: %v\n", err)
	}

	fmt.Println("")
}
