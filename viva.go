package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type vivaStation struct {
	ID   int
	Lat  float64
	Lon  float64
	Name string
}

type vivaStationsResponse struct {
	Result struct {
		Stations []vivaStation
	} `json:"GetStationsResult"`
}

type vivaSample struct {
	Calmh               int
	Heading             int
	Msg                 string
	Name                string
	Quality             string
	StationID           int
	Trend               string
	Type                string
	Unit                string
	Updated             string
	Value               string
	WaterLevelOffset    float64
	WaterLevelReference string
}

type vivaSamplesResponse struct {
	Result struct {
		Samples []vivaSample
	} `json:"GetSingleStationResult"`
}

func viva(pats []string) error {
	const stationsURL = "https://services.viva.sjofartsverket.se:8080/output/vivaoutputservice.svc/vivastation/"
	res, err := http.Get(stationsURL)
	if err != nil {
		return err
	}

	var stations vivaStationsResponse
	if err := json.NewDecoder(res.Body).Decode(&stations); err != nil {
		return err
	}

	for _, s := range stations.Result.Stations {
		if match(s.Name, pats) {
			res, err := http.Get(stationsURL + strconv.Itoa(s.ID))
			if err != nil {
				return err
			}

			var samples vivaSamplesResponse
			if err := json.NewDecoder(res.Body).Decode(&samples); err != nil {
				return err
			}

			fmt.Println(s.Name)
			fmt.Println(strings.Repeat("=", len(s.Name)))
			for _, s := range samples.Result.Samples {
				fmt.Printf("%s: %s %s\n", s.Name, s.Value, s.Unit)
			}
			fmt.Println()
		}
	}

	return nil
}

func match(s string, pats []string) bool {
	for _, m := range pats {
		m = strings.ToLower(m)
		if strings.Contains(strings.ToLower(s), m) {
			return true
		}
	}
	return false
}
