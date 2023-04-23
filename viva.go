package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

//go:embed tzdata/Stockholm
var stockholmData []byte

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
	Calm                int
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
	if len(pats) == 0 {
		return true
	}
	for _, m := range pats {
		m = strings.ToLower(m)
		if strings.Contains(strings.ToLower(s), m) {
			return true
		}
	}
	return false
}

var metrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "viva",
	Name:      "station_metrics",
}, []string{"station", "name"})

func vivaMetrics(pats []string) error {
	const stationsURL = "https://services.viva.sjofartsverket.se:8080/output/vivaoutputservice.svc/vivastation/"
	res, err := http.Get(stationsURL)
	if err != nil {
		return err
	}

	var stations vivaStationsResponse
	if err := json.NewDecoder(res.Body).Decode(&stations); err != nil {
		return err
	}

	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		loc, err = time.LoadLocationFromTZData("Europe/Stockholm", stockholmData)
	}
	if err != nil {
		loc = time.Local
	}

	for _, station := range stations.Result.Stations {
		if match(station.Name, pats) {
			stationName := sanitizeString(station.Name)
			res, err := http.Get(stationsURL + strconv.Itoa(station.ID))
			if err != nil {
				metrics.DeletePartialMatch(prometheus.Labels{"station": stationName})
				return err
			}

			var samples vivaSamplesResponse
			if err := json.NewDecoder(res.Body).Decode(&samples); err != nil {
				metrics.DeletePartialMatch(prometheus.Labels{"station": stationName})
				return err
			}

			for _, sample := range samples.Result.Samples {
				sampleName := sanitizeString(sample.Name)
				val := strings.TrimLeft(sample.Value, ">NOSV ")
				v, err := strconv.ParseFloat(val, 64)
				if err != nil {
					metrics.Delete(prometheus.Labels{"station": stationName, "name": sampleName})
					continue
				}
				metrics.WithLabelValues(stationName, sampleName).Set(v)

				// See if the value was prefixed with a direction.
				dir := -1
				before, _, _ := strings.Cut(sample.Value, " ")
				switch before {
				case "N":
					dir = 0
				case "NO":
					dir = 45
				case "O":
					dir = 90
				case "SO":
					dir = 135
				case "S":
					dir = 180
				case "SV":
					dir = 225
				case "V":
					dir = 270
				case "NV":
					dir = 315
				}
				if dir != -1 {
					metrics.WithLabelValues(stationName, sampleName+" Riktning").Set(float64(dir))
				}

				if t, err := time.ParseInLocation("2006-01-02 15:04:05", sample.Updated, loc); err == nil {
					metrics.WithLabelValues(stationName, sampleName+" Updated").Set(float64(t.Unix()))
					continue
				}

			}
			metrics.WithLabelValues(stationName, "Updated").Set(float64(time.Now().Unix()))
		}
	}

	return nil
}

func sanitizeString(s string) string {
	// Remove diacritics.
	t := transform.Chain(
		// Split runes with diacritics into base character and mark.
		norm.NFD,
		runes.Remove(runes.Predicate(func(r rune) bool {
			return unicode.Is(unicode.Mn, r) || r > unicode.MaxASCII
		})))
	res, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return res
}
