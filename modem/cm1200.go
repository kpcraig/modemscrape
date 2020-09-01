package modem

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/malyonsus/modemscrape"
)

type CM1200 struct {
	URL      string
	Username string
	Password string
}

const (
	TagUpstream       = "Us"
	TagDownstream     = "Ds"
	TagUpstreamOFDM   = "UsOfdma"
	TagDownstreamOFDM = "DsOfdm"
)

var (
	// Valid Firmware versions for this implementation.
	// I'm not really sure how I intend this to be used as things need to be rewritten.
	FirmwareVersion = []string{"2.02.05"}
)

// var CM1200VarRegExp = regexp.MustCompile(`tagValueList = '([0-9a-zA-Z|. &;:+~]+)';`)
var CM1200VarRegExp = regexp.MustCompile(`(?s)Init(\w+)TableTagValue.*?tagValueList = '([0-9a-zA-Z|. &;:+~]+)';`)

func (c *CM1200) GetStats() ([]modemscrape.UpstreamChannel, []modemscrape.DownstreamChannel, []modemscrape.UpstreamOFDMChannel, []modemscrape.DownstreamOFDMChannel, error) {
	req, err := http.NewRequest(http.MethodGet, c.URL+"/DocsisStatus.htm", nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// fail in order to get the xsrf cookie
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	xsrf := resp.Header.Get("Set-Cookie")
	resp.Body.Close()

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0")
	req.Header.Set("Cookie", xsrf)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// if resp.StatusCode != http.StatusOK {
	// 	return nil, nil, nil, nil, fmt.Errorf("status not ok: %d", resp.StatusCode)
	// }
	if resp.Body == nil {
		return nil, nil, nil, nil, errors.New("no body")
	}

	defer resp.Body.Close()
	bt, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(bt))
	if err != nil {
		return nil, nil, nil, nil, err
	}

	s := string(bt)

	var up []modemscrape.UpstreamChannel
	var down []modemscrape.DownstreamChannel
	var down31 []modemscrape.DownstreamOFDMChannel
	var up31 []modemscrape.UpstreamOFDMChannel

	matches := CM1200VarRegExp.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		fmt.Println("Case", m[1])
		switch strings.TrimSpace(m[1]) {
		case TagDownstream:
			down, err = parseDownstream(m[2])
			if err != nil {
				fmt.Println(err)
				down = nil
			}
		case TagUpstream:
			up, err = parseUpstream(m[2])
			if err != nil {
				fmt.Println(err)
				up = nil
			}
		case TagDownstreamOFDM:
			down31, err = parseDownstreamOFDM(m[2])
			if err != nil {
				fmt.Println(err)
				down31 = nil
			}
		case TagUpstreamOFDM:
			up31, err = parseUpstreamOFDM(m[2])
			if err != nil {
				fmt.Println(err)
				up31 = nil
			}
		default:
			fmt.Printf("No case for %s\n", strings.TrimSpace(m[1]))
		}
	}

	return up, down, up31, down31, nil
}

func parseUpstream(tagValues string) ([]modemscrape.UpstreamChannel, error) {
	vals := strings.Split(strings.TrimRight(tagValues, "|"), "|")
	rows, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, fmt.Errorf("unexpected non-numeric value at head of tagValues: %s", vals[0])
	}
	data, err := splitRows(rows, vals[1:])
	if err != nil {
		return nil, err
	}

	// '8|1|Locked|ATDMA|1|5120|36500000 Hz|33.3|2|Locked|ATDMA|2|5120|30100000 Hz|34.3|3|Locked|ATDMA|3|5120|23700000 Hz|34.3|4|Locked|ATDMA|4|5120|17300000 Hz|34.8|5|Not Locked|Unknown|0|0|0 Hz|0.0|6|Not Locked|Unknown|0|0|0 Hz|0.0|7|Not Locked|Unknown|0|0|0 Hz|0.0|8|Not Locked|Unknown|0|0|0 Hz|0.0|';

	usc := make([]modemscrape.UpstreamChannel, rows)
	for i := 0; i < rows; i++ {
		// fmt.Println(data[i])
		usc[i].Locked = data[i][1] == "Locked"
		usc[i].Modulation = data[i][2]
		fmt.Sscanf(data[i][3], "%d", &usc[i].BondedChannelID)
		fmt.Sscanf(data[i][5], "%d Hz", &usc[i].Frequency)
		fmt.Sscanf(data[i][6], "%f", &usc[i].Power)
	}

	return usc, nil
}

