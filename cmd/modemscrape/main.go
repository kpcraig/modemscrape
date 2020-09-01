package main

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/malyonsus/modemscrape/modem"
	"github.com/malyonsus/modemscrape/output"
)

type config struct {
	modem modemInfo  `toml:"modem"`
	ifx   influxInfo `toml:"influxdb"`
}

type modemInfo struct {
	host     string
	username string
	password string
}

type influxInfo struct {
	host     string
	username string
	password string
	database string
}

func main() {
	cfg := &config{}
	_, err := toml.DecodeFile("modemscrape.toml", cfg)
	if err != nil {
		panic(err)
	}

	tick := time.NewTicker(15 * time.Second)
	mo := modem.CM1200{
		URL:      cfg.modem.host,
		Username: cfg.modem.username,
		Password: cfg.modem.password,
	}

	idb, err := output.NewInfluxDBOutput(cfg.ifx.host, cfg.ifx.username, cfg.ifx.password, cfg.ifx.database)
	if err != nil {
		panic(err)
	}

	for range tick.C {
		us, ds, us31, ds31, err := mo.GetStats()
		err = idb.PutStats(us, ds, us31, ds31)
		if err != nil {
			fmt.Printf("couldn't put stats: %s\n", err.Error())
		}
		// fmt.Printf("%+v\n%+v\n%+v\n%+v\n%+v\n", us, ds, us31, ds31, err)
	}

	fmt.Println("ticker exited, terminating monitor")
}
