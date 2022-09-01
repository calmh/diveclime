package main

import "github.com/alecthomas/kong"

var cli struct {
	HOV  bool   `short:"h" long:"havochvatten" description:"Use havochvatten.se instead of viva.sjofartsverket.se"`
	ViVa string `short:"v" long:"viva" description:"Station(s) to fetch data for" default:"flinten"`
}

func main() {
	kong.Parse(&cli)

	if cli.HOV {
		havOchVatten()
	} else {
		viva(cli.ViVa)
	}
}