func parseDownstream(tagValues string) ([]modemscrape.DownstreamChannel, error) {
	vals := strings.Split(strings.TrimRight(tagValues, "|"), "|")
	rows, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, fmt.Errorf("unexpected non-numeric value at head of tagValues: %s", vals[0])
	}
	data, err := splitRows(rows, vals[1:])
	if err != nil {
		return nil, err
	}

	// '1|Locked|QAM256|13|531000000 Hz|5.5|43|15103|23698|2|Locked|QAM256|9|507000000 Hz|5.3|42.8|16368|25363|3|Locked|QAM256|10|513000000 Hz|5.4|42.9|16005|25002|4|Locked|QAM256|11|519000000 Hz|5.3|43|15616|24419|5|Locked|QAM256|12|525000000 Hz|5.3|42.9|15442|23978|6|Locked|QAM256|14|537000000 Hz|5.5|43|15084|23193|7|Locked|QAM256|15|543000000 Hz|5.3|42.8|15220|23317|8|Locked|QAM256|16|549000000 Hz|5.4|42.9|15213|23125|9|Locked|QAM256|17|555000000 Hz|5.3|42.9|14862|22979|10|Locked|QAM256|18|561000000 Hz|5.3|42.8|14922|22439|11|Locked|QAM256|19|567000000 Hz|5.1|42.8|14917|22407|12|Locked|QAM256|20|573000000 Hz|5|42.7|14828|21917|13|Locked|QAM256|21|579000000 Hz|5|42.7|14687|21115|14|Locked|QAM256|22|585000000 Hz|4.9|42.7|14445|20283|15|Locked|QAM256|23|591000000 Hz|4.7|42.4|14451|20162|16|Locked|QAM256|24|597000000 Hz|4.6|42.3|14278|19689|17|Locked|QAM256|25|603000000 Hz|4.4|42.3|14320|19095|18|Locked|QAM256|26|609000000 Hz|4.3|42.4|14035|19114|19|Locked|QAM256|27|615000000 Hz|4.3|42.3|14051|19154|20|Locked|QAM256|28|621000000 Hz|4.3|42.3|14109|19009|21|Locked|QAM256|29|627000000 Hz|4.2|42.2|14127|19039|22|Locked|QAM256|30|633000000 Hz|4.1|42.2|14003|18752|23|Locked|QAM256|31|639000000 Hz|4.4|42.4|13950|18638|24|Locked|QAM256|32|645000000 Hz|4.4|42.3|14104|18560|25|Locked|QAM256|33|651000000 Hz|4.5|42.4|13621|20778|26|Locked|QAM256|34|657000000 Hz|4.5|42.3|13861|20335|27|Locked|QAM256|35|663000000 Hz|4.4|42.2|13400|20327|28|Locked|QAM256|36|669000000 Hz|4.4|42.3|13276|19970|29|Locked|QAM256|37|675000000 Hz|4.3|41|13059|20309|30|Locked|QAM256|38|681000000 Hz|4.1|42.1|12965|18979|31|Locked|QAM256|39|687000000 Hz|4.1|42.1|13349|18580|32|Locked|QAM256|40|693000000 Hz|4|42|13026|18676|';

	dsc := make([]modemscrape.DownstreamChannel, rows)
	for i := 0; i < rows; i++ {
		// fmt.Println(data[i])
		dsc[i].Locked = data[i][1] == "Locked"
		dsc[i].Modulation = data[i][2]
		fmt.Sscanf(data[i][3], "%d", &dsc[i].BondedChannelID)
		fmt.Sscanf(data[i][4], "%d Hz", &dsc[i].Frequency)
		fmt.Sscanf(data[i][5], "%f", &dsc[i].Power)
		fmt.Sscanf(data[i][6], "%f", &dsc[i].SNR)
		fmt.Sscanf(data[i][7], "%d", &dsc[i].CorrectableWords)
		fmt.Sscanf(data[i][8], "%d", &dsc[i].UncorrectableWords)
	}

	return dsc, nil
}

func parseDownstreamOFDM(tagValues string) ([]modemscrape.DownstreamOFDMChannel, error) {
	vals := strings.Split(strings.TrimRight(tagValues, "|"), "|")
	rows, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, fmt.Errorf("unexpected non-numeric value at head of tagValues: %s", vals[0])
	}
	data, err := splitRows(rows, vals[1:])
	if err != nil {
		return nil, err
	}

	dsc := make([]modemscrape.DownstreamOFDMChannel, rows)
	for i := 0; i < rows; i++ {
		// fmt.Println(data[i])
		dsc[i].Locked = data[i][1] == "Locked"
		fmt.Sscanf(data[i][3], "%d", &dsc[i].BondedChannelID)
		fmt.Sscanf(data[i][4], "%d Hz", &dsc[i].Frequency)
		fmt.Sscanf(data[i][5], "%f dBmV", &dsc[i].Power)
		fmt.Sscanf(data[i][6], "%f dB", &dsc[i].SNR)
		fmt.Sscanf(data[i][9], "%d", &dsc[i].CorrectableWords)
		fmt.Sscanf(data[i][10], "%d", &dsc[i].UncorrectableWords)
	}
	return dsc, nil
}

func parseUpstreamOFDM(tagValues string) ([]modemscrape.UpstreamOFDMChannel, error) {
	vals := strings.Split(strings.TrimRight(tagValues, "|"), "|")
	rows, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, fmt.Errorf("unexpected non-numeric value at head of tagValues: %s", vals[0])
	}
	data, err := splitRows(rows, vals[1:])
	if err != nil {
		return nil, err
	}

	usc := make([]modemscrape.UpstreamOFDMChannel, rows)
	for i := 0; i < rows; i++ {
		// fmt.Println(data[i])
		usc[i].Locked = data[i][1] == "Locked"
		if !usc[i].Locked {
			continue
		}
	}

	return usc, nil
}

func splitRows(height int, splitVals []string) ([][]string, error) {
	if len(splitVals)%height != 0 {
		return nil, fmt.Errorf("requested split is not rectangular with height %d and total size %d", height, len(splitVals))
	}

	grid := make([][]string, height)
	stride := len(splitVals) / height
	for i := 0; i < height; i++ {
		grid[i] = splitVals[i*stride : i*stride+stride]
	}

	return grid, nil
}
