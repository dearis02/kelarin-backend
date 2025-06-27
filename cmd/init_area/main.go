package main

import (
	"context"
	"encoding/json"
	"fmt"
	"kelarin/internal/config"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"os"
	"strconv"
	"sync"

	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type jsonData map[string]string

func main() {
	cfg := config.NewApp("config/config.yaml")
	config.NewLogger(cfg)

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Error().Stack().Err(errors.New(err)).Send()
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go initProvinces(db, &wg)
	go initCities(db, &wg)

	wg.Wait()
	log.Info().Msg("Area initialization completed successfully")
}

func initProvinces(db *sqlx.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	provinceData, err := os.Open("area/provinces/provinces.json")
	if err != nil {
		log.Error().Err(err).Send()
	}

	defer provinceData.Close()

	decoder := json.NewDecoder(provinceData)

	data := jsonData{}
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	provinces := []types.Province{}
	for key, val := range data {
		id, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			log.Error().Err(err).Send()
		}

		provinces = append(provinces, types.Province{
			ID:   id,
			Name: val,
		})
	}

	query := `INSERT INTO provinces (id, name) VALUES (:id, :name) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;`
	_, err = db.NamedExecContext(context.Background(), query, provinces)
	if err != nil {
		log.Error().Err(errors.New(err)).Send()
	}
}

func initCities(db *sqlx.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 11; i <= 97; i++ {
		fileName := fmt.Sprintf("area/cities/kab-%d.json", i)

		cityData, err := os.Open(fileName)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			log.Error().Err(err).Send()
		}

		decoder := json.NewDecoder(cityData)

		data := jsonData{}
		err = decoder.Decode(&data)
		if err != nil {
			log.Error().Err(err).Send()
		}

		cities := []types.City{}
		for _, val := range data {
			cities = append(cities, types.City{
				ProvinceID: int64(i),
				Name:       val,
			})
		}

		query := `INSERT INTO cities (province_id, name) VALUES (:province_id, :name) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;`
		_, err = db.NamedExecContext(context.Background(), query, cities)
		if err != nil {
			log.Error().Err(errors.New(err)).Send()
		}

		cityData.Close()
	}
}
