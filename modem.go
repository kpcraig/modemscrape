package modemscrape

type Modem interface {
	GetStats() ([]UpstreamChannel, []DownstreamChannel, []UpstreamOFDMChannel, []DownstreamOFDMChannel, error)
}

type Output interface {
	PutStats([]UpstreamChannel, []DownstreamChannel, []UpstreamOFDMChannel, []DownstreamOFDMChannel) error
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
