package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cli struct {
	HOV                    bool          `short:"o" description:"Use havochvatten.se instead of viva.sjofartsverket.se"`
	Viva                   []string      `short:"v" description:"Station(s) to fetch data for" default:"flinten 7,malmö"`
	PrometheusListen       string        `short:"l" description:"Listen address for Prometheus metrics"`
	PrometheusPollInterval time.Duration `short:"i" description:"Poll interval in seconds" default:"60s"`
}

func main() {
	kong.Parse(&cli)

	if cli.HOV {
		if err := havOchVatten(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	if cli.PrometheusListen != "" {
		go func() {
			if err := vivaMetrics(cli.Viva); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			time.Sleep(cli.PrometheusPollInterval)
		}()
		if err := http.ListenAndServe(cli.PrometheusListen, promhttp.Handler()); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	if err := viva(cli.Viva); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
