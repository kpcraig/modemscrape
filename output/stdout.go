package output

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/kpcraig/modemscrape"
)

// A WriterOutput is an Output and a LogOutput that writes to the supplied io.Writers. They can be the same Writer.
type WriterOutput struct {
	Out    io.Writer
	LogOut io.Writer
}

func NewStdOutWriter() *WriterOutput {
	return &WriterOutput{
		Out:    os.Stdout,
		LogOut: os.Stdout,
	}
}

func (w *WriterOutput) PutStats(us []modemscrape.UpstreamChannel, ds []modemscrape.DownstreamChannel, usOFDM []modemscrape.UpstreamOFDMChannel, dsOFDM []modemscrape.DownstreamOFDMChannel) error {

	var buf bytes.Buffer
	buf.WriteString("Downstream:\n")
	for i, v := range ds {
		buf.WriteString(fmt.Sprintf("channel %d: %+v\n", i+1, v))
	}
	buf.WriteString("Downstream OFDM:\n")
	for i, v := range dsOFDM {
		buf.WriteString(fmt.Sprintf("channel %d: %+v\n", i+1, v))
	}
	buf.WriteString("Upstream:\n")
	for i, v := range us {
		buf.WriteString(fmt.Sprintf("channel %d: %+v\n", i+1, v))
	}
	buf.WriteString("Upstream OFDM:\n")
	for i, v := range usOFDM {
		buf.WriteString(fmt.Sprintf("channel %d: %+v\n", i+1, v))
	}

	n, err := w.Out.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != buf.Len() {
		return errors.New("didn't write enough bytes")
	}

	return nil
}

func (w *WriterOutput) PutLogs(logs []modemscrape.Log) error {
	var buf bytes.Buffer
	for _, l := range logs {
		buf.WriteString(fmt.Sprintf("%s: (%d) %s\n", l.Timestamp, l.SeverityCode, l.Message))
	}

	_, err := w.LogOut.Write(buf.Bytes())

	return err
}
