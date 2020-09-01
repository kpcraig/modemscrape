package modem

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const URL = "http://192.168.100.1"
const Login = "GenieLogin.asp"
const Stats = "DocsisStatus.asp"

var (
	freqRegExp = regexp.MustCompile(`(\d+) Hz`)
	pwrRegExp  = regexp.MustCompile(`(\d+\.\d+) dBmV`)
	snrRegExp  = regexp.MustCompile(`(\d+\.\d+) dB`)
)

type CM1000 struct {
	URL      string
	Username string
	Password string
}

func (c *CM1000) GetStats() {

	stats, err := c.getStats()

	if err != nil && err == ErrMustLogin {
		fmt.Println("logging in...")
		// log in flow
		tk, err := c.GetWebToken()
		fmt.Printf("Token is: %s\n", tk)
		if err != nil {
			panic(err)
		}

		http.PostForm(URL+"/goform/GenieLogin",
			url.Values{
				"loginUsername": {c.Username},
				"loginPassword": {c.Password},
				"webToken":      {tk},
			},
		)

		stats, err = c.getStats()
	} else if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", stats)

}

var (
	ErrMustLogin = errors.New("must log in to see stats")
)

func (c *CM1000) GetWebToken() (string, error) {
	req, err := http.NewRequest(http.MethodGet, URL+"/"+Login, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non 200 status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	sel := doc.Find("input[name='webToken']")
	if sel.Length() != 1 {
		return "", errors.New("not found")
	}

	return sel.Eq(0).AttrOr("value", ""), nil
}

func (c *CM1000) getStats() (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, URL+"/"+Stats, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, err
	}

	// if we hit a login page, it will contain the text `onload="redirectPage();"`
	doc, err := goquery.NewDocumentFromResponse(resp)
	sel := doc.Find("body")
	if sel.Eq(0).AttrOr("onload", "empty") == "redirectPage();" {
		return nil, ErrMustLogin
	}
	// bt, _ := ioutil.ReadAll(resp.Body)
	// if strings.Contains(string(bt), `onload="redirectPage();"`) {
	// 	return nil, ErrMustLogin
	// }

	// otherwise parse as correct
	data := make(map[string]interface{})
	ds := doc.Find("#dsTable tr")
	ds.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return // skip header row
		}

		var num int
		var dec float64

		// channel, locked, modulation, channel_id, frequency, power, snr_mer, unerrored codewords, correctable codewords, uncorrectable codewords
		cs := s.Children()
		if cs.Eq(1).Text() != "Locked" {
			return
		}
		prefix := fmt.Sprintf("channel_%d_", i)

		id, err := strconv.Atoi(cs.Eq(3).Text())
		if err == nil {
			data[prefix+"channel_id"] = id
		} else {
			fmt.Printf("skipping channel id due to parse error: %s\n", err.Error())
		}

		rawFrequency := cs.Eq(4).Text() // nnnnnnn Hz
		_, err = fmt.Sscanf(rawFrequency, "%d Hz", &num)
		if err == nil {
			data[prefix+"frequency"] = num
		} else {
			fmt.Printf("skipping freq due to parse error: %s\n", err.Error())
		}

		rawPower := cs.Eq(5).Text() // nn.nn dBmV
		_, err = fmt.Sscanf(rawPower, "%f dBmV", &dec)
		if err == nil {
			data[prefix+"power"] = dec
		} else {
			fmt.Printf("skipping power due to parse error: %s\n", err.Error())
		}

		rawSNR := cs.Eq(6).Text() // nn.nn dB
		_, err = fmt.Sscanf(rawSNR, "%f dBmV", &dec)
		if err == nil {
			data[prefix+"snr"] = dec
		} else {
			fmt.Printf("skipping snr due to parse error: %s\n", err.Error())
		}
	})

	us := doc.Find("#usTable tr") // dsTable //d31dsTable //d31usTable
	us.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		// channel, locked, modulation, channel_id, frequency, power
	})

	ds31 := doc.Find("#d31dsTable tr") // dsTable //d31dsTable //d31usTable
	ds31.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		// channel, locked, modulation (a number), channel id, frequency, power, snr_mer, active subcarrier number range, unerrored codewords, correctable codewords, uncorrectable codewords
	})

	us31 := doc.Find("#d31usTable tr")
	us31.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		// channel, locked, modulation, channel id, frequency, power
	})

	return data, nil
}
