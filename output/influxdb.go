package output

import (
	"fmt"
	"strconv"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influx "github.com/influxdata/influxdb1-client/v2"

	"github.com/kpcraig/modemscrape"
)

type InfluxDBOutput struct {
	cl influx.Client
	db string
}

func (idb *InfluxDBOutput) PutStats(us []modemscrape.UpstreamChannel, ds []modemscrape.DownstreamChannel, usOFDM []modemscrape.UpstreamOFDMChannel, dsOFDM []modemscrape.DownstreamOFDMChannel) error {
	set, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database: idb.db,
	})
	if err != nil {
		return err
	}

	instant := time.Now()

	for i, v := range us {
		if !v.Locked {
			continue
		}
		tags := map[string]string{
			"channel":           strconv.Itoa(i),
			"bonded_channel_id": strconv.Itoa(v.BondedChannelID),
			"modulation":        v.Modulation,
			"locked":            fmt.Sprintf("%t", v.Locked),
		}

		fields := map[string]interface{}{
			"frequency": v.Frequency,
			"power":     v.Power,
		}

		pt, err := influx.NewPoint("modem_upstream", tags, fields, instant)
		if err != nil {
			fmt.Printf("couldn't add points: %s", err.Error())
			continue
		}
		set.AddPoint(pt)
	}

	for i, v := range ds {
		if !v.Locked {
			continue
		}
		tags := map[string]string{
			"channel":           strconv.Itoa(i),
			"bonded_channel_id": strconv.Itoa(v.BondedChannelID),
			"modulation":        v.Modulation,
			"locked":            fmt.Sprintf("%t", v.Locked),
		}

		fields := map[string]interface{}{
			"frequency":               v.Frequency,
			"power":                   v.Power,
			"snr":                     v.SNR,
			"correctable_codewords":   v.CorrectableWords,
			"uncorrectable_codewords": v.UncorrectableWords,
		}

		pt, err := influx.NewPoint("modem_downstream", tags, fields, instant)
		if err != nil {
			fmt.Printf("couldn't add points: %s", err.Error())
			continue
		}
		set.AddPoint(pt)
	}

	for i, v := range usOFDM {
		if !v.Locked {
			continue
		}
		tags := map[string]string{
			"channel":           strconv.Itoa(i),
			"bonded_channel_id": strconv.Itoa(v.BondedChannelID),
			"locked":            fmt.Sprintf("%t", v.Locked),
		}

		fields := map[string]interface{}{
			"frequency": v.Frequency,
			"power":     v.Power,
		}

		pt, err := influx.NewPoint("modem_upstream_ofdm", tags, fields, instant)
		if err != nil {
			fmt.Printf("couldn't add points: %s", err.Error())
			continue
		}
		set.AddPoint(pt)
	}

	for i, v := range dsOFDM {
		if !v.Locked {
			continue
		}
		tags := map[string]string{
			"channel":           strconv.Itoa(i),
			"bonded_channel_id": strconv.Itoa(v.BondedChannelID),
			"locked":            fmt.Sprintf("%t", v.Locked),
		}

		fields := map[string]interface{}{
			"frequency":               v.Frequency,
			"power":                   v.Power,
			"snr":                     v.SNR,
			"correctable_codewords":   v.CorrectableWords,
			"uncorrectable_codewords": v.UncorrectableWords,
		}

		pt, err := influx.NewPoint("modem_downstream_ofdm", tags, fields, instant)
		if err != nil {
			fmt.Printf("couldn't add points: %s", err.Error())
			continue
		}
		set.AddPoint(pt)
	}

	err = idb.cl.Write(set)

	return err
}

func NewInfluxDBOutput(addr, username, password, database string) (*InfluxDBOutput, error) {
	cl, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't make client: %s", err.Error())
	}

	return &InfluxDBOutput{
		cl: cl,
		db: database,
	}, nil
}
