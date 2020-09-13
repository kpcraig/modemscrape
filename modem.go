package modemscrape

import "time"

type Modem interface {
	GetStats() ([]UpstreamChannel, []DownstreamChannel, []UpstreamOFDMChannel, []DownstreamOFDMChannel, error)
}

// LoggableModem should be implemented for modems that can retrieve logs in a syslog-like format as defined in this
// package.
type LoggableModem interface {
	// GetLogs should return all logs created since the last run of this function for an implementing modem.
	GetLogs() ([]Log, error)
}

type Output interface {
	PutStats([]UpstreamChannel, []DownstreamChannel, []UpstreamOFDMChannel, []DownstreamOFDMChannel) error
}

type LogOutput interface {
	PutLogs([]Log) error
}

type UpstreamChannel struct {
	Locked          bool
	Modulation      string
	BondedChannelID int
	Frequency       int64
	Power           float64
}

type DownstreamChannel struct {
	Locked             bool // if this isn't set, many other values may also not be set
	Modulation         string
	BondedChannelID    int
	Frequency          int64
	Power              float64
	SNR                float64
	CorrectableWords   int64
	UncorrectableWords int64
}

type UpstreamOFDMChannel struct {
	Locked          bool
	BondedChannelID int
	Frequency       int64
	Power           float64
}

type DownstreamOFDMChannel struct {
	Locked             bool
	BondedChannelID    int
	Frequency          int64
	Power              float64
	SNR                float64
	CorrectableWords   int64
	UncorrectableWords int64
}

type Log struct {
	SeverityCode int // should match syslog, which is netgear - 1
	Severity     string
	Timestamp    time.Time
	Message      string
}
