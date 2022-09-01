package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/gocolly/colly"
)

func havOchVatten() error {
	var match *regexp.Regexp
	var err error
	if m := flag.Arg(0); m != "" {
		match, err = regexp.Compile(m)
		if err != nil {
			return fmt.Errorf("invalid regexp: %w", err)
		}
	}

	c := colly.NewCollector()

	temps := make(map[string]float64)
	var key string

	c.OnHTML("td[headers]", func(e *colly.HTMLElement) {
		switch e.Attr("headers") {
		case "head11":
			key = strings.ReplaceAll(e.Text, "\u00ad", "")
			if idx := strings.IndexAny(key, ","); idx > 0 {
				key = key[:idx]
			}
		case "head13":
			fields := strings.Fields(e.Text)
			temps[key], _ = strconv.ParseFloat(fields[0], 64)
		}
	})
	if err := c.Visit("https://www.havochvatten.se/badplatser-och-badvatten/vattenprov-badtemperatur/vattentemperatur-och-kvalitet-pa-badvatten-pa-sydkusten.html"); err != nil {
		return fmt.Errorf("visit: %w", err)
	}

	keys := make([]string, 0, len(temps))
	for key := range temps {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for _, key := range keys {
		arrow := ""
		if match != nil && match.MatchString(key) {
			arrow = "<--"
		}
		fmt.Fprintf(tw, "%s\t%4.1f\t%s\n", key, temps[key], arrow)
	}
	tw.Flush()
	return nil
}
