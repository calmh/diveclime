package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var cli struct {
	HOV  bool     `short:"o" description:"Use havochvatten.se instead of viva.sjofartsverket.se"`
	Viva []string `short:"v" description:"Station(s) to fetch data for" default:"flinten 7,malmö"`
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

	if err := viva(cli.Viva); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
