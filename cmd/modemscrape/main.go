package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/mapstructure"

	"github.com/kpcraig/modemscrape"
	"github.com/kpcraig/modemscrape/modem"
	"github.com/kpcraig/modemscrape/output"
)

type config struct {
	Modem   modemInfo              `toml:"modem"`
	Outputs map[string]interface{} `toml:"output"`
}

type modemInfo struct {
	Host     string `toml:"host"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

var (
	test bool

	outputs []modemscrape.Output
)

func init() {
	flag.BoolVar(&test, "test", false, "just parse the toml and quit")
}

func main() {
	flag.Parse()

	cfg := &config{}
	_, err := toml.DecodeFile("modemscrape.toml", cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", cfg)
	if test {

		os.Exit(0)
	}
	mo := modem.CM1200{
		URL:      cfg.Modem.Host,
		Username: cfg.Modem.Username,
		Password: cfg.Modem.Password,
	}

	for k, v := range cfg.Outputs {
		vm := v.(map[string]interface{})
		switch k {
		case "influxdb":
			c := struct {
				Hostname string
				Username string
				Password string
				Database string
			}{}
			mapstructure.Decode(vm, &c)
			idb, err := output.NewInfluxDBOutput(c.Hostname, c.Username, c.Password, c.Database)
			if err != nil {
				panic(err)
			}
			outputs = append(outputs, idb)
		case "stdout":
			s := &output.WriterOutput{
				Out:    os.Stdout,
				LogOut: os.Stdout,
			}
			outputs = append(outputs, s)
		}
	}

	tick := time.NewTicker(15 * time.Second)
	for range tick.C {
		us, ds, us31, ds31, err := mo.GetStats()
		for _, o := range outputs {
			err = o.PutStats(us, ds, us31, ds31)
			if err != nil {
				fmt.Printf("couldn't put stats: %s\n", err.Error())
			}
		}
	}

	fmt.Println("ticker exited, terminating monitor")
}
